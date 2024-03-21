package keeper_test

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	evmostypes "github.com/evmos/evmos/v16/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/precompiles/staking"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
)

var templateAccessListTx = &ethtypes.AccessListTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateLegacyTx = &ethtypes.LegacyTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateDynamicFeeTx = &ethtypes.DynamicFeeTx{
	GasFeeCap: big.NewInt(10),
	GasTipCap: big.NewInt(2),
	Gas:       21000,
	To:        &common.Address{},
	Value:     big.NewInt(0),
	Data:      []byte{},
}

func newSignedEthTx(
	txData ethtypes.TxData,
	nonce uint64,
	addr sdk.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
) (*ethtypes.Transaction, error) {
	var ethTx *ethtypes.Transaction
	switch txData := txData.(type) {
	case *ethtypes.AccessListTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.LegacyTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.DynamicFeeTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	default:
		return nil, errors.New("unknown transaction type")
	}

	sig, _, err := krSigner.SignByAddress(addr, ethTx.Hash().Bytes())
	if err != nil {
		return nil, err
	}

	ethTx, err = ethTx.WithSignature(ethSigner, sig)
	if err != nil {
		return nil, err
	}

	return ethTx, nil
}

func newEthMsgTx(
	nonce uint64,
	address common.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	txType byte,
	data []byte,
	accessList ethtypes.AccessList,
) (*evmtypes.MsgEthereumTx, *big.Int, error) {
	var (
		ethTx   *ethtypes.Transaction
		baseFee *big.Int
	)
	switch txType {
	case ethtypes.LegacyTxType:
		templateLegacyTx.Nonce = nonce
		if data != nil {
			templateLegacyTx.Data = data
		}
		ethTx = ethtypes.NewTx(templateLegacyTx)
	case ethtypes.AccessListTxType:
		templateAccessListTx.Nonce = nonce
		if data != nil {
			templateAccessListTx.Data = data
		} else {
			templateAccessListTx.Data = []byte{}
		}

		templateAccessListTx.AccessList = accessList
		ethTx = ethtypes.NewTx(templateAccessListTx)
	case ethtypes.DynamicFeeTxType:
		templateDynamicFeeTx.Nonce = nonce

		if data != nil {
			templateAccessListTx.Data = data
		} else {
			templateAccessListTx.Data = []byte{}
		}
		templateAccessListTx.AccessList = accessList
		ethTx = ethtypes.NewTx(templateDynamicFeeTx)
		baseFee = big.NewInt(3)
	default:
		return nil, baseFee, errors.New("unsupport tx type")
	}

	msg := &evmtypes.MsgEthereumTx{}
	err := msg.FromEthereumTx(ethTx)
	if err != nil {
		return nil, nil, err
	}

	msg.From = address.Hex()

	return msg, baseFee, msg.Sign(ethSigner, krSigner)
}

