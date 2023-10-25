package keeper_test

import (
	"math"
	"strings"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v15/app"
	cosmosante "github.com/evmos/evmos/v15/app/ante/cosmos"
	evmante "github.com/evmos/evmos/v15/app/ante/evm"
	"github.com/evmos/evmos/v15/contracts"
	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v15/encoding"
	"github.com/evmos/evmos/v15/testutil"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
	"github.com/evmos/evmos/v15/utils"
	epochstypes "github.com/evmos/evmos/v15/x/epochs/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	"github.com/evmos/evmos/v15/x/vesting/types"

	"github.com/stretchr/testify/require"
)

func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)
	suite.priv = priv

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// Init app
	chainID := utils.TestnetChainID + "-1"
	suite.app = app.Setup(checkTx, nil, chainID)

	// Set Context
	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, suite.consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

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

	// fund signer acc to pay for tx fees
	amt := sdk.NewInt(int64(math.Pow10(18) * 2))
	err = testutil.FundAccount(
		suite.ctx,
		suite.app.BankKeeper,
		suite.priv.PubKey().Address().Bytes(),
		sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amt)),
	)
	suite.Require().NoError(err)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
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

	// Set correct denom in govKeeper
	govParams := suite.app.GovKeeper.GetParams(suite.ctx)
	govParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1e6)))
	votingPeriod := time.Second
	govParams.VotingPeriod = &votingPeriod
	err = suite.app.GovKeeper.SetParams(suite.ctx, govParams)
	suite.Require().NoError(err, "failed to set gov params")
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.CommitAndCreateNewCtx(suite.ctx, suite.app, t, nil)
	suite.Require().NoError(err)
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
	suite.Commit()
	addr, err := testutil.DeployContract(
		suite.ctx,
		suite.app,
		suite.priv,
		suite.queryClientEvm,
		contracts.ERC20MinterBurnerDecimalsContract,
		name, symbol, decimals,
	)
	suite.Commit()
	return addr, err
}

// assertEthFails is a helper function that takes in 1 or more messages and checks
// that they can neither be validated nor delivered using the EthVesting.
func assertEthFails(msgs ...sdk.Msg) {
	insufficientUnlocked := "insufficient unlocked"

	err := validateEthVestingTransactionDecorator(msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficientUnlocked))

	// Sanity check that delivery fails as well
	_, err = testutil.DeliverEthTx(s.app, nil, msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficientUnlocked))
}

// assertEthSucceeds is a helper function, that checks if 1 or more messages
// can be validated and delivered.
func assertEthSucceeds(testAccounts []TestClawbackAccount, funder sdk.AccAddress, dest sdk.AccAddress, amount sdkmath.Int, denom string, msgs ...sdk.Msg) {
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
	_, err = testutil.DeliverEthTx(s.app, nil, msgs...)
	Expect(err).To(BeNil())

	fb := s.app.BankKeeper.GetBalance(s.ctx, funder, denom)
	db := s.app.BankKeeper.GetBalance(s.ctx, dest, denom)

	s.Require().Equal(funderBalance, fb)
	s.Require().Equal(destBalance.AddAmount(amount).Amount.Mul(sdkmath.NewInt(int64(numTestAccounts))), db.Amount)

	for i, account := range testAccounts {
		gb := s.app.BankKeeper.GetBalance(s.ctx, account.address, denom)
		// Use GreaterOrEqual because the gas fee is non-recoverable
		s.Require().GreaterOrEqual(granteeBalances[i].SubAmount(amount).Amount.Uint64(), gb.Amount.Uint64())
	}
}

// delegate is a helper function which creates a message to delegate a given amount of tokens
// to a validator and checks if the Cosmos vesting delegation decorator returns no error.
func delegate(account TestClawbackAccount, coins sdk.Coins) (*stakingtypes.MsgDelegate, error) {
	msg := stakingtypes.NewMsgDelegate(account.address, s.validator.GetOperator(), coins[0])
	dec := cosmosante.NewVestingDelegationDecorator(s.app.AccountKeeper, s.app.StakingKeeper, s.app.BankKeeper, types.ModuleCdc)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, msg)
	return msg, err
}

// validateEthVestingTransactionDecorator is a helper function to execute the eth vesting transaction decorator
// with 1 or more given messages and return any occurring error.
func validateEthVestingTransactionDecorator(msgs ...sdk.Msg) error {
	dec := evmante.NewEthVestingTransactionDecorator(s.app.AccountKeeper, s.app.BankKeeper, s.app.EvmKeeper)
	err = testutil.ValidateAnteForMsgs(s.ctx, dec, msgs...)
	return err
}
