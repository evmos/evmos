package keeper_test

import (
	"math/big"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/testutil"
	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
	"github.com/evmos/evmos/v11/x/revenue/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) SetupApp() {
	t := suite.T()
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)

	suite.denom = utils.BaseDenom

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(privCons.PubKey().Address())
	header := testutil.NewHeader(
		1, time.Now().UTC(), "evmos_9001-1", suite.consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(false, header)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.RevenueKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelperEvm)

	params := types.DefaultParams()
	params.EnableRevenue = true
	err = suite.app.RevenueKeeper.SetParams(suite.ctx, params)
	require.NoError(t, err)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = suite.denom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	evmParams.EvmDenom = suite.denom
	err = suite.app.EvmKeeper.SetParams(suite.ctx, evmParams)
	require.NoError(t, err)

	inflationParams := suite.app.InflationKeeper.GetParams(suite.ctx)
	inflationParams.EnableInflation = false
	err = suite.app.InflationKeeper.SetParams(suite.ctx, inflationParams)
	require.NoError(t, err)

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

	types.RegisterQueryServer(queryHelper, suite.app.RevenueKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelperEvm)
}

func calculateFees(
	denom string,
	params types.Params,
	res abci.ResponseDeliverTx,
	gasPrice *big.Int,
) (sdk.Coin, sdk.Coin) {
	feeDistribution := sdk.NewInt(res.GasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
	developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
	developerCoins := sdk.NewCoin(denom, developerFee.TruncateInt())
	validatorShares := sdk.OneDec().Sub(params.DeveloperShares)
	validatorFee := sdk.NewDecFromInt(feeDistribution).Mul(validatorShares)
	validatorCoins := sdk.NewCoin(denom, validatorFee.TruncateInt())
	return developerCoins, validatorCoins
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
}

func registerFee(
	priv *ethsecp256k1.PrivKey,
	contractAddress *common.Address,
	withdrawerAddress sdk.AccAddress,
	nonces []uint64,
) abci.ResponseDeliverTx {
	deployerAddress := sdk.AccAddress(priv.PubKey().Address())
	msg := types.NewMsgRegisterRevenue(*contractAddress, deployerAddress, withdrawerAddress, nonces)

	res, err := testutil.DeliverTx(s.ctx, s.app, priv, nil, msg)
	s.Require().NoError(err)
	s.Commit()

	if res.IsOK() {
		registerEvent := res.GetEvents()[8]
		Expect(registerEvent.Type).To(Equal(types.EventTypeRegisterRevenue))
		Expect(string(registerEvent.Attributes[0].Key)).To(Equal(sdk.AttributeKeySender))
		Expect(string(registerEvent.Attributes[1].Key)).To(Equal(types.AttributeKeyContract))
		Expect(string(registerEvent.Attributes[2].Key)).To(Equal(types.AttributeKeyWithdrawerAddress))
	}
	return res
}

func deployContractWithFactory(priv *ethsecp256k1.PrivKey, factoryAddress *common.Address) common.Address {
	factoryNonce := getNonce(factoryAddress.Bytes())
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)

	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    nonce,
		To:       factoryAddress,
		GasLimit: 100000,
		GasPrice: big.NewInt(1000000000),
		Input:    data,
	}
	msgEthereumTx := evmtypes.NewTx(&ethTxParams)
	msgEthereumTx.From = from.String()

	res, err := testutil.DeliverEthTx(s.app, priv, msgEthereumTx)
	Expect(err).To(BeNil())
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	s.Commit()

	ethereumTx := res.GetEvents()[12]
	Expect(ethereumTx.Type).To(Equal("tx_log"))
	Expect(string(ethereumTx.Attributes[0].Key)).To(Equal("txLog"))
	txLog := string(ethereumTx.Attributes[0].Value)

	contractAddress := crypto.CreateAddress(*factoryAddress, factoryNonce)
	Expect(
		strings.Contains(txLog, strings.ToLower(contractAddress.String()[2:])),
	).To(BeTrue(), "log topic does not match created contract address")

	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc, "contract not created")
	s.Require().True(acc.IsContract(), "not a contract")
	return contractAddress
}

func deployContract(priv *ethsecp256k1.PrivKey, contractCode string) common.Address {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())

	data := common.Hex2Bytes(contractCode)
	gasLimit := uint64(100000)
	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		GasLimit:  gasLimit,
		GasTipCap: big.NewInt(1),
		GasFeeCap: s.app.FeeMarketKeeper.GetBaseFee(s.ctx),
		Input:     data,
		Accesses:  &ethtypes.AccessList{},
	}
	msgEthereumTx := evmtypes.NewTx(&ethTxParams)
	msgEthereumTx.From = from.String()

	res, err := testutil.DeliverEthTx(s.app, priv, msgEthereumTx)
	Expect(err).To(BeNil())
	s.Commit()

	ethereumTx := res.GetEvents()[11]
	Expect(ethereumTx.Type).To(Equal("ethereum_tx"))
	Expect(string(ethereumTx.Attributes[1].Key)).To(Equal("ethereumTxHash"))

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func contractInteract(
	priv *ethsecp256k1.PrivKey,
	contractAddr *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) abci.ResponseDeliverTx {
	msgEthereumTx := buildEthTx(priv, contractAddr, gasPrice, gasFeeCap, gasTipCap, accesses)
	res, err := testutil.DeliverEthTx(s.app, priv, msgEthereumTx)
	Expect(err).To(BeNil())
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	return res
}

func buildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        to,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Input:     data,
		Accesses:  accesses,
	}
	msgEthereumTx := evmtypes.NewTx(&ethTxParams)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}
