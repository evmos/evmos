package keeper_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/tests"
	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/statedb"
	evm "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/erc20/types"
	"github.com/tharsis/evmos/x/erc20/types/contracts"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx              sdk.Context
	app              *app.Evmos
	queryClientEvm   evm.QueryClient
	queryClient      types.QueryClient
	address          common.Address
	consAddress      sdk.ConsAddress
	clientCtx        client.Context
	ethSigner        ethtypes.Signer
	signer           keyring.Signer
	mintFeeCollector bool
}

// Test helpers
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

	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false
	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)

	// init app
	suite.app = app.Setup(checkTx, feemarketGenesis)

	if suite.mintFeeCollector {
		// mint some coin to fee collector
		coins := sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt(int64(params.TxGas)-1)))
		genesisState := app.ModuleBasics.DefaultGenesis(suite.app.AppCodec())
		balances := []banktypes.Balance{
			{
				Address: suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName).String(),
				Coins:   coins,
			},
		}
		// update total supply
		bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt((int64(params.TxGas)-1)))), []banktypes.Metadata{})
		bz := suite.app.AppCodec().MustMarshalJSON(bankGenesis)
		require.NotNil(t, bz)
		genesisState[banktypes.ModuleName] = suite.app.AppCodec().MustMarshalJSON(bankGenesis)

		// we marshal the genesisState of all module to a byte array
		stateBytes, err := tmjson.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		suite.app.InitChain(
			abci.RequestInitChain{
				ChainId:         "evmos_9000-1",
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

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

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.Erc20Keeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	acc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash().Bytes())))
}

func (suite *KeeperTestSuite) MintFeeCollector(coins sdk.Coins) {
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) DeployContract(name string, symbol string, decimals uint8) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", name, symbol, decimals)
	suite.Require().NoError(err)

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTxContract(
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
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) DeployContractMaliciousDelayed(name string, symbol string) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20MaliciousDelayedContract.ABI.Pack("", big.NewInt(1000000000000000000))
	suite.Require().NoError(err)

	data := append(contracts.ERC20MaliciousDelayedContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTxContract(
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
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}
func (suite *KeeperTestSuite) DeployContractDirectBalanceManipulation(name string, symbol string) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20DirectBalanceManipulationContract.ABI.Pack("", big.NewInt(1000000000000000000))
	suite.Require().NoError(err)

	data := append(contracts.ERC20DirectBalanceManipulationContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTxContract(
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
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) Commit() {
	_ = suite.app.Commit()
	header := suite.ctx.BlockHeader()
	header.Height += 1
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) MintERC20Token(contractAddr, from, to common.Address, amount *big.Int) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("mint", to, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func (suite *KeeperTestSuite) BurnERC20Token(contractAddr, from common.Address, amount *big.Int) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", types.ModuleAddress, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func (suite *KeeperTestSuite) GrantERC20Token(contractAddr, from, to common.Address, role_string string) *evm.MsgEthereumTx {
	// 0xCc508cD0818C85b8b8a1aB4cEEef8d981c8956A6 MINTER_ROLE
	role := crypto.Keccak256([]byte(role_string))
	// needs to be an array not a slice
	var v [32]byte
	copy(v[:], role)

	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("grantRole", v, to)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func (suite *KeeperTestSuite) sendTx(contractAddr, from common.Address, transferData []byte) *evm.MsgEthereumTx {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	args, err := json.Marshal(&evm.TransactionArgs{To: &contractAddr, From: &from, Data: (*hexutil.Bytes)(&transferData)})
	suite.Require().NoError(err)
	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	// Mint the max gas to the FeeCollector to ensure balance in case of refund
	suite.MintFeeCollector(sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt(suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx).Int64()*int64(res.Gas)))))

	ercTransferTx := evm.NewTx(
		chainID,
		nonce,
		&contractAddr,
		nil,
		res.Gas,
		nil,
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(1),
		transferData,
		&ethtypes.AccessList{}, // accesses
	)

	ercTransferTx.From = suite.address.Hex()
	err = ercTransferTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, ercTransferTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return ercTransferTx
}

func (suite *KeeperTestSuite) BalanceOf(contract, account common.Address) interface{} {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, "balanceOf", account)
	if err != nil {
		return nil
	}

	unpacked, err := erc20.Unpack("balanceOf", res.Ret)
	if len(unpacked) == 0 {
		return nil
	}

	return unpacked[0]
}

func (suite *KeeperTestSuite) NameOf(contract common.Address) string {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, "name")
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	unpacked, err := erc20.Unpack("name", res.Ret)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(unpacked)

	return fmt.Sprintf("%v", unpacked[0])
}

func (suite *KeeperTestSuite) TransferERC20Token(contractAddr, from, to common.Address, amount *big.Int) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", to, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
