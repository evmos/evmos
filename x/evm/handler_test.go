package evm_test

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/x/evm/keeper"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/evm/statedb"
	"github.com/evmos/evmos/v16/x/evm/types"
)

// TODO move these to msg_server_test.go
// because Handler was deprecated

type EvmTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring testkeyring.Keyring
	factory factory.TxFactory
	server  types.MsgServer

	dynamicTxFee bool
}

// DoSetupTest setup test environment, it uses`require.TestingT` to support both `testing.T` and `testing.B`.
func (suite *EvmTestSuite) DoSetupTest(_ require.TestingT) {
	keys := testkeyring.New(2)
	// Set custom balance based on test params
	customGenesis := network.CustomGenesisState{}
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	if suite.dynamicTxFee {
		feemarketGenesis.Params.EnableHeight = 1
		feemarketGenesis.Params.NoBaseFee = false
	} else {
		feemarketGenesis.Params.NoBaseFee = true
	}
	customGenesis[feemarkettypes.ModuleName] = feemarketGenesis

	// mint some coin to fee collector
	// to pay gas refunds
	coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromUint64(math.MaxUint64)))
	balances := []banktypes.Balance{
		{
			Address: authtypes.NewModuleAddress(authtypes.FeeCollectorName).String(),
			Coins:   coins,
		},
	}
	bankGenesis := banktypes.DefaultGenesisState()
	bankGenesis.Balances = balances
	customGenesis[banktypes.ModuleName] = bankGenesis

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	suite.network = nw
	suite.factory = tf
	suite.handler = gh
	suite.keyring = keys
	suite.server = nw.App.EvmKeeper
}

func (suite *EvmTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *EvmTestSuite) SignTx(tx *types.MsgEthereumTx) {
	krSigner := utiltx.NewSigner(suite.keyring.GetPrivKey(0))
	tx.From = suite.keyring.GetAddr(0).Hex()
	err := tx.Sign(ethtypes.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID()), krSigner)
	suite.Require().NoError(err)
}

func (suite *EvmTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.network.GetContext(), suite.network.App.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.network.GetContext().HeaderHash())))
}

func TestEvmTestSuite(t *testing.T) {
	suite.Run(t, new(EvmTestSuite))
}

