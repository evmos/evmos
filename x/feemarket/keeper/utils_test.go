package keeper_test

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/evmos/evmos/v15/utils"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v15/encoding"
	"github.com/evmos/evmos/v15/testutil"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	"github.com/evmos/evmos/v15/x/feemarket/types"

	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
)

func (suite *KeeperTestSuite) SetupApp(checkTx bool, chainID string) {
	t := suite.T()
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, suite.consAddress, nil, nil,
	)

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	acc := &evmostypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)

	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = utils.BaseDenom
	err = suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
	require.NoError(t, err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = encodingConfig.Codec
	suite.denom = evmtypes.DefaultEVMDenom
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
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// setupTestWithContext sets up a test chain with an example Cosmos send msg,
// given a local (validator config) and a global (feemarket param) minGasPrice
//
//nolint:unparam
func setupTestWithContext(chainID, valMinGasPrice string, minGasPrice sdk.Dec, baseFee sdkmath.Int) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	privKey, msg := setupTest(valMinGasPrice+s.denom, chainID)
	params := types.DefaultParams()
	params.MinGasPrice = minGasPrice
	err := s.app.FeeMarketKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	s.app.FeeMarketKeeper.SetBaseFee(s.ctx, baseFee.BigInt())
	s.Commit()

	return privKey, msg
}

func setupTest(localMinGasPrices, chainID string) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	setupChain(localMinGasPrices, chainID)

	address, privKey := utiltx.NewAccAddressAndKey()
	amount, ok := sdkmath.NewIntFromString("10000000000000000000")
	s.Require().True(ok)
	initBalance := sdk.Coins{sdk.Coin{
		Denom:  s.denom,
		Amount: amount,
	}}
	err := testutil.FundAccount(s.ctx, s.app.BankKeeper, address, initBalance)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: address.String(),
		ToAddress:   address.String(),
		Amount: sdk.Coins{sdk.Coin{
			Denom:  s.denom,
			Amount: sdkmath.NewInt(10000),
		}},
	}
	s.Commit()
	return privKey, msg
}

func setupChain(localMinGasPricesStr string, chainID string) {
	// Initialize the app, so we can use SetMinGasPrices to set the
	// validator-specific min-gas-prices setting
	db := dbm.NewMemDB()
	newapp := app.NewEvmos(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		5,
		encoding.MakeConfig(app.ModuleBasics),
		simutils.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
		baseapp.SetChainID(chainID),
		baseapp.SetMinGasPrices(localMinGasPricesStr),
	)

	genesisState := app.NewTestGenesisState(newapp.AppCodec())
	genesisState[types.ModuleName] = newapp.AppCodec().MustMarshalJSON(types.DefaultGenesisState())

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	s.Require().NoError(err)

	// Initialize the chain
	newapp.InitChain(
		abci.RequestInitChain{
			ChainId:         chainID,
			Validators:      []abci.ValidatorUpdate{},
			AppStateBytes:   stateBytes,
			ConsensusParams: app.DefaultConsensusParams,
		},
	)

	s.app = newapp
	s.SetupApp(false, chainID)
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
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
	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        to,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Input:     data,
		Accesses:  accesses,
	}
	msgEthereumTx := evmtypes.NewTx(ethTxParams)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}
