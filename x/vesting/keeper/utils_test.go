package keeper_test

import (
	"encoding/json"
	"math/big"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v11/app"
	cosmosante "github.com/evmos/evmos/v11/app/ante/cosmos"
	evmante "github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/contracts"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/server/config"
	"github.com/evmos/evmos/v11/tests"
	"github.com/evmos/evmos/v11/testutil"
	evmostypes "github.com/evmos/evmos/v11/types"
	"github.com/evmos/evmos/v11/utils"
	epochstypes "github.com/evmos/evmos/v11/x/epochs/types"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
	"github.com/evmos/evmos/v11/x/vesting/types"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// Init app
	suite.app = app.Setup(checkTx, nil)

	// Set Context
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	// Setup query helpers
	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelperEvm)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.VestingKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	// Set epoch start time and height for all epoch identifiers from the epoch
	// module
	identifiers := []string{epochstypes.WeekEpochID, epochstypes.DayEpochID}
	for _, identifier := range identifiers {
		epoch, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, identifier)
		suite.Require().True(found)
		epoch.StartTime = suite.ctx.BlockTime()
		epoch.CurrentEpochStartHeight = suite.ctx.BlockHeight()
		suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epoch)
	}

	acc := &evmostypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
	suite.validator = validators[0]

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	// Deploy contracts
	contract, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	require.NoError(t, err)
	contract2, err = suite.DeployContract(erc20Name2, erc20Symbol2, erc20Decimals)
	require.NoError(t, err)
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	_ = suite.app.Commit()
	header := suite.ctx.BlockHeader()
	header.Height++
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelper)
}

// MintFeeCollector mints coins with the bank modules and sends them to the fee
// collector.
func (suite *KeeperTestSuite) MintFeeCollector(coins sdk.Coins) {
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

// DeployContract deploys the ERC20MinterBurnerDecimalsContract.
func (suite *KeeperTestSuite) DeployContract(
	name, symbol string,
	decimals uint8,
) (common.Address, error) {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", name, symbol, decimals)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	if err != nil {
		return common.Address{}, err
	}

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	if err != nil {
		return common.Address{}, err
	}

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evmtypes.NewTxContract(
		chainID,
		nonce,
		nil,     // amount
		res.Gas, // gasLimit
		nil,     // gasPrice
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(1),
		data,                   // input
		&ethtypes.AccessList{}, // accesses
	)

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	if err != nil {
		return common.Address{}, err
	}

	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	if err != nil {
		return common.Address{}, err
	}

	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce), nil
}

// assertEthFails is a helper function that takes in 1 or more messages and checks
// that they can neither be validated nor delivered using the EthVesting.
func assertEthFails(msgs ...sdk.Msg) {
	insufficientUnlocked := "insufficient unlocked"

	err := validateEthVestingTransactionDecorator(msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficientUnlocked))

	// Sanity check that delivery fails as well
	_, err = testutil.DeliverEthTx(s.ctx, s.app, nil, msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficientUnlocked))
}

// assertEthSucceeds is a helper function, that checks if 1 or more messages
// can be validated and delivered.
func assertEthSucceeds(testAccounts []TestClawbackAccount, funder sdk.AccAddress, dest sdk.AccAddress, amount math.Int, denom string, msgs ...sdk.Msg) {
	numTestAccounts := len(testAccounts)

	// Track starting balances for all accounts
	granteeBalances := make(sdk.Coins, numTestAccounts)
	funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, denom)
	destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, denom)

	for i, grantee := range testAccounts {
		granteeBalances[i] = s.app.BankKeeper.GetBalance(s.ctx, grantee.address, denom)
	}

	// Validate the AnteHandler passes without issue
	err := validateEthVestingTransactionDecorator(msgs...)
	Expect(err).To(BeNil())

	// Expect delivery to succeed, then compare balances
	_, err = testutil.DeliverEthTx(s.ctx, s.app, nil, msgs...)
	Expect(err).To(BeNil())

	fb := s.app.BankKeeper.GetBalance(s.ctx, funder, denom)
	db := s.app.BankKeeper.GetBalance(s.ctx, dest, denom)

	s.Require().Equal(funderBalance, fb)
	s.Require().Equal(destBalance.AddAmount(amount).Amount.Mul(math.NewInt(int64(numTestAccounts))), db.Amount)

	for i, account := range testAccounts {
		gb := s.app.BankKeeper.GetBalance(s.ctx, account.address, denom)
		// Use GreaterOrEqual because the gas fee is non-recoverable
		s.Require().GreaterOrEqual(granteeBalances[i].SubAmount(amount).Amount.Uint64(), gb.Amount.Uint64())
	}
}

// delegate is a helper function which creates a message to delegate a given amount of tokens
// to a validator and checks if the Cosmos vesting delegation decorator returns no error.
func delegate(clawbackAccount *types.ClawbackVestingAccount, amount math.Int) error {
	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)

	val, err := sdk.ValAddressFromBech32("evmosvaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn")
	s.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(utils.BaseDenom, amount))

	dec := cosmosante.NewVestingDelegationDecorator(s.app.AccountKeeper, s.app.StakingKeeper, types.ModuleCdc)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, delegateMsg)
	return err
}

// validateEthVestingTransactionDecorator is a helper function to execute the eth vesting transaction decorator
// with 1 or more given messages and return any occurring error.
func validateEthVestingTransactionDecorator(msgs ...sdk.Msg) error {
	dec := evmante.NewEthVestingTransactionDecorator(s.app.AccountKeeper, s.app.BankKeeper, s.app.EvmKeeper)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, msgs...)
	return err
}