func (suite *EvmTestSuite) TestHandleMsgEthereumTx() {
	var (
		tx  *types.MsgEthereumTx
		ctx sdk.Context
	)

	defaultEthTxParams := &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    0,
		Amount:   big.NewInt(100),
		GasLimit: 0,
		GasPrice: big.NewInt(10000),
	}

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"passed",
			func() {
				to := suite.keyring.GetAddr(1)
				ethTxParams := &types.EvmTxArgs{
					ChainID:  suite.network.App.EvmKeeper.ChainID(),
					Nonce:    0,
					To:       &to,
					Amount:   big.NewInt(10),
					GasLimit: 10_000_000,
					GasPrice: big.NewInt(10000),
				}
				tx = types.NewTx(ethTxParams)
				suite.SignTx(tx)
			},
			true,
		},
		{
			"insufficient balance",
			func() {
				tx = types.NewTx(defaultEthTxParams)
				suite.SignTx(tx)
			},
			false,
		},
		{
			"tx encoding failed",
			func() {
				tx = types.NewTx(defaultEthTxParams)
			},
			false,
		},
		{
			"invalid chain ID",
			func() {
				ctx = ctx.WithChainID("chainID")
			},
			false,
		},
		{
			"VerifySig failed",
			func() {
				tx = types.NewTx(defaultEthTxParams)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset
			ctx = suite.network.GetContext()

			tc.malleate()
			res, err := suite.server.EthereumTx(ctx, tx)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *EvmTestSuite) TestHandlerLogs() {
	// Test contract:

	// pragma solidity ^0.5.1;

	// contract Test {
	//     event Hello(uint256 indexed world);

	//     constructor() public {
	//         emit Hello(17);
	//     }
	// }

	// {
	// 	"linkReferences": {},
	// 	"object": "6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029",
	// 	"opcodes": "PUSH1 0x80 PUSH1 0x40 MSTORE CALLVALUE DUP1 ISZERO PUSH1 0xF JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH1 0x11 PUSH32 0x775A94827B8FD9B519D36CD827093C664F93347070A554F65E4A6F56CD738898 PUSH1 0x40 MLOAD PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 LOG2 PUSH1 0x35 DUP1 PUSH1 0x4B PUSH1 0x0 CODECOPY PUSH1 0x0 RETURN INVALID PUSH1 0x80 PUSH1 0x40 MSTORE PUSH1 0x0 DUP1 REVERT INVALID LOG1 PUSH6 0x627A7A723058 KECCAK256 PUSH13 0xAB665F0F557620554BB45ADF26 PUSH8 0x8D2BD349B8A4314 0xbd SELFDESTRUCT KECCAK256 0x5e 0xe8 DIFFICULTY 0xe EXTCODECOPY 0x24 STOP 0x29 ",
	// 	"sourceMap": "25:119:0:-;;;90:52;8:9:-1;5:2;;;30:1;27;20:12;5:2;90:52:0;132:2;126:9;;;;;;;;;;25:119;;;;;;"
	// }

	gasLimit := uint64(100000)
	gasPrice := big.NewInt(1000000)

	bytecode := common.FromHex("0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029")

	ethTxParams := &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Input:    bytecode,
	}
	tx := types.NewTx(ethTxParams)
	suite.SignTx(tx)

	result, err := suite.server.EthereumTx(suite.network.GetContext(), tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")

	suite.Require().Equal(len(result.Logs), 1)
	suite.Require().Equal(len(result.Logs[0].Topics), 2)
}

func (suite *EvmTestSuite) TestDeployAndCallContract() {
	// Test contract:
	// http://remix.ethereum.org/#optimize=false&evmVersion=istanbul&version=soljson-v0.5.15+commit.6a57276f.js
	// 2_Owner.sol
	//
	// pragma solidity >=0.4.22 <0.7.0;
	//
	///**
	// * @title Owner
	// * @dev Set & change owner
	// */
	// contract Owner {
	//
	//	address private owner;
	//
	//	// event for EVM logging
	//	event OwnerSet(address indexed oldOwner, address indexed newOwner);
	//
	//	// modifier to check if caller is owner
	//	modifier isOwner() {
	//	// If the first argument of 'require' evaluates to 'false', execution terminates and all
	//	// changes to the state and to Ether balances are reverted.
	//	// This used to consume all gas in old EVM versions, but not anymore.
	//	// It is often a good idea to use 'require' to check if functions are called correctly.
	//	// As a second argument, you can also provide an explanation about what went wrong.
	//	require(msg.sender == owner, "Caller is not owner");
	//	_;
	//}
	//
	//	/**
	//	 * @dev Set contract deployer as owner
	//	 */
	//	constructor() public {
	//	owner = msg.sender; // 'msg.sender' is sender of current call, contract deployer for a constructor
	//	emit OwnerSet(address(0), owner);
	//}
	//
	//	/**
	//	 * @dev Change owner
	//	 * @param newOwner address of new owner
	//	 */
	//	function changeOwner(address newOwner) public isOwner {
	//	emit OwnerSet(owner, newOwner);
	//	owner = newOwner;
	//}
	//
	//	/**
	//	 * @dev Return owner address
	//	 * @return address of owner
	//	 */
	//	function getOwner() external view returns (address) {
	//	return owner;
	//}
	//}

	// Deploy contract - Owner.sol
	gasLimit := uint64(100000000)
	gasPrice := big.NewInt(10000)

	bytecode := common.FromHex("0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a36102c4806100dc6000396000f3fe608060405234801561001057600080fd5b5060043610610053576000357c010000000000000000000000000000000000000000000000000000000090048063893d20e814610058578063a6f9dae1146100a2575b600080fd5b6100606100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100e4600480360360208110156100b857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061010f565b005b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146101d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f43616c6c6572206973206e6f74206f776e65720000000000000000000000000081525060200191505060405180910390fd5b8073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a3806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505056fea265627a7a72315820f397f2733a89198bc7fed0764083694c5b828791f39ebcbc9e414bccef14b48064736f6c63430005100032")
	ethTxParams := &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Input:    bytecode,
	}
	tx := types.NewTx(ethTxParams)
	suite.SignTx(tx)

	ctx := suite.network.GetContext()

	res, err := suite.server.EthereumTx(ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")
	suite.Require().Equal(res.VmError, "", "failed to handle eth tx msg")

	// store - changeOwner
	gasLimit = uint64(100000000000)
	gasPrice = big.NewInt(100)
	receiver := crypto.CreateAddress(suite.keyring.GetAddr(0), 1)

	storeAddr := "0xa6f9dae10000000000000000000000006a82e4a67715c8412a9114fbd2cbaefbc8181424"
	bytecode = common.FromHex(storeAddr)

	ethTxParams = &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    2,
		To:       &receiver,
		Amount:   big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Input:    bytecode,
	}
	tx = types.NewTx(ethTxParams)
	suite.SignTx(tx)

	res, err = suite.server.EthereumTx(ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")
	suite.Require().Equal(res.VmError, "", "failed to handle eth tx msg")

	// query - getOwner
	bytecode = common.FromHex("0x893d20e8")

	ethTxParams = &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    2,
		To:       &receiver,
		Amount:   big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Input:    bytecode,
	}
	tx = types.NewTx(ethTxParams)
	suite.SignTx(tx)

	res, err = suite.server.EthereumTx(ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")
	suite.Require().Equal(res.VmError, "", "failed to handle eth tx msg")

	// FIXME: correct owner?
	// getAddr := strings.ToLower(hexutils.BytesToHex(res.Ret))
	// suite.Require().Equal(true, strings.HasSuffix(storeAddr, getAddr), "Fail to query the address")
}

func (suite *EvmTestSuite) TestSendTransaction() {
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(0x55ae82600)

	// send simple value transfer with gasLimit=21000
	ethTxParams := &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    1,
		To:       &common.Address{0x1},
		Amount:   big.NewInt(1),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}
	tx := types.NewTx(ethTxParams)
	suite.SignTx(tx)

	result, err := suite.server.EthereumTx(suite.network.GetContext(), tx)
	suite.Require().NoError(err)
	suite.Require().NotNil(result)
}