func newNativeMessage(
	nonce uint64,
	blockHeight int64,
	address common.Address,
	cfg *params.ChainConfig,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	txType byte,
	data []byte,
	accessList ethtypes.AccessList,
) (core.Message, error) {
	msgSigner := ethtypes.MakeSigner(cfg, big.NewInt(blockHeight))

	msg, baseFee, err := newEthMsgTx(nonce, address, krSigner, ethSigner, txType, data, accessList)
	if err != nil {
		return nil, err
	}

	m, err := msg.AsMessage(msgSigner, baseFee)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func BenchmarkApplyTransaction(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateAccessListTx,
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(suite.ctx, tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateLegacyTx,
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(suite.ctx, tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithDynamicFeeTx(b *testing.B) {
	suite := KeeperTestSuite{enableFeemarket: true, enableLondonHF: true}
	suite.SetupTestWithT(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateDynamicFeeTx,
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(suite.ctx, tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

//nolint:all
func BenchmarkApplyMessage(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			ethtypes.AccessListTxType,
			nil,
			nil,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

//nolint:all
func BenchmarkApplyMessageWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			ethtypes.LegacyTxType,
			nil,
			nil,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyMessageWithDynamicFeeTx(b *testing.B) {
	suite := KeeperTestSuite{enableFeemarket: true, enableLondonHF: true}
	suite.SetupTestWithT(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			ethtypes.DynamicFeeTxType,
			nil,
			nil,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

type benchmarkSuite struct {
	network     *network.UnitTestNetwork
	grpcHandler grpc.Handler
	txFactory   factory.TxFactory
	keyring     testkeyring.Keyring
}

var table = []struct {
	txType      string
	dynamicAccs []int
}{
	{
		txType:      "transfer",
		dynamicAccs: []int{1, 50},
	},
	{
		txType:      "deployment",
		dynamicAccs: []int{1, 50},
	},
	{
		txType:      "contract_call",
		dynamicAccs: []int{1, 50},
	},
	{
		txType:      "static_precompile",
		dynamicAccs: []int{1, 50},
	},
	{
		txType:      "dynamic_precompile",
		dynamicAccs: []int{1, 50},
	},
}

func BenchmarkApplyTransactionV2(b *testing.B) {
	for _, v := range table {
		for _, dynamicAccs := range v.dynamicAccs {
			// Reset chain on every tx type to have a clean state
			// and a fair benchmark
			b.StopTimer()
			keyring := testkeyring.New(dynamicAccs)

			// Custom genesis state to add erc20 token pairs for dynamic precompiles
			customGenesisState := generateCustomGenesisState(keyring)

			// Avoid overlapping with dynamic precompiles addresses
			sender := keyring.AddKey()
			recipient := keyring.AddKey()

			// Because we are not going thorugh the ante handler,
			// we need to configure the context to execution mode
			unitNetwork := network.NewUnitTestNetwork(
				network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
				network.WithCustomGenesis(customGenesisState),
			)

			grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
			txFactory := factory.New(unitNetwork, grpcHandler)
			suite := benchmarkSuite{
				network:     unitNetwork,
				grpcHandler: grpcHandler,
				txFactory:   txFactory,
				keyring:     keyring,
			}

			// Disable revenue to avoid gas refund issues
			params := unitNetwork.App.RevenueKeeper.GetParams(unitNetwork.GetContext())
			params.EnableRevenue = false
			err := unitNetwork.App.RevenueKeeper.SetParams(unitNetwork.GetContext(), params)
			if err != nil {
				break
			}

			b.Run(fmt.Sprintf("tx_type_%v_%v", v.txType, dynamicAccs), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					// Start with a clean block
					if err := unitNetwork.NextBlock(); err != nil {
						fmt.Println(err)
						break
					}

					// Generate fresh tx type
					tx, err := suite.generateTxType(v.txType, sender, recipient)
					if err != nil {
						fmt.Println(err)
						break
					}

					gasMeter := evmostypes.NewInfiniteGasMeterWithLimit(tx.Gas())
					ctx := unitNetwork.GetContext().WithGasMeter(gasMeter).
						WithKVGasConfig(storetypes.GasConfig{}).
						WithTransientKVGasConfig(storetypes.GasConfig{})

					b.StartTimer()
					// Run benchmark
					resp, err := unitNetwork.App.EvmKeeper.ApplyTransaction(
						ctx,
						tx,
					)
					b.StopTimer()

					if err != nil {
						panic(err)
					}
					if resp.Failed() {
						panic(err)
					}
				}
			})
		}
	}
}

func (suite *benchmarkSuite) generateTxType(txType string, sender, recipient int) (*ethtypes.Transaction, error) {
	senderKey := suite.keyring.GetKey(sender)

	var args evmtypes.EvmTxArgs

	switch txType {
	case "transfer":
		recipient := suite.keyring.GetAddr(recipient)
		args = evmtypes.EvmTxArgs{
			To:     &recipient,
			Amount: big.NewInt(1000000),
		}
	case "deployment":
		var err error
		args, err = suite.txFactory.GenerateDeployContractArgs(
			senderKey.Addr,
			evmtypes.EvmTxArgs{},
			factory.ContractDeploymentData{
				Contract:        contracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{"Coin", "CTKN", uint8(18)},
			},
		)
		if err != nil {
			return nil, err
		}
	case "contract_call":
		var (
			name     = "Coin Token 2"
			symbol   = "CTKN2"
			decimals = uint8(18)
		)
		wevmosAddr, err := suite.txFactory.DeployContract(
			senderKey.Priv,
			evmtypes.EvmTxArgs{},
			factory.ContractDeploymentData{
				Contract:        contracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{name, symbol, decimals},
			},
		)
		if err != nil {
			return nil, err
		}
		callArgs := evmtypes.EvmTxArgs{
			To: &wevmosAddr,
		}
		args, err = suite.txFactory.GenerateContractCallArgs(callArgs,
			factory.CallArgs{
				ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "mint",
				Args:        []interface{}{suite.keyring.GetAddr(1), big.NewInt(100)},
			},
		)
		if err != nil {
			return nil, err
		}
	case "static_precompile":
		contractAddress := common.HexToAddress(staking.PrecompileAddress)
		txArgs := evmtypes.EvmTxArgs{
			To: &contractAddress,
		}
		contractABI, err := staking.LoadABI()
		if err != nil {
			return nil, err
		}

		validatorAddress := suite.network.GetValidators()[1].OperatorAddress
		callArgs := factory.CallArgs{
			ContractABI: contractABI,
			MethodName:  staking.DelegationMethod,
			Args:        []interface{}{senderKey.Addr, validatorAddress},
		}

		args, err = suite.txFactory.GenerateContractCallArgs(txArgs, callArgs)
		if err != nil {
			return nil, err
		}
	case "dynamic_precompile":
		dynamicContract := suite.keyring.GetAddr(1)
		callArgs := evmtypes.EvmTxArgs{
			To: &dynamicContract,
		}
		var err error
		args, err = suite.txFactory.GenerateContractCallArgs(
			callArgs,
			factory.CallArgs{
				ContractABI: erc20.GetABI(),
				MethodName:  erc20.BalanceOfMethod,
				Args:        []interface{}{suite.keyring.GetAddr(sender)},
			},
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown tx type: %v", txType)
	}

	msg, err := suite.txFactory.GenerateMsgEthereumTx(senderKey.Priv, args)
	if err != nil {
		return nil, err
	}
	signMsg, err := suite.txFactory.SignMsgEthereumTx(senderKey.Priv, msg)
	if err != nil {
		return nil, err
	}

	return signMsg.AsTransaction(), nil
}

func sortPrecompiles(precompiles []string) {
	sort.Slice(precompiles, func(i, j int) bool {
		return precompiles[i] < precompiles[j]
	})
}

const DENOM = "ABCD"

func generateCustomGenesisState(keyring testkeyring.Keyring) network.CustomGenesisState {
	addresses := keyring.GetAllAccs()
	tokenPairs := make([]erc20types.TokenPair, len(addresses))
	precompileAddresses := make([]string, len(addresses))

	for i := range addresses {
		tokenPairs[i] = erc20types.TokenPair{
			Erc20Address:  addresses[i].String(),
			Denom:         DENOM,
			Enabled:       true,
			ContractOwner: erc20types.OWNER_MODULE,
		}
		precompileAddresses[i] = addresses[i].String()
	}

	erc20GenesisState := erc20types.DefaultGenesisState()
	erc20GenesisState.TokenPairs = tokenPairs
	evmGenesisState := evmtypes.DefaultGenesisState()
	sortPrecompiles(precompileAddresses)
	evmGenesisState.Params.ActiveDynamicPrecompiles = precompileAddresses

	return network.CustomGenesisState{
		evmtypes.ModuleName:   evmGenesisState,
		erc20types.ModuleName: erc20GenesisState,
	}
}
