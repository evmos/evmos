package keeper_test

import (
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/contracts"
	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v15/testutil"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
	"github.com/evmos/evmos/v15/utils"
	"github.com/evmos/evmos/v15/x/claims/types"
	evm "github.com/evmos/evmos/v15/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v15/x/feemarket/types"
	incentivestypes "github.com/evmos/evmos/v15/x/incentives/types"
	"github.com/stretchr/testify/require"
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

	chainID := utils.TestnetChainID + "-1"
	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState(), chainID)
	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

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
	err = suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
	require.NoError(t, err)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
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
	suite.ctx, err = testutil.CommitAndCreateNewCtx(suite.ctx, suite.app, t, nil)
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
	contractAddress, err := testutil.DeployContract(
		s.ctx,
		s.app,
		priv,
		s.queryClientEvm,
		contracts.ERC20MinterBurnerDecimalsContract,
		"Test", "TTT", uint8(18),
	)
	s.Require().NoError(err)
	s.ctx, err = testutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Second*0, nil)
	s.Require().NoError(err)
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
	_, err := testutil.DeliverEthTx(s.app, priv, msgEthereumTx)
	s.Require().NoError(err)
}

func getEthTxFee() sdk.Coin {
	baseFee := s.app.FeeMarketKeeper.GetBaseFee(s.ctx)
	baseFee.Mul(baseFee, big.NewInt(100000))
	feeAmt := baseFee.Quo(baseFee, big.NewInt(2))
	return sdk.NewCoin(utils.BaseDenom, sdkmath.NewIntFromBigInt(feeAmt))
}