func (suite *EvmTestSuite) TestOutOfGasWhenDeployContract() {
	// Test contract:
	// http://remix.ethereum.org/#optimize=false&evmVersion=istanbul&version=soljson-v0.5.15+commit.6a57276f.js
	// 2_Owner.sol
	//
	// pragma solidity >=0.4.22 <0.7.0;
	//
	///**
	// * @title Owner
	// * @dev Set & change owner
	// */
	// contract Owner {
	//
	//	address private owner;
	//
	//	// event for EVM logging
	//	event OwnerSet(address indexed oldOwner, address indexed newOwner);
	//
	//	// modifier to check if caller is owner
	//	modifier isOwner() {
	//	// If the first argument of 'require' evaluates to 'false', execution terminates and all
	//	// changes to the state and to Ether balances are reverted.
	//	// This used to consume all gas in old EVM versions, but not anymore.
	//	// It is often a good idea to use 'require' to check if functions are called correctly.
	//	// As a second argument, you can also provide an explanation about what went wrong.
	//	require(msg.sender == owner, "Caller is not owner");
	//	_;
	//}
	//
	//	/**
	//	 * @dev Set contract deployer as owner
	//	 */
	//	constructor() public {
	//	owner = msg.sender; // 'msg.sender' is sender of current call, contract deployer for a constructor
	//	emit OwnerSet(address(0), owner);
	//}
	//
	//	/**
	//	 * @dev Change owner
	//	 * @param newOwner address of new owner
	//	 */
	//	function changeOwner(address newOwner) public isOwner {
	//	emit OwnerSet(owner, newOwner);
	//	owner = newOwner;
	//}
	//
	//	/**
	//	 * @dev Return owner address
	//	 * @return address of owner
	//	 */
	//	function getOwner() external view returns (address) {
	//	return owner;
	//}
	//}

	// Deploy contract - Owner.sol
	gasLimit := uint64(1)
	ctx := suite.network.GetContext().WithGasMeter(storetypes.NewGasMeter(gasLimit))
	gasPrice := big.NewInt(10000)

	bytecode := common.FromHex("0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a36102c4806100dc6000396000f3fe608060405234801561001057600080fd5b5060043610610053576000357c010000000000000000000000000000000000000000000000000000000090048063893d20e814610058578063a6f9dae1146100a2575b600080fd5b6100606100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100e4600480360360208110156100b857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061010f565b005b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146101d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f43616c6c6572206973206e6f74206f776e65720000000000000000000000000081525060200191505060405180910390fd5b8073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a3806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505056fea265627a7a72315820f397f2733a89198bc7fed0764083694c5b828791f39ebcbc9e414bccef14b48064736f6c63430005100032")
	ethTxParams := &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Input:    bytecode,
	}
	tx := types.NewTx(ethTxParams)
	suite.SignTx(tx)

	defer func() {
		//nolint:revive // allow empty code block that just contains TODO in test code
		if r := recover(); r != nil {
			// TODO: snapshotting logic
		} else {
			suite.Require().Fail("panic did not happen")
		}
	}()

	_, err := suite.server.EthereumTx(ctx, tx)
	suite.Require().NoError(err)

	suite.Require().Fail("panic did not happen")
}

