package keeper_test

import (
	"fmt"
	"math"
	"math/big"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/evm/keeper"
	"github.com/evmos/evmos/v16/x/evm/statedb"
	"github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *KeeperTestSuite) TestGetHashFn() {
	header := suite.network.GetContext().BlockHeader()
	h, _ := cmttypes.HeaderFromProto(&header)
	hash := h.Hash()

	testCases := []struct {
		msg      string
		height   uint64
		malleate func() sdk.Context
		expHash  common.Hash
	}{
		{
			"case 1.1: context hash cached",
			uint64(suite.network.GetContext().BlockHeight()),
			func() sdk.Context {
				return suite.network.GetContext().WithHeaderHash(tmhash.Sum([]byte("header")))
			},
			common.BytesToHash(tmhash.Sum([]byte("header"))),
		},
		{
			"case 1.2: failed to cast Tendermint header",
			uint64(suite.network.GetContext().BlockHeight()),
			func() sdk.Context {
				header := tmproto.Header{}
				header.Height = suite.network.GetContext().BlockHeight()
				return suite.network.GetContext().WithBlockHeader(header)
			},
			common.Hash{},
		},
		{
			"case 1.3: hash calculated from Tendermint header",
			uint64(suite.network.GetContext().BlockHeight()),
			func() sdk.Context {
				return suite.network.GetContext().WithBlockHeader(header)
			},
			common.BytesToHash(hash),
		},
		{
			"case 2.1: height lower than current one, hist info not found",
			1,
			func() sdk.Context {
				return suite.network.GetContext().WithBlockHeight(10)
			},
			common.Hash{},
		},
		{
			"case 2.2: height lower than current one, invalid hist info header",
			1,
			func() sdk.Context {
				suite.network.App.StakingKeeper.SetHistoricalInfo(suite.network.GetContext(), 1, &stakingtypes.HistoricalInfo{})
				return suite.network.GetContext().WithBlockHeight(10)
			},
			common.Hash{},
		},
		{
			"case 2.3: height lower than current one, calculated from hist info header",
			1,
			func() sdk.Context {
				histInfo := &stakingtypes.HistoricalInfo{
					Header: header,
				}
				suite.network.App.StakingKeeper.SetHistoricalInfo(suite.network.GetContext(), 1, histInfo)
				return suite.network.GetContext().WithBlockHeight(10)
			},
			common.BytesToHash(hash),
		},
		{
			"case 3: height greater than current one",
			200,
			func() sdk.Context { return suite.network.GetContext() },
			common.Hash{},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			ctx := tc.malleate()

			hash := suite.network.App.EvmKeeper.GetHashFn(ctx)(tc.height)
			suite.Require().Equal(tc.expHash, hash)
		})
	}
}

