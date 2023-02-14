package keeper_test

import (
	"encoding/json"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/tests"
	evmostypes "github.com/evmos/evmos/v11/types"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
	"github.com/evmos/evmos/v11/x/feemarket/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
	dbm "github.com/tendermint/tm-db"
)

func (suite *KeeperTestSuite) SetupApp(checkTx bool) {
	t := suite.T()
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9000-1",
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
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)

	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

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
	header := suite.ctx.BlockHeader()
	suite.app.EndBlock(abci.RequestEndBlock{Height: header.Height})
	_ = suite.app.Commit()

	header.Height++
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// setupTestWithContext sets up a test chain with an example Cosmos send msg,
// given a local (validator config) and a global (feemarket param) minGasPrice
func setupTestWithContext(valMinGasPrice string, minGasPrice sdk.Dec, baseFee sdkmath.Int) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	privKey, msg := setupTest(valMinGasPrice + s.denom)
	params := types.DefaultParams()
	params.MinGasPrice = minGasPrice
	err := s.app.FeeMarketKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	s.app.FeeMarketKeeper.SetBaseFee(s.ctx, baseFee.BigInt())
	s.Commit()

	return privKey, msg
}

func setupTest(localMinGasPrices string) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	setupChain(localMinGasPrices)

	address, privKey := tests.NewAccAddressAndKey()
	amount, ok := sdkmath.NewIntFromString("10000000000000000000")
	s.Require().True(ok)
	initBalance := sdk.Coins{sdk.Coin{
		Denom:  s.denom,
		Amount: amount,
	}}
	err := fundAccount(s.app.BankKeeper, s.ctx, address, initBalance)
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

// fundAccount is a utility function that funds an account by minting and
// sending the coins to the address.
func fundAccount(bankKeeper bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, evmtypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, addr, amounts)
}

func setupChain(localMinGasPricesStr string) {
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
		simapp.EmptyAppOptions{},
		baseapp.SetMinGasPrices(localMinGasPricesStr),
	)

	genesisState := app.NewTestGenesisState(newapp.AppCodec())
	genesisState[types.ModuleName] = newapp.AppCodec().MustMarshalJSON(types.DefaultGenesisState())

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	s.Require().NoError(err)

	// Initialize the chain
	newapp.InitChain(
		abci.RequestInitChain{
			ChainId:         "evmos_9000-1",
			Validators:      []abci.ValidatorUpdate{},
			AppStateBytes:   stateBytes,
			ConsensusParams: app.DefaultConsensusParams,
		},
	)

	s.app = newapp
	s.SetupApp(false)
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
	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		to,
		nil,
		gasLimit,
		gasPrice,
		gasFeeCap,
		gasTipCap,
		data,
		accesses,
	)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}

func prepareEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) []byte {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	s.Require().NoError(err)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	s.Require().True(ok)
	builder.SetExtensionOptions(option)

	err = msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	msgEthereumTx.From = ""
	err = txBuilder.SetMsgs(msgEthereumTx)
	s.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msgEthereumTx.Data)
	s.Require().NoError(err)

	evmDenom := s.app.EvmKeeper.GetParams(s.ctx).EvmDenom
	fees := sdk.Coins{{Denom: evmDenom, Amount: sdkmath.NewIntFromBigInt(txData.Fee())}}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msgEthereumTx.GetGas())

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	return bz
}

func checkEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseCheckTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestCheckTx{Tx: bz}
	res := s.app.CheckTx(req)
	return res
}

func deliverEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func prepareCosmosTx(priv *ethsecp256k1.PrivKey, gasPrice *sdkmath.Int, msgs ...sdk.Msg) []byte {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	if gasPrice == nil {
		_gasPrice := sdkmath.NewInt(1)
		gasPrice = &_gasPrice
	}
	fees := &sdk.Coins{{Denom: s.denom, Amount: gasPrice.MulRaw(1000000)}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.ctx, accountAddress)
	s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	s.Require().NoError(err)

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}

func checkTx(priv *ethsecp256k1.PrivKey, gasPrice *sdkmath.Int, msgs ...sdk.Msg) abci.ResponseCheckTx {
	bz := prepareCosmosTx(priv, gasPrice, msgs...)
	req := abci.RequestCheckTx{Tx: bz}
	res := s.app.CheckTx(req)
	return res
}

func deliverTx(priv *ethsecp256k1.PrivKey, gasPrice *sdkmath.Int, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := prepareCosmosTx(priv, gasPrice, msgs...)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}