func (suite *EvmTestSuite) TestErrorWhenDeployContract() {
	gasLimit := uint64(1000000)
	gasPrice := big.NewInt(10000)

	bytecode := common.FromHex("0xa6f9dae10000000000000000000000006a82e4a67715c8412a9114fbd2cbaefbc8181424")

	ethTxParams := &types.EvmTxArgs{
		ChainID:  suite.network.App.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Input:    bytecode,
	}
	tx := types.NewTx(ethTxParams)
	suite.SignTx(tx)

	res, _ := suite.server.EthereumTx(suite.network.GetContext(), tx)
	suite.Require().Equal("invalid opcode: opcode 0xa6 not defined", res.VmError, "correct evm error")

	// TODO: snapshot checking
}

func (suite *EvmTestSuite) deployERC20Contract() common.Address {
	k := suite.network.App.EvmKeeper
	ctx := suite.network.GetContext()
	from := suite.keyring.GetAddr(0)
	nonce := k.GetNonce(ctx, from)
	ctorArgs, err := types.ERC20Contract.ABI.Pack("", from, big.NewInt(10000000000))
	suite.Require().NoError(err)
	msg := ethtypes.NewMessage(
		from,
		nil,
		nonce,
		big.NewInt(0),
		2000000,
		big.NewInt(1),
		nil,
		nil,
		append(types.ERC20Contract.Bin, ctorArgs...),
		nil,
		true,
	)
	rsp, err := k.ApplyMessage(ctx, msg, nil, true)
	suite.Require().NoError(err)
	suite.Require().False(rsp.Failed())
	return crypto.CreateAddress(from, nonce)
}