func (suite *KeeperTestSuite) TestGetCoinbaseAddress() {
	valOpAddr := utiltx.GenerateAddress()

	testCases := []struct {
		msg      string
		malleate func() sdk.Context
		expPass  bool
	}{
		{
			"validator not found",
			func() sdk.Context {
				header := suite.network.GetContext().BlockHeader()
				header.ProposerAddress = []byte{}
				return suite.network.GetContext().WithBlockHeader(header)
			},
			false,
		},
		{
			"success",
			func() sdk.Context {
				valConsAddr, privkey := utiltx.NewAddrKey()

				pkAny, err := codectypes.NewAnyWithValue(privkey.PubKey())
				suite.Require().NoError(err)

				validator := stakingtypes.Validator{
					OperatorAddress: sdk.ValAddress(valOpAddr.Bytes()).String(),
					ConsensusPubkey: pkAny,
				}

				suite.network.App.StakingKeeper.SetValidator(suite.network.GetContext(), validator)
				err = suite.network.App.StakingKeeper.SetValidatorByConsAddr(suite.network.GetContext(), validator)
				suite.Require().NoError(err)

				header := suite.network.GetContext().BlockHeader()
				header.ProposerAddress = valConsAddr.Bytes()
				ctx := suite.network.GetContext().WithBlockHeader(header)

				_, err = suite.network.App.StakingKeeper.GetValidatorByConsAddr(ctx, valConsAddr.Bytes())
				suite.Require().NoError(err)

				suite.Require().NotEmpty(ctx.BlockHeader().ProposerAddress)
				return ctx
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			ctx := tc.malleate()
			proposerAddress := ctx.BlockHeader().ProposerAddress
			coinbase, err := suite.network.App.EvmKeeper.GetCoinbaseAddress(ctx, sdk.ConsAddress(proposerAddress))
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(valOpAddr, coinbase)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetEthIntrinsicGas() {
	testCases := []struct {
		name               string
		data               []byte
		accessList         ethtypes.AccessList
		height             int64
		isContractCreation bool
		noError            bool
		expGas             uint64
	}{
		{
			"no data, no accesslist, not contract creation, not homestead, not istanbul",
			nil,
			nil,
			1,
			false,
			true,
			params.TxGas,
		},
		{
			"with one zero data, no accesslist, not contract creation, not homestead, not istanbul",
			[]byte{0},
			nil,
			1,
			false,
			true,
			params.TxGas + params.TxDataZeroGas*1,
		},
		{
			"with one non zero data, no accesslist, not contract creation, not homestead, not istanbul",
			[]byte{1},
			nil,
			1,
			true,
			true,
			params.TxGas + params.TxDataNonZeroGasFrontier*1,
		},
		{
			"no data, one accesslist, not contract creation, not homestead, not istanbul",
			nil,
			[]ethtypes.AccessTuple{
				{},
			},
			1,
			false,
			true,
			params.TxGas + params.TxAccessListAddressGas,
		},
		{
			"no data, one accesslist with one storageKey, not contract creation, not homestead, not istanbul",
			nil,
			[]ethtypes.AccessTuple{
				{StorageKeys: make([]common.Hash, 1)},
			},
			1,
			false,
			true,
			params.TxGas + params.TxAccessListAddressGas + params.TxAccessListStorageKeyGas*1,
		},
		{
			"no data, no accesslist, is contract creation, is homestead, not istanbul",
			nil,
			nil,
			2,
			true,
			true,
			params.TxGasContractCreation,
		},
		{
			"with one zero data, no accesslist, not contract creation, is homestead, is istanbul",
			[]byte{1},
			nil,
			3,
			false,
			true,
			params.TxGas + params.TxDataNonZeroGasEIP2028*1,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
			ethCfg := params.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
			ethCfg.HomesteadBlock = big.NewInt(2)
			ethCfg.IstanbulBlock = big.NewInt(3)
			signer := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())

			ctx := suite.network.GetContext().WithBlockHeight(tc.height)

			addr := suite.keyring.GetAddr(0)
			krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
			nonce := suite.network.App.EvmKeeper.GetNonce(ctx, addr)
			m, err := newNativeMessage(
				nonce,
				ctx.BlockHeight(),
				addr,
				ethCfg,
				krSigner,
				signer,
				ethtypes.AccessListTxType,
				tc.data,
				tc.accessList,
			)
			suite.Require().NoError(err)

			gas, err := suite.network.App.EvmKeeper.GetEthIntrinsicGas(ctx, m, ethCfg, tc.isContractCreation)
			if tc.noError {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			suite.Require().Equal(tc.expGas, gas)
		})
	}
}

func (suite *KeeperTestSuite) TestGasToRefund() {
	testCases := []struct {
		name           string
		gasconsumed    uint64
		refundQuotient uint64
		expGasRefund   uint64
		expPanic       bool
	}{
		{
			"gas refund 5",
			5,
			1,
			5,
			false,
		},
		{
			"gas refund 10",
			10,
			1,
			10,
			false,
		},
		{
			"gas refund availableRefund",
			11,
			1,
			10,
			false,
		},
		{
			"gas refund quotient 0",
			11,
			0,
			0,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest() // reset
			vmdb := suite.StateDB()
			vmdb.AddRefund(10)

			if tc.expPanic {
				panicF := func() {
					//nolint:staticcheck
					keeper.GasToRefund(vmdb.GetRefund(), tc.gasconsumed, tc.refundQuotient)
				}
				suite.Require().Panics(panicF)
			} else {
				gr := keeper.GasToRefund(vmdb.GetRefund(), tc.gasconsumed, tc.refundQuotient)
				suite.Require().Equal(tc.expGasRefund, gr)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestRefundGas() {
	var (
		m   core.Message
		err error
	)

	testCases := []struct {
		name           string
		leftoverGas    uint64
		refundQuotient uint64
		noError        bool
		expGasRefund   uint64
		malleate       func()
	}{
		{
			name:           "leftoverGas more than tx gas limit",
			leftoverGas:    params.TxGas + 1,
			refundQuotient: params.RefundQuotient,
			noError:        false,
			expGasRefund:   params.TxGas + 1,
		},
		{
			name:           "leftoverGas equal to tx gas limit, insufficient fee collector account",
			leftoverGas:    params.TxGas,
			refundQuotient: params.RefundQuotient,
			noError:        true,
			expGasRefund:   0,
		},
		{
			name:           "leftoverGas less than to tx gas limit",
			leftoverGas:    params.TxGas - 1,
			refundQuotient: params.RefundQuotient,
			noError:        true,
			expGasRefund:   0,
		},
		{
			name:           "no leftoverGas, refund half used gas ",
			leftoverGas:    0,
			refundQuotient: params.RefundQuotient,
			noError:        true,
			expGasRefund:   params.TxGas / params.RefundQuotient,
		},
		{
			name:           "invalid Gas value in msg",
			leftoverGas:    0,
			refundQuotient: params.RefundQuotient,
			noError:        false,
			expGasRefund:   params.TxGas,
			malleate: func() {
				keeperParams := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				m, err = suite.createContractGethMsg(
					suite.StateDB().GetNonce(suite.keyring.GetAddr(0)),
					ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID()),
					keeperParams.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID()),
					big.NewInt(-100),
				)
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest() // reset

			keeperParams := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
			ethCfg := keeperParams.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
			signer := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())
			vmdb := suite.StateDB()

			addr := suite.keyring.GetAddr(0)
			krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
			m, err = newNativeMessage(
				vmdb.GetNonce(addr),
				suite.network.GetContext().BlockHeight(),
				addr,
				ethCfg,
				krSigner,
				signer,
				ethtypes.AccessListTxType,
				nil,
				nil,
			)
			suite.Require().NoError(err)

			vmdb.AddRefund(params.TxGas)

			if tc.leftoverGas > m.Gas() {
				return
			}

			if tc.malleate != nil {
				tc.malleate()
			}

			gasUsed := m.Gas() - tc.leftoverGas
			refund := keeper.GasToRefund(vmdb.GetRefund(), gasUsed, tc.refundQuotient)
			suite.Require().Equal(tc.expGasRefund, refund)

			err = suite.network.App.EvmKeeper.RefundGas(suite.network.GetContext(), m, refund, types.DefaultEVMDenom)
			if tc.noError {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestResetGasMeterAndConsumeGas() {
	testCases := []struct {
		name        string
		gasConsumed uint64
		gasUsed     uint64
		expPanic    bool
	}{
		{
			"gas consumed 5, used 5",
			5,
			5,
			false,
		},
		{
			"gas consumed 5, used 10",
			5,
			10,
			false,
		},
		{
			"gas consumed 10, used 10",
			10,
			10,
			false,
		},
		{
			"gas consumed 11, used 10, NegativeGasConsumed panic",
			11,
			10,
			true,
		},
		{
			"gas consumed 1, used 10, overflow panic",
			1,
			math.MaxUint64,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			panicF := func() {
				gm := storetypes.NewGasMeter(10)
				gm.ConsumeGas(tc.gasConsumed, "")
				ctx := suite.network.GetContext().WithGasMeter(gm)
				suite.network.App.EvmKeeper.ResetGasMeterAndConsumeGas(ctx, tc.gasUsed)
			}

			if tc.expPanic {
				suite.Require().Panics(panicF)
			} else {
				suite.Require().NotPanics(panicF)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEVMConfig() {
	proposerAddress := suite.network.GetContext().BlockHeader().ProposerAddress
	cfg, err := suite.network.App.EvmKeeper.EVMConfig(suite.network.GetContext(), proposerAddress, big.NewInt(9000))
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultParams(), cfg.Params)
	// london hardfork is enabled by default
	suite.Require().Equal(big.NewInt(0), cfg.BaseFee)
	suite.Require().Equal(suite.keyring.GetAddr(0), cfg.CoinBase)
	suite.Require().Equal(types.DefaultParams().ChainConfig.EthereumConfig(big.NewInt(9000)), cfg.ChainConfig)
}

func (suite *KeeperTestSuite) TestContractDeployment() {
	contractAddress := suite.DeployTestContract(suite.T(), suite.keyring.GetAddr(0), big.NewInt(10000000000000))
	db := suite.StateDB()
	suite.Require().Greater(db.GetCodeSize(contractAddress), 0)
}

func (suite *KeeperTestSuite) TestApplyMessage() {
	expectedGasUsed := params.TxGas
	var msg core.Message

	proposerAddress := suite.network.GetContext().BlockHeader().ProposerAddress
	config, err := suite.network.App.EvmKeeper.EVMConfig(suite.network.GetContext(), proposerAddress, big.NewInt(9000))
	suite.Require().NoError(err)

	keeperParams := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
	chainCfg := keeperParams.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())
	tracer := suite.network.App.EvmKeeper.Tracer(suite.network.GetContext(), msg, config.ChainConfig)
	vmdb := suite.StateDB()

	addr := suite.keyring.GetAddr(0)
	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	msg, err = newNativeMessage(
		vmdb.GetNonce(addr),
		suite.network.GetContext().BlockHeight(),
		addr,
		chainCfg,
		krSigner,
		signer,
		ethtypes.AccessListTxType,
		nil,
		nil,
	)
	suite.Require().NoError(err)

	res, err := suite.network.App.EvmKeeper.ApplyMessage(suite.network.GetContext(), msg, tracer, true)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedGasUsed, res.GasUsed)
	suite.Require().False(res.Failed())
}

func (suite *KeeperTestSuite) TestApplyMessageWithConfig() {
	var (
		msg             core.Message
		err             error
		expectedGasUsed uint64
		config          *statedb.EVMConfig
		keeperParams    types.Params
		signer          ethtypes.Signer
		vmdb            *statedb.StateDB
		txConfig        statedb.TxConfig
		chainCfg        *params.ChainConfig
	)

	testCases := []struct {
		name     string
		malleate func()
		expErr   bool
	}{
		{
			"messsage applied ok",
			func() {
				addr := suite.keyring.GetAddr(0)
				krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
				msg, err = newNativeMessage(
					vmdb.GetNonce(addr),
					suite.network.GetContext().BlockHeight(),
					addr,
					chainCfg,
					krSigner,
					signer,
					ethtypes.AccessListTxType,
					nil,
					nil,
				)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"call contract tx with config param EnableCall = false",
			func() {
				config.Params.EnableCall = false
				addr := suite.keyring.GetAddr(0)
				krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
				msg, err = newNativeMessage(
					vmdb.GetNonce(addr),
					suite.network.GetContext().BlockHeight(),
					addr,
					chainCfg,
					krSigner,
					signer,
					ethtypes.AccessListTxType,
					nil,
					nil,
				)
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"create contract tx with config param EnableCreate = false",
			func() {
				msg, err = suite.createContractGethMsg(vmdb.GetNonce(suite.keyring.GetAddr(0)), signer, chainCfg, big.NewInt(1))
				suite.Require().NoError(err)
				config.Params.EnableCreate = false
			},
			true,
		},
		{
			"fix panic when minimumGasUsed is not uint64",
			func() {
				addr := suite.keyring.GetAddr(0)
				krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
				msg, err = newNativeMessage(
					vmdb.GetNonce(addr),
					suite.network.GetContext().BlockHeight(),
					addr,
					chainCfg,
					krSigner,
					signer,
					ethtypes.AccessListTxType,
					nil,
					nil,
				)
				suite.Require().NoError(err)
				params := suite.network.App.FeeMarketKeeper.GetParams(suite.network.GetContext())
				params.MinGasMultiplier = sdkmath.LegacyNewDec(math.MaxInt64).MulInt64(100)
				err = suite.network.App.FeeMarketKeeper.SetParams(suite.network.GetContext(), params)
				suite.Require().NoError(err)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			expectedGasUsed = params.TxGas

			proposerAddress := suite.network.GetContext().BlockHeader().ProposerAddress
			config, err = suite.network.App.EvmKeeper.EVMConfig(suite.network.GetContext(), proposerAddress, big.NewInt(9000))
			suite.Require().NoError(err)

			keeperParams = suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
			chainCfg = keeperParams.ChainConfig.EthereumConfig(suite.network.App.EvmKeeper.ChainID())
			signer = ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())
			vmdb = suite.StateDB()
			txConfig = suite.network.App.EvmKeeper.TxConfig(suite.network.GetContext(), common.Hash{})

			tc.malleate()
			res, err := suite.network.App.EvmKeeper.ApplyMessageWithConfig(suite.network.GetContext(), msg, nil, true, config, txConfig)

			if tc.expErr {
				suite.Require().Error(err)
				return
			}

			suite.Require().NoError(err)
			suite.Require().False(res.Failed())
			suite.Require().Equal(expectedGasUsed, res.GasUsed)
		})
	}
}

func (suite *KeeperTestSuite) createContractGethMsg(nonce uint64, signer ethtypes.Signer, cfg *params.ChainConfig, gasPrice *big.Int) (core.Message, error) {
	ethMsg, err := suite.createContractMsgTx(nonce, signer, gasPrice)
	if err != nil {
		return nil, err
	}

	msgSigner := ethtypes.MakeSigner(cfg, big.NewInt(suite.network.GetContext().BlockHeight()))
	return ethMsg.AsMessage(msgSigner, nil)
}

func (suite *KeeperTestSuite) createContractMsgTx(nonce uint64, signer ethtypes.Signer, gasPrice *big.Int) (*types.MsgEthereumTx, error) {
	contractCreateTx := &ethtypes.AccessListTx{
		GasPrice: gasPrice,
		Gas:      params.TxGasContractCreation,
		To:       nil,
		Data:     []byte("contract_data"),
		Nonce:    nonce,
	}
	ethTx := ethtypes.NewTx(contractCreateTx)
	ethMsg := &types.MsgEthereumTx{}
	err := ethMsg.FromEthereumTx(ethTx)
	suite.Require().NoError(err)
	ethMsg.From = suite.keyring.GetAddr(0).Hex()
	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	return ethMsg, ethMsg.Sign(signer, krSigner)
}

func (suite *KeeperTestSuite) TestGetProposerAddress() {
	var a sdk.ConsAddress
	address := sdk.ConsAddress(suite.keyring.GetAddr(0).Bytes())
	proposerAddress := sdk.ConsAddress(suite.network.GetContext().BlockHeader().ProposerAddress)
	testCases := []struct {
		msg    string
		adr    sdk.ConsAddress
		expAdr sdk.ConsAddress
	}{
		{
			"proposer address provided",
			address,
			address,
		},
		{
			"nil proposer address provided",
			nil,
			proposerAddress,
		},
		{
			"typed nil proposer address provided",
			a,
			proposerAddress,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.Require().Equal(tc.expAdr, keeper.GetProposerAddress(suite.network.GetContext(), tc.adr))
		})
	}
}
