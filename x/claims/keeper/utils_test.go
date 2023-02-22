package keeper_test

import (
	"encoding/json"
	"math/big"
	"time"

	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/contracts"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/server/config"
	"github.com/evmos/evmos/v11/testutil"
	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	evmostypes "github.com/evmos/evmos/v11/types"
	"github.com/evmos/evmos/v11/utils"
	"github.com/evmos/evmos/v11/x/claims/types"
	evm "github.com/evmos/evmos/v11/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v11/x/feemarket/types"
	incentivestypes "github.com/evmos/evmos/v11/x/incentives/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.ClaimsKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	params := types.DefaultParams()
	params.AirdropStartTime = suite.ctx.BlockTime().UTC()
	err = suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
	require.NoError(t, err)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = params.GetClaimsDenom()
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
	suite.validator = validators[0]

	suite.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())
}

func (suite *KeeperTestSuite) SetupTestWithEscrow() {
	suite.SetupTest()
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	coins := sdk.NewCoins(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(10000000)))
	err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
	suite.Require().NoError(err)
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.Commit(suite.ctx, suite.app, t, nil)
	suite.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())

	types.RegisterQueryServer(queryHelper, suite.app.ClaimsKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)
}

func newEthAccount(baseAccount *authtypes.BaseAccount) evmostypes.EthAccount {
	return evmostypes.EthAccount{
		BaseAccount: baseAccount,
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
}

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return sdk.AccAddress(priv.PubKey().Address().Bytes())
}

func govProposal(priv *ethsecp256k1.PrivKey) (uint64, error) {
	contractAddress := deployContract(priv)
	content := incentivestypes.NewRegisterIncentiveProposal(
		"test",
		"description",
		contractAddress.String(),
		sdk.DecCoins{sdk.NewDecCoinFromDec(utils.BaseDenom, sdk.NewDecWithPrec(5, 2))},
		1000,
	)
	return testutil.SubmitProposal(s.ctx, s.app, priv, content, 8)
}

func sendEthToSelf(priv *ethsecp256k1.PrivKey) {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	ethTxParams := evm.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &from,
		GasLimit:  100000,
		GasFeeCap: s.app.FeeMarketKeeper.GetBaseFee(s.ctx),
		GasTipCap: big.NewInt(1),
		Accesses:  &ethtypes.AccessList{},
	}
	msgEthereumTx := evm.NewTx(&ethTxParams)
	msgEthereumTx.From = from.String()
	performEthTx(priv, msgEthereumTx)
}

func deployContract(priv *ethsecp256k1.PrivKey) common.Address {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "Test", "TTT", uint8(18))
	s.Require().NoError(err)

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &s.address,
		Data: (*hexutil.Bytes)(&data),
	})
	s.Require().NoError(err)

	ctx := sdk.WrapSDKContext(s.ctx)
	res, err := s.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	s.Require().NoError(err)

	ethTxParams := evm.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		GasLimit:  res.Gas,
		GasFeeCap: s.app.FeeMarketKeeper.GetBaseFee(s.ctx),
		GasTipCap: big.NewInt(1),
		Input:     data,
		Accesses:  &ethtypes.AccessList{},
	}
	msgEthereumTx := evm.NewTx(&ethTxParams)
	msgEthereumTx.From = from.String()

	performEthTx(priv, msgEthereumTx)
	s.Commit()

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evm.MsgEthereumTx) {
	// Sign transaction
	err := msgEthereumTx.Sign(s.ethSigner, utiltx.NewSigner(priv))
	s.Require().NoError(err)

	// Assemble transaction from fields
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	tx, err := msgEthereumTx.BuildTx(txBuilder, utils.BaseDenom)
	s.Require().NoError(err)

	// Encode transaction by default Tx encoder and broadcasted over the network
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(tx)
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	Expect(res.IsOK()).To(Equal(true))
}

func getEthTxFee() sdk.Coin {
	baseFee := s.app.FeeMarketKeeper.GetBaseFee(s.ctx)
	baseFee.Mul(baseFee, big.NewInt(100000))
	feeAmt := baseFee.Quo(baseFee, big.NewInt(2))
	return sdk.NewCoin(utils.BaseDenom, sdkmath.NewIntFromBigInt(feeAmt))
}