// TestERC20TransferReverted checks:
// - when transaction reverted, gas refund works.
// - when transaction reverted, nonce is still increased.
func (suite *EvmTestSuite) TestERC20TransferReverted() {
	intrinsicGas := uint64(21572)
	// test different hooks scenarios
	testCases := []struct {
		msg      string
		gasLimit uint64
		hooks    types.EvmHooks
		expErr   string
	}{
		{
			"no hooks",
			intrinsicGas, // enough for intrinsicGas, but not enough for execution
			nil,
			"out of gas",
		},
		{
			"success hooks",
			intrinsicGas, // enough for intrinsicGas, but not enough for execution
			&DummyHook{},
			"out of gas",
		},
		{
			"failure hooks",
			1000000, // enough gas limit, but hooks fails.
			&FailureHook{},
			"failed to execute post processing",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			suite.SetupTest()
			k := suite.network.App.EvmKeeper.CleanHooks()
			k.SetHooks(tc.hooks)

			ctx := suite.network.GetContext()
			from := suite.keyring.GetAddr(0)

			// add some fund to pay gas fee
			err := k.SetBalance(ctx, from, big.NewInt(1000000000000000))
			suite.Require().NoError(err)

			contract := suite.deployERC20Contract()

			data, err := types.ERC20Contract.ABI.Pack("transfer", from, big.NewInt(10))
			suite.Require().NoError(err)

			gasPrice := big.NewInt(1000000000) // must be bigger than or equal to baseFee
			nonce := k.GetNonce(ctx, from)
			ethTxParams := &types.EvmTxArgs{
				ChainID:  suite.network.App.EvmKeeper.ChainID(),
				Nonce:    nonce,
				To:       &contract,
				Amount:   big.NewInt(0),
				GasPrice: gasPrice,
				GasLimit: tc.gasLimit,
				Input:    data,
			}
			tx := types.NewTx(ethTxParams)
			suite.SignTx(tx)

			before := k.GetBalance(ctx, from)

			evmParams := suite.network.App.EvmKeeper.GetParams(ctx)
			ethCfg := evmParams.GetChainConfig().EthereumConfig(nil)
			baseFee := suite.network.App.EvmKeeper.GetBaseFee(ctx, ethCfg)

			txData, err := types.UnpackTxData(tx.Data)
			suite.Require().NoError(err)
			fees, err := keeper.VerifyFee(txData, types.DefaultEVMDenom, baseFee, true, true, ctx.IsCheckTx())
			suite.Require().NoError(err)
			err = k.DeductTxCostsFromUserBalance(ctx, fees, common.HexToAddress(tx.From))
			suite.Require().NoError(err)

			res, err := k.EthereumTx(ctx, tx)
			suite.Require().NoError(err)

			suite.Require().True(res.Failed())
			suite.Require().Equal(tc.expErr, res.VmError)
			suite.Require().Empty(res.Logs)

			after := k.GetBalance(ctx, from)

			if tc.expErr == "out of gas" {
				suite.Require().Equal(tc.gasLimit, res.GasUsed)
			} else {
				suite.Require().Greater(tc.gasLimit, res.GasUsed)
			}

			// check gas refund works: only deducted fee for gas used, rather than gas limit.
			suite.Require().Equal(new(big.Int).Mul(gasPrice, big.NewInt(int64(res.GasUsed))), new(big.Int).Sub(before, after))

			// nonce should not be increased.
			nonce2 := k.GetNonce(ctx, from)
			suite.Require().Equal(nonce, nonce2)
		})
	}
}

func (suite *EvmTestSuite) TestContractDeploymentRevert() {
	intrinsicGas := uint64(134180)
	testCases := []struct {
		msg      string
		gasLimit uint64
		hooks    types.EvmHooks
	}{
		{
			"no hooks",
			intrinsicGas,
			nil,
		},
		{
			"success hooks",
			intrinsicGas,
			&DummyHook{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			suite.SetupTest()
			k := suite.network.App.EvmKeeper.CleanHooks()
			ctx := suite.network.GetContext()
			from := suite.keyring.GetAddr(0)

			// test with different hooks scenarios
			k.SetHooks(tc.hooks)

			nonce := k.GetNonce(ctx, from)
			ctorArgs, err := types.ERC20Contract.ABI.Pack("", from, big.NewInt(0))
			suite.Require().NoError(err)

			ethTxParams := &types.EvmTxArgs{
				Nonce:    nonce,
				GasLimit: tc.gasLimit,
				Input:    append(types.ERC20Contract.Bin, ctorArgs...),
			}
			tx := types.NewTx(ethTxParams)
			suite.SignTx(tx)

			// simulate nonce increment in ante handler
			db := suite.StateDB()
			db.SetNonce(from, nonce+1)
			suite.Require().NoError(db.Commit())

			rsp, err := k.EthereumTx(ctx, tx)
			suite.Require().NoError(err)
			suite.Require().True(rsp.Failed())

			// nonce don't change
			nonce2 := k.GetNonce(ctx, from)
			suite.Require().Equal(nonce+1, nonce2)
		})
	}
}

// DummyHook implements EvmHooks interface
type DummyHook struct{}

func (dh *DummyHook) PostTxProcessing(_ sdk.Context, _ core.Message, _ *ethtypes.Receipt) error {
	return nil
}

// FailureHook implements EvmHooks interface
type FailureHook struct{}

func (dh *FailureHook) PostTxProcessing(_ sdk.Context, _ core.Message, _ *ethtypes.Receipt) error {
	return errors.New("mock error")
}
