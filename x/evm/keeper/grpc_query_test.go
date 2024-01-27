package keeper_test

import (
	"encoding/json"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	ethlogger "github.com/ethereum/go-ethereum/eth/tracers/logger"
	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/evmos/evmos/v16/server/config"
	// "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	// "github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/evm/statedb"
	"github.com/evmos/evmos/v16/x/evm/types"
)

// Not valid Ethereum address
const invalidAddress = "0x0000"

// expGasConsumed is the gas consumed in traceTx setup (GetProposerAddr + CalculateBaseFee)
const expGasConsumed = 8781

// expGasConsumedWithFeeMkt is the gas consumed in traceTx setup (GetProposerAddr + CalculateBaseFee) with enabled feemarket
const expGasConsumedWithFeeMkt = 8775

func (suite *EvmKeeperTestSuite) TestQueryAccount() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		msg               string
		getReq            func() *types.QueryAccountRequest
		exptectedResponse *types.QueryAccountResponse
		expPass           bool
	}{
		{
			"invalid address",
			func() *types.QueryAccountRequest {
				return &types.QueryAccountRequest{
					Address: invalidAddress,
				}
			},
			nil,
			false,
		},
		{
			"success",
			func() *types.QueryAccountRequest {
				amt := sdk.Coins{sdk.NewInt64Coin(types.DefaultEVMDenom, 100)}

				// Add new unfunded key
				index := keyring.AddKey()
				addr := keyring.GetAddr(index)

				err := unitNetwork.App.BankKeeper.MintCoins(
					unitNetwork.GetContext(),
					types.ModuleName,
					amt,
				)
				suite.Require().NoError(err)

				err = unitNetwork.App.BankKeeper.SendCoinsFromModuleToAccount(
					unitNetwork.GetContext(),
					types.ModuleName,
					addr.Bytes(),
					amt,
				)
				suite.Require().NoError(err)

				return &types.QueryAccountRequest{
					Address: addr.String(),
				}
			},
			&types.QueryAccountResponse{
				Balance:  "100",
				CodeHash: common.BytesToHash(crypto.Keccak256(nil)).Hex(),
				Nonce:    0,
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			req := tc.getReq()
			expectedResponse := tc.exptectedResponse

			ctx := unitNetwork.GetContext()
			// Function under test
			res, err := unitNetwork.GetEvmClient().Account(ctx, req)

			suite.Require().Equal(expectedResponse, res)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *EvmKeeperTestSuite) TestQueryCosmosAccount() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		msg           string
		getReqAndResp func() (*types.QueryCosmosAccountRequest, *types.QueryCosmosAccountResponse)
		expPass       bool
	}{
		{
			"invalid address",
			func() (*types.QueryCosmosAccountRequest, *types.QueryCosmosAccountResponse) {
				req := &types.QueryCosmosAccountRequest{
					Address: invalidAddress,
				}
				return req, nil
			},
			false,
		},
		{
			"success",
			func() (*types.QueryCosmosAccountRequest, *types.QueryCosmosAccountResponse) {
				key := keyring.GetKey(0)
				expAccount := &types.QueryCosmosAccountResponse{
					CosmosAddress: key.AccAddr.String(),
					Sequence:      0,
					AccountNumber: 0,
				}
				req := &types.QueryCosmosAccountRequest{
					Address: key.Addr.String(),
				}

				return req, expAccount
			},
			true,
		},
		{
			"success with seq and account number",
			func() (*types.QueryCosmosAccountRequest, *types.QueryCosmosAccountResponse) {
				index := keyring.AddKey()
				newKey := keyring.GetKey(index)
				accountNumber := uint64(100)
				acc := unitNetwork.App.AccountKeeper.NewAccountWithAddress(
					unitNetwork.GetContext(),
					newKey.AccAddr,
				)

				suite.Require().NoError(acc.SetSequence(10))
				suite.Require().NoError(acc.SetAccountNumber(accountNumber))
				unitNetwork.App.AccountKeeper.SetAccount(unitNetwork.GetContext(), acc)

				expAccount := &types.QueryCosmosAccountResponse{
					CosmosAddress: newKey.AccAddr.String(),
					Sequence:      10,
					AccountNumber: accountNumber,
				}

				req := &types.QueryCosmosAccountRequest{
					Address: newKey.Addr.String(),
				}
				return req, expAccount
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			req, expectedResponse := tc.getReqAndResp()

			ctx := unitNetwork.GetContext()

			// Function under test
			res, err := unitNetwork.GetEvmClient().CosmosAccount(ctx, req)

			suite.Require().Equal(expectedResponse, res)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *EvmKeeperTestSuite) TestQueryBalance() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		msg           string
		getReqAndResp func() (*types.QueryBalanceRequest, *types.QueryBalanceResponse)
		expPass       bool
	}{
		{
			"invalid address",
			func() (*types.QueryBalanceRequest, *types.QueryBalanceResponse) {
				req := &types.QueryBalanceRequest{
					Address: invalidAddress,
				}
				return req, nil
			},
			false,
		},
		{
			"success",
			func() (*types.QueryBalanceRequest, *types.QueryBalanceResponse) {
				newIndex := keyring.AddKey()
				addr := keyring.GetAddr(newIndex)

				balance := int64(100)
				amt := sdk.Coins{sdk.NewInt64Coin(types.DefaultEVMDenom, balance)}

				err := unitNetwork.App.BankKeeper.MintCoins(unitNetwork.GetContext(), types.ModuleName, amt)
				suite.Require().NoError(err)
				err = unitNetwork.App.BankKeeper.SendCoinsFromModuleToAccount(unitNetwork.GetContext(), types.ModuleName, addr.Bytes(), amt)
				suite.Require().NoError(err)

				req := &types.QueryBalanceRequest{
					Address: addr.String(),
				}
				return req, &types.QueryBalanceResponse{
					Balance: fmt.Sprint(balance),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			req, resp := tc.getReqAndResp()

			ctx := unitNetwork.GetContext()
			res, err := unitNetwork.GetEvmClient().Balance(ctx, req)

			suite.Require().Equal(resp, res)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *EvmKeeperTestSuite) TestQueryStorage() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		msg           string
		getReqAndResp func() (*types.QueryStorageRequest, *types.QueryStorageResponse)
		expPass       bool
	}{
		{
			"invalid address",
			func() (*types.QueryStorageRequest, *types.QueryStorageResponse) {
				req := &types.QueryStorageRequest{
					Address: invalidAddress,
				}
				return req, nil
			},
			false,
		},
		{
			"success",
			func() (*types.QueryStorageRequest, *types.QueryStorageResponse) {
				key := common.BytesToHash([]byte("key"))
				value := []byte("value")
				expValue := common.BytesToHash(value)

				newIndex := keyring.AddKey()
				addr := keyring.GetAddr(newIndex)

				unitNetwork.App.EvmKeeper.SetState(
					unitNetwork.GetContext(),
					addr,
					key,
					value,
				)

				req := &types.QueryStorageRequest{
					Address: addr.String(),
					Key:     key.String(),
				}
				return req, &types.QueryStorageResponse{
					Value: expValue.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			req, expectedResp := tc.getReqAndResp()

			ctx := unitNetwork.GetContext()
			res, err := unitNetwork.GetEvmClient().Storage(ctx, req)

			suite.Require().Equal(expectedResp, res)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *EvmKeeperTestSuite) TestQueryCode() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	var (
		req     *types.QueryCodeRequest
		expCode []byte
	)

	testCases := []struct {
		msg           string
		getReqAndResp func() (*types.QueryCodeRequest, *types.QueryCodeResponse)
		expPass       bool
	}{
		{
			"invalid address",
			func() (*types.QueryCodeRequest, *types.QueryCodeResponse) {
				req = &types.QueryCodeRequest{
					Address: invalidAddress,
				}
				return req, nil
			},
			false,
		},
		{
			"success",
			func() (*types.QueryCodeRequest, *types.QueryCodeResponse) {
				newIndex := keyring.AddKey()
				addr := keyring.GetAddr(newIndex)

				expCode = []byte("code")
				stateDbB := unitNetwork.GetStateDB()
				stateDbB.SetCode(addr, expCode)
				suite.Require().NoError(stateDbB.Commit())

				req = &types.QueryCodeRequest{
					Address: addr.String(),
				}
				return req, &types.QueryCodeResponse{
					Code: hexutil.Bytes(expCode),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			req, expectedResponse := tc.getReqAndResp()

			ctx := unitNetwork.GetContext()
			res, err := unitNetwork.GetEvmClient().Code(ctx, req)

			suite.Require().Equal(expectedResponse, res)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TODO: Fix this one
func (suite *KeeperTestSuite) TestQueryTxLogs() {
	var expLogs []*types.Log
	txHash := common.BytesToHash([]byte("tx_hash"))
	txIndex := uint(1)
	logIndex := uint(1)

	testCases := []struct {
		msg      string
		malleate func(vm.StateDB)
	}{
		{
			"empty logs",
			func(vm.StateDB) {
				expLogs = nil
			},
		},
		{
			"success",
			func(vmdb vm.StateDB) {
				addr := suite.keyring.GetAddr(0)
				expLogs = []*types.Log{
					{
						Address:     addr.String(),
						Topics:      []string{common.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      txHash.String(),
						TxIndex:     uint64(txIndex),
						BlockHash:   common.BytesToHash(suite.network.GetContext().HeaderHash()).Hex(),
						Index:       uint64(logIndex),
						Removed:     false,
					},
				}

				for _, log := range types.LogsToEthereum(expLogs) {
					vmdb.AddLog(log)
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			vmdb := statedb.New(suite.network.GetContext(), suite.network.App.EvmKeeper, statedb.NewTxConfig(common.BytesToHash(suite.network.GetContext().HeaderHash()), txHash, txIndex, logIndex))
			tc.malleate(vmdb)
			suite.Require().NoError(vmdb.Commit())

			logs := vmdb.Logs()
			suite.Require().Equal(expLogs, types.NewLogsFromEth(logs))
		})
	}
}

func (suite *EvmKeeperTestSuite) TestQueryParams() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	ctx := unitNetwork.GetContext()
	expParams := types.DefaultParams()

	res, err := unitNetwork.GetEvmClient().Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

// TODO: fix this issue
func (suite *EvmKeeperTestSuite) TestQueryValidatorAccount() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	var (
		req        *types.QueryValidatorAccountRequest
		expAccount *types.QueryValidatorAccountResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid address",
			func() {
				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(common.Address{}.Bytes()).String(),
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: "",
				}
			},
			false,
		},
		{
			"success",
			func() {
				val := unitNetwork.GetValidators()[0]
				consAddr, err := val.GetConsAddr()
				suite.Require().NoError(err)

				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: val.OperatorAddress,
					Sequence:       0,
					AccountNumber:  0,
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: sdk.ConsAddress(consAddr).String(),
				}
			},
			true,
		},
		{
			"success with seq and account number",
			func() {
				accNumber := uint64(100)
				accSeq := uint64(10)

				val := unitNetwork.GetValidators()[0]
				consAddr, err := val.GetConsAddr()
				suite.Require().NoError(err)

				acc := unitNetwork.App.AccountKeeper.GetAccount(
					unitNetwork.GetContext(),
					sdk.AccAddress(val.OperatorAddress).Bytes(),
				)
				suite.Require().NoError(acc.SetSequence(accSeq))
				suite.Require().NoError(acc.SetAccountNumber(accNumber))

				// Function under test
				unitNetwork.App.AccountKeeper.SetAccount(unitNetwork.GetContext(), acc)

				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: val.OperatorAddress,
					Sequence:       accSeq,
					AccountNumber:  accNumber,
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: sdk.ConsAddress(consAddr).String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			ctx := unitNetwork.GetContext()
			res, err := unitNetwork.GetEvmClient().ValidatorAccount(ctx, req)

			suite.Require().Equal(expAccount, res)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *EvmKeeperTestSuite) TestEstimateGas() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grcpHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grcpHandler)

	gasHelper := hexutil.Uint64(20000)
	higherGas := hexutil.Uint64(25000)
	hexBigInt := hexutil.Big(*big.NewInt(1))

	testCases := []struct {
		msg             string
		getArgs         func() types.TransactionArgs
		expPass         bool
		expGas          uint64
		enableFeemarket bool
		gasCap          uint64
	}{
		// should success, because transfer value is zero
		{
			"success - default args - special case for ErrIntrinsicGas on contract creation, raise gas limit",
			func() types.TransactionArgs {
				return types.TransactionArgs{}
			},
			true,
			ethparams.TxGasContractCreation,
			false,
			config.DefaultGasCap,
		},
		// should success, because transfer value is zero
		{
			"success - default args with 'to' address",
			func() types.TransactionArgs {
				return types.TransactionArgs{To: &common.Address{}}
			},
			true,
			ethparams.TxGas,
			false,
			config.DefaultGasCap,
		},
		// should fail, because the default From address(zero address) don't have fund
		{
			"fail - not enough balance",
			func() types.TransactionArgs {
				return types.TransactionArgs{
					To:    &common.Address{},
					Value: (*hexutil.Big)(big.NewInt(100)),
				}
			},
			false,
			0,
			false,
			config.DefaultGasCap,
		},
		// should success, enough balance now
		{
			"success - enough balance",
			func() types.TransactionArgs {
				addr := keyring.GetAddr(0)
				return types.TransactionArgs{
					To:    &common.Address{},
					From:  &addr,
					Value: (*hexutil.Big)(big.NewInt(100)),
				}
			},
			true,
			ethparams.TxGas,
			false,
			config.DefaultGasCap,
		},
		// should success, because gas limit lower than 21000 is ignored
		{
			"gas exceed allowance",
			func() types.TransactionArgs {
				return types.TransactionArgs{To: &common.Address{}, Gas: &gasHelper}
			},
			true,
			ethparams.TxGas,
			false,
			config.DefaultGasCap,
		},
		// should fail, invalid gas cap
		{
			"gas exceed global allowance",
			func() types.TransactionArgs {
				return types.TransactionArgs{To: &common.Address{}}
			},
			false,
			0,
			false,
			20000,
		},
		// estimate gas of an erc20 contract deployment, the exact gas number is checked with geth
		{
			"contract deployment",
			func() types.TransactionArgs {
				addr := keyring.GetAddr(0)
				ctorArgs, err := types.ERC20Contract.ABI.Pack(
					"",
					&addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				)
				suite.Require().NoError(err)

				data := types.ERC20Contract.Bin
				data = append(data, ctorArgs...)
				return types.TransactionArgs{
					From: &addr,
					Data: (*hexutil.Bytes)(&data),
				}
			},
			true,
			1186778,
			false,
			config.DefaultGasCap,
		},
		// estimate gas of an erc20 transfer, the exact gas number is checked with geth
		{
			"erc20 transfer",
			func() types.TransactionArgs {
				key := keyring.GetKey(0)
				constructorArgs := []interface{}{
					key.Addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				}
				compiledContract := types.ERC20Contract

				contractAddr, err := txFactory.DeployContract(
					key.Priv,
					types.EvmTxArgs{}, // Default values
					factory.ContractDeploymentData{
						Contract:        compiledContract,
						ConstructorArgs: constructorArgs,
					},
				)
				suite.Require().NoError(err)

				err = unitNetwork.NextBlock()
				suite.Require().NoError(err)

				newIndex := keyring.AddKey()
				recipient := keyring.GetAddr(newIndex)
				transferData, err := types.ERC20Contract.ABI.Pack(
					"transfer",
					recipient,
					big.NewInt(1000),
				)
				suite.Require().NoError(err)
				return types.TransactionArgs{
					To:   &contractAddr,
					From: &key.Addr,
					Data: (*hexutil.Bytes)(&transferData),
				}
			},
			true,
			51880,
			false,
			config.DefaultGasCap,
		},
		// repeated tests with enableFeemarket
		{
			"default args w/ enableFeemarket",
			func() types.TransactionArgs {
				return types.TransactionArgs{To: &common.Address{}}
			},
			true,
			ethparams.TxGas,
			true,
			config.DefaultGasCap,
		},
		{
			"not enough balance w/ enableFeemarket",
			func() types.TransactionArgs {
				return types.TransactionArgs{
					To:    &common.Address{},
					Value: (*hexutil.Big)(big.NewInt(100)),
				}
			},
			false,
			0,
			true,
			config.DefaultGasCap,
		},
		{
			"enough balance w/ enableFeemarket",
			func() types.TransactionArgs {
				addr := keyring.GetAddr(0)
				return types.TransactionArgs{
					To:    &common.Address{},
					From:  &addr,
					Value: (*hexutil.Big)(big.NewInt(100)),
				}
			},
			true,
			ethparams.TxGas,
			true,
			config.DefaultGasCap,
		},
		{
			"gas exceed allowance w/ enableFeemarket",
			func() types.TransactionArgs {
				return types.TransactionArgs{To: &common.Address{}, Gas: &gasHelper}
			},
			true,
			ethparams.TxGas,
			true,
			config.DefaultGasCap,
		},
		{
			"gas exceed global allowance w/ enableFeemarket",
			func() types.TransactionArgs {
				return types.TransactionArgs{To: &common.Address{}}
			},
			false,
			0,
			true,
			20000,
		},
		{
			"contract deployment w/ enableFeemarket",
			func() types.TransactionArgs {
				addr := keyring.GetAddr(0)
				ctorArgs, err := types.ERC20Contract.ABI.Pack("", &addr, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Require().NoError(err)
				data := types.ERC20Contract.Bin
				data = append(data, ctorArgs...)
				return types.TransactionArgs{
					From: &addr,
					Data: (*hexutil.Bytes)(&data),
				}
			},
			true,
			1186778,
			true,
			config.DefaultGasCap,
		},
		{
			"erc20 transfer w/ enableFeemarket",
			func() types.TransactionArgs {
				key := keyring.GetKey(0)
				constructorArgs := []interface{}{
					key.Addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				}
				compiledContract := types.ERC20Contract

				contractAddr, err := txFactory.DeployContract(
					key.Priv,
					types.EvmTxArgs{}, // Default values
					factory.ContractDeploymentData{
						Contract:        compiledContract,
						ConstructorArgs: constructorArgs,
					},
				)
				suite.Require().NoError(err)

				err = unitNetwork.NextBlock()
				suite.Require().NoError(err)

				newIndex := keyring.AddKey()
				recipient := keyring.GetAddr(newIndex)
				transferData, err := types.ERC20Contract.ABI.Pack(
					"transfer",
					recipient,
					big.NewInt(1000),
				)
				suite.Require().NoError(err)

				return types.TransactionArgs{
					To:   &contractAddr,
					From: &key.Addr,
					Data: (*hexutil.Bytes)(&transferData),
				}
			},
			true,
			51880,
			true,
			config.DefaultGasCap,
		},
		{
			"contract creation but 'create' param disabled",
			func() types.TransactionArgs {
				addr := keyring.GetAddr(0)
				ctorArgs, err := types.ERC20Contract.ABI.Pack(
					"",
					&addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				)
				suite.Require().NoError(err)

				data := types.ERC20Contract.Bin
				data = append(data, ctorArgs...)

				args := types.TransactionArgs{
					From: &addr,
					Data: (*hexutil.Bytes)(&data),
				}
				params := unitNetwork.App.EvmKeeper.GetParams(unitNetwork.GetContext())
				params.EnableCreate = false
				err = unitNetwork.App.EvmKeeper.SetParams(
					unitNetwork.GetContext(),
					params,
				)
				suite.Require().NoError(err)

				return args
			},
			false,
			0,
			false,
			config.DefaultGasCap,
		},
		{
			"specified gas in args higher than ethparams.TxGas (21,000)",
			func() types.TransactionArgs {
				return types.TransactionArgs{
					To:  &common.Address{},
					Gas: &higherGas,
				}
			},
			true,
			ethparams.TxGas,
			false,
			config.DefaultGasCap,
		},
		{
			"specified gas in args higher than request gasCap",
			func() types.TransactionArgs {
				return types.TransactionArgs{
					To:  &common.Address{},
					Gas: &higherGas,
				}
			},
			true,
			ethparams.TxGas,
			false,
			22_000,
		},
		{
			"invalid args - specified both gasPrice and maxFeePerGas",
			func() types.TransactionArgs {
				return types.TransactionArgs{
					To:           &common.Address{},
					GasPrice:     &hexBigInt,
					MaxFeePerGas: &hexBigInt,
				}
			},
			false,
			0,
			false,
			config.DefaultGasCap,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			if tc.enableFeemarket {
				evmParams := unitNetwork.App.FeeMarketKeeper.GetParams(
					unitNetwork.GetContext(),
				)
				evmParams.NoBaseFee = false

				err := unitNetwork.App.FeeMarketKeeper.SetParams(
					unitNetwork.GetContext(),
					evmParams,
				)
				suite.Require().NoError(err)
			}

			args := tc.getArgs()

			marshalArgs, err := json.Marshal(args)
			suite.Require().NoError(err)
			req := types.EthCallRequest{
				Args:            marshalArgs,
				GasCap:          tc.gasCap,
				ProposerAddress: unitNetwork.GetContext().BlockHeader().ProposerAddress,
			}

			rsp, err := unitNetwork.GetEvmClient().EstimateGas(
				unitNetwork.GetContext(),
				&req,
			)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(int64(tc.expGas), int64(rsp.Gas))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func getDefaultTraceTxRequest(unitNetwork network.Network) types.QueryTraceTxRequest {
	ctx := unitNetwork.GetContext()
	chainID := unitNetwork.GetEIP155ChainID().Int64()
	return types.QueryTraceTxRequest{
		BlockMaxGas: ctx.ConsensusParams().Block.MaxGas,
		ChainId:     chainID,
		BlockTime:   ctx.BlockTime(),
		TraceConfig: &types.TraceConfig{},
	}
}

func (suite *EvmKeeperTestSuite) TestTraceTx() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grcpHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grcpHandler)

	testCases := []struct {
		msg             string
		malleate        func()
		getRequest      func() types.QueryTraceTxRequest
		getPredecessors func() []*types.MsgEthereumTx
		expPass         bool
		expectedTrace   string
	}{
		{
			msg: "default trace",
			getRequest: func() types.QueryTraceTxRequest {
				return getDefaultTraceTxRequest(unitNetwork)
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass:       true,
			expectedTrace: "{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "default trace with filtered response",
			getRequest: func() types.QueryTraceTxRequest {
				defaultRequest := getDefaultTraceTxRequest(unitNetwork)
				defaultRequest.TraceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
				return defaultRequest
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass:       true,
			expectedTrace: "{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "javascript tracer",
			getRequest: func() types.QueryTraceTxRequest {
				traceConfig := &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				defaultRequest := getDefaultTraceTxRequest(unitNetwork)
				defaultRequest.TraceConfig = traceConfig
				return defaultRequest
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass:       true,
			expectedTrace: "[]",
		},
		{
			msg: "default tracer with predecessors",
			getRequest: func() types.QueryTraceTxRequest {
				return getDefaultTraceTxRequest(unitNetwork)
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				// Create predecessor tx
				// Use different address to avoid nonce collision
				senderKey := keyring.GetKey(1)

				constructorArgs := []interface{}{
					senderKey.Addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				}
				compiledContract := types.ERC20Contract
				contractAddr, err := txFactory.DeployContract(
					senderKey.Priv,
					types.EvmTxArgs{}, // Default values
					factory.ContractDeploymentData{
						Contract:        compiledContract,
						ConstructorArgs: constructorArgs,
					},
				)
				suite.Require().NoError(err)

				err = unitNetwork.NextBlock()
				suite.Require().NoError(err)

				recipientIndex := keyring.AddKey()
				recipientAddr := keyring.GetAddr(recipientIndex)

				transferArgs := types.EvmTxArgs{
					To: &contractAddr,
				}
				callArgs := factory.CallArgs{
					ContractABI: compiledContract.ABI,
					MethodName:  "transfer",
					Args:        []interface{}{recipientAddr, big.NewInt(1000)},
				}

				transferArgs, err = txFactory.GenerateContractCallArgs(transferArgs, callArgs)
				suite.Require().NoError(err)

				// Generate the message to add to predecessors
				firstTx, err := txFactory.GenerateMsgEthereumTx(senderKey.Priv, transferArgs)
				suite.Require().NoError(err)

				result, err := txFactory.ExecuteContractCall(senderKey.Priv, transferArgs, callArgs)
				suite.Require().NoError(err)
				suite.Require().True(result.IsOK())

				return []*types.MsgEthereumTx{&firstTx}
			},
			expPass:       true,
			expectedTrace: "{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "invalid trace config - Negative Limit",
			getRequest: func() types.QueryTraceTxRequest {
				defaultRequest := getDefaultTraceTxRequest(unitNetwork)
				defaultRequest.TraceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Limit:          -1,
				}
				return defaultRequest
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass: false,
		},
		{
			msg: "invalid trace config - Invalid Tracer",
			getRequest: func() types.QueryTraceTxRequest {
				defaultRequest := getDefaultTraceTxRequest(unitNetwork)
				defaultRequest.TraceConfig = &types.TraceConfig{
					Tracer: "invalid_tracer",
				}
				return defaultRequest
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass: false,
		},
		{
			msg: "invalid trace config - Invalid Timeout",
			getRequest: func() types.QueryTraceTxRequest {
				defaultRequest := getDefaultTraceTxRequest(unitNetwork)
				defaultRequest.TraceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Timeout:        "wrong_time",
				}
				return defaultRequest
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass: false,
		},
		{
			msg: "default tracer with contract creation tx as predecessor but 'create' param disabled",
			getRequest: func() types.QueryTraceTxRequest {
				return getDefaultTraceTxRequest(unitNetwork)
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				// use different address to avoid nonce collision
				senderKey := keyring.GetKey(1)

				constructorArgs := []interface{}{
					senderKey.Addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				}
				compiledContract := types.ERC20Contract
				deploymentData := factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				}

				txArgs, err := txFactory.GenerateDeployContractArgs(senderKey.Addr, types.EvmTxArgs{}, deploymentData)
				suite.Require().NoError(err)

				txMsg, err := txFactory.GenerateMsgEthereumTx(senderKey.Priv, txArgs)
				suite.Require().NoError(err)

				_, err = txFactory.ExecuteEthTx(
					senderKey.Priv,
					txArgs, // Default values
				)
				suite.Require().NoError(err)

				// Disable create param
				params := unitNetwork.App.EvmKeeper.GetParams(unitNetwork.GetContext())
				params.EnableCreate = false
				err = unitNetwork.App.EvmKeeper.SetParams(unitNetwork.GetContext(), params)
				suite.Require().NoError(err)
				return []*types.MsgEthereumTx{&txMsg}
			},
			expPass:       true,
			expectedTrace: "{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "invalid chain id",
			getRequest: func() types.QueryTraceTxRequest {
				defaultRequest := getDefaultTraceTxRequest(unitNetwork)
				defaultRequest.ChainId = -1
				return defaultRequest
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			// ----- Contract Deployment -----
			key := keyring.GetKey(0)
			constructorArgs := []interface{}{
				key.Addr,
				sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
			}
			compiledContract := types.ERC20Contract
			contractAddr, err := txFactory.DeployContract(
				key.Priv,
				types.EvmTxArgs{}, // Default values
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
			)
			suite.Require().NoError(err)

			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)

			// --- Contract Call ---
			newIndex := keyring.AddKey()
			recipient := keyring.GetAddr(newIndex)

			transferArgs := types.EvmTxArgs{
				To: &contractAddr,
			}
			callArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "transfer",
				Args:        []interface{}{recipient, big.NewInt(1000)},
			}

			// --- Add predecessor ---
			predecessors := tc.getPredecessors()

			// Create the transaction to be traced
			// Note - it is important this is on the same block as the predecessor
			input, err := callArgs.ContractABI.Pack(callArgs.MethodName, callArgs.Args...)
			suite.Require().NoError(err)
			transferArgs.Input = input

			msgToTrace, err := txFactory.GenerateMsgEthereumTx(key.Priv, transferArgs)
			suite.Require().NoError(err)

			signer := ethtypes.LatestSignerForChainID(unitNetwork.GetEIP155ChainID())
			err = msgToTrace.Sign(signer, utiltx.NewSigner(key.Priv))
			suite.Require().NoError(err)

			txRes, err := txFactory.ExecuteContractCall(key.Priv, transferArgs, callArgs)
			suite.Require().NoError(err)
			suite.Require().True(txRes.IsOK())

			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)

			// Get the trace request
			traceReq := tc.getRequest()
			// Add predecessor to trace request
			traceReq.Predecessors = predecessors
			traceReq.Msg = &msgToTrace

			// Function under test
			res, err := unitNetwork.GetEvmClient().TraceTx(
				unitNetwork.GetContext(),
				&traceReq,
			)

			if tc.expPass {
				suite.Require().NoError(err)

				// if data is to big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.expectedTrace, string(res.Data[:150]))
				} else {
					suite.Require().Equal(tc.expectedTrace, string(res.Data))
				}
				if traceReq.TraceConfig == nil || traceReq.TraceConfig.Tracer == "" {
					var result ethlogger.ExecutionResult
					suite.Require().NoError(json.Unmarshal(res.Data, &result))
					suite.Require().Positive(result.Gas)
				}
			} else {
				suite.Require().Error(err)
			}

			// Clean up per test
			defaultEvmParams := types.DefaultParams()
			err = unitNetwork.App.EvmKeeper.SetParams(unitNetwork.GetContext(), defaultEvmParams)
			suite.Require().NoError(err)

			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}

func getDefaultTraceBlockRequest(unitNetwork network.Network) types.QueryTraceBlockRequest {
	ctx := unitNetwork.GetContext()
	chainID := unitNetwork.GetEIP155ChainID().Int64()
	return types.QueryTraceBlockRequest{
		BlockMaxGas: ctx.ConsensusParams().Block.MaxGas,
		ChainId:     chainID,
		BlockTime:   ctx.BlockTime(),
	}
}

func (suite *EvmKeeperTestSuite) TestTraceBlock() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grcpHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grcpHandler)

	testCases := []struct {
		msg              string
		getRequest       func() types.QueryTraceBlockRequest
		getAdditionalTxs func() []*types.MsgEthereumTx
		malleate         func()
		expPass          bool
		traceResponse    string

		enableFeemarket bool
		expFinalGas     uint64
	}{
		{
			msg: "default trace",
			getRequest: func() types.QueryTraceBlockRequest {
				return getDefaultTraceBlockRequest(unitNetwork)
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			// malleate: func() {
			// 	traceConfig = nil
			// },
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
			expFinalGas:   expGasConsumed,
		},
		{
			msg: "filtered trace",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(unitNetwork)
				defaultReq.TraceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			// malleate: func() {
			// 	traceConfig = &types.TraceConfig{
			// 		DisableStack:   true,
			// 		DisableStorage: true,
			// 		EnableMemory:   false,
			// 	}
			// },
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
			expFinalGas:   expGasConsumed,
		},
		{
			msg: "javascript tracer",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(unitNetwork)
				defaultReq.TraceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			// malleate: func() {
			// 	traceConfig = &types.TraceConfig{
			// 		Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
			// 	}
			// },
			expPass:       true,
			traceResponse: "[{\"result\":[]}]",
			expFinalGas:   expGasConsumed,
		},
		// {
		// 	msg: "default trace with enableFeemarket and filtered return",
		//           getRequest: func() types.QueryTraceBlockRequest {
		//             defaultReq := getDefaultTraceBlockRequest(unitNetwork)
		//             defaultReq.TraceConfig = &types.TraceConfig{
		//               DisableStack:   true,
		//
		// 	malleate: func() {
		// 		traceConfig = &types.TraceConfig{
		// 			DisableStack:   true,
		// 			DisableStorage: true,
		// 			EnableMemory:   false,
		// 		}
		// 	},
		// 	expPass:         true,
		// 	traceResponse:   "[{\"result\":{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
		// 	enableFeemarket: true,
		// 	expFinalGas:     expGasConsumedWithFeeMkt,
		// },
		// {
		// 	msg: "javascript tracer with enableFeemarket",
		// 	malleate: func() {
		// 		traceConfig = &types.TraceConfig{
		// 			Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
		// 		}
		// 	},
		// 	expPass:         true,
		// 	traceResponse:   "[{\"result\":[]}]",
		// 	enableFeemarket: true,
		// 	expFinalGas:     expGasConsumedWithFeeMkt,
		// },
		{
			msg: "tracer with multiple transactions",
			getRequest: func() types.QueryTraceBlockRequest {
				return getDefaultTraceBlockRequest(unitNetwork)
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				// Create predecessor tx
				// Use different address to avoid nonce collision
				senderKey := keyring.GetKey(1)
				constructorArgs := []interface{}{
					senderKey.Addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				}
				compiledContract := types.ERC20Contract
				contractAddr, err := txFactory.DeployContract(
					senderKey.Priv,
					types.EvmTxArgs{}, // Default values
					factory.ContractDeploymentData{
						Contract:        compiledContract,
						ConstructorArgs: constructorArgs,
					},
				)
				suite.Require().NoError(err)

				err = unitNetwork.NextBlock()
				suite.Require().NoError(err)

				recipientIndex := keyring.AddKey()
				recipientAddr := keyring.GetAddr(recipientIndex)

				transferArgs := types.EvmTxArgs{
					To: &contractAddr,
				}
				callArgs := factory.CallArgs{
					ContractABI: compiledContract.ABI,
					MethodName:  "transfer",
					Args:        []interface{}{recipientAddr, big.NewInt(1000)},
				}
				transferArgs, err = txFactory.GenerateContractCallArgs(transferArgs, callArgs)
				suite.Require().NoError(err)

				// We need to get access to the message
				firstSignedTX, err := txFactory.GenerateSignedEthTx(senderKey.Priv, transferArgs)
				suite.Require().NoError(err)
				firstTx := firstSignedTX.GetMsgs()[0].(*types.MsgEthereumTx)

				fmt.Println("Height", unitNetwork.GetContext().BlockHeight())

				result, err := txFactory.ExecuteContractCall(senderKey.Priv, transferArgs, callArgs)
				suite.Require().NoError(err)
				suite.Require().True(result.IsOK())

				secondArgs, err := txFactory.GenerateContractCallArgs(transferArgs, callArgs)
				secondArgs.Nonce = firstTx.AsTransaction().Nonce() + 1
				secondSignedTx, err := txFactory.GenerateSignedEthTx(senderKey.Priv, secondArgs)
				suite.Require().NoError(err)
				secondTx := secondSignedTx.GetMsgs()[0].(*types.MsgEthereumTx)

				secondResult, err := txFactory.ExecuteContractCall(senderKey.Priv, secondArgs, callArgs)
				suite.Require().NoError(err)
				suite.Require().True(secondResult.IsOK())

				suite.Require().NoError(unitNetwork.NextBlock())
				return []*types.MsgEthereumTx{firstTx, secondTx}
			},
			// malleate: func() {
			// 	traceConfig = nil
			//
			// 	// increase nonce to avoid address collision
			// 	vmdb := suite.StateDB()
			// 	addr := suite.keyring.GetAddr(0)
			// 	vmdb.SetNonce(addr, vmdb.GetNonce(addr)+1)
			// 	suite.Require().NoError(vmdb.Commit())
			//
			// 	contractAddr := suite.DeployTestContract(suite.T(), addr, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
			// 	err := suite.network.NextBlock()
			// 	suite.Require().NoError(err)
			// 	// create multiple transactions in the same block
			// 	firstTx := suite.TransferERC20Token(suite.T(), contractAddr, addr, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
			// 	secondTx := suite.TransferERC20Token(suite.T(), contractAddr, addr, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
			// 	err = suite.network.NextBlock()
			// 	suite.Require().NoError(err)
			// 	// overwrite txs to include only the ones on new block
			// 	txs = append([]*types.MsgEthereumTx{}, firstTx, secondTx)
			// },
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",

			enableFeemarket: false,
			expFinalGas:     expGasConsumed,
		},
		{
			msg: "invalid trace config - Negative Limit",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(unitNetwork)
				defaultReq.TraceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Limit:          -1,
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			// malleate: func() {
			// 	traceConfig = &types.TraceConfig{
			// 		DisableStack:   true,
			// 		DisableStorage: true,
			// 		EnableMemory:   false,
			// 		Limit:          -1,
			// 	}
			// },
			expPass:     false,
			expFinalGas: 0,
		},
		{
			msg: "invalid trace config - Invalid Tracer",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(unitNetwork)
				defaultReq.TraceConfig = &types.TraceConfig{
					Tracer: "invalid_tracer",
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},

			//
			// malleate: func() {
			// 	traceConfig = &types.TraceConfig{
			// 		DisableStack:   true,
			// 		DisableStorage: true,
			// 		EnableMemory:   false,
			// 		Tracer:         "invalid_tracer",
			// 	}
			// },
			expPass:       true,
			traceResponse: "[{\"error\":\"rpc error: code = Internal desc = tracer not found\"}]",
			expFinalGas:   expGasConsumed,
		},
		{
			msg: "invalid chain id",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(unitNetwork)
				defaultReq.ChainId = -1
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			// malleate: func() {
			// 	traceConfig = nil
			// 	tmp := sdkmath.NewInt(1)
			// 	chainID = &tmp
			// },
			expPass:       true,
			traceResponse: "[{\"error\":\"rpc error: code = Internal desc = invalid chain id for signer\"}]",
			expFinalGas:   expGasConsumed,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// ----- Contract Deployment -----
			key := keyring.GetKey(0)
			constructorArgs := []interface{}{
				key.Addr,
				sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
			}
			compiledContract := types.ERC20Contract
			contractAddr, err := txFactory.DeployContract(
				key.Priv,
				types.EvmTxArgs{}, // Default values
				factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				},
			)
			suite.Require().NoError(err)

			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)

			// --- Contract Call ---
			newIndex := keyring.AddKey()
			recipient := keyring.GetAddr(newIndex)

			transferArgs := types.EvmTxArgs{
				To: &contractAddr,
			}
			callArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "transfer",
				Args:        []interface{}{recipient, big.NewInt(1000)},
			}

			// --- Add predecessor ---
			txs := tc.getAdditionalTxs()

			// Create the transaction to be traced
			// Note - it is important this is on the same block as the predecessor
			input, err := callArgs.ContractABI.Pack(callArgs.MethodName, callArgs.Args...)
			suite.Require().NoError(err)
			transferArgs.Input = input

			msgToTrace, err := txFactory.GenerateMsgEthereumTx(key.Priv, transferArgs)
			suite.Require().NoError(err)

			signer := ethtypes.LatestSignerForChainID(unitNetwork.GetEIP155ChainID())
			err = msgToTrace.Sign(signer, utiltx.NewSigner(key.Priv))
			suite.Require().NoError(err)

			txRes, err := txFactory.ExecuteContractCall(key.Priv, transferArgs, callArgs)
			suite.Require().NoError(err)
			suite.Require().True(txRes.IsOK())

			txs = append(txs, &msgToTrace)

			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)

			// Get the trace request
			traceReq := tc.getRequest()
			// Add txs to trace request
			traceReq.Txs = txs

			res, err := unitNetwork.GetEvmClient().TraceBlock(unitNetwork.GetContext(), &traceReq)

			if tc.expPass {
				suite.Require().NoError(err)
				// if data is to big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.traceResponse, string(res.Data[:150]))
				} else {
					suite.Require().Equal(tc.traceResponse, string(res.Data))
				}
			} else {
				suite.Require().Error(err)
			}
			suite.Require().Equal(int64(tc.expFinalGas), int64(unitNetwork.GetContext().GasMeter().GasConsumed()), "expected different gas consumption")

			// Clean up per test
			suite.Require().NoError(unitNetwork.NextBlock())
		})
	}
}

func (suite *KeeperTestSuite) TestNonceInQuery() {
	address := utiltx.GenerateAddress()
	suite.Require().Equal(uint64(0), suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), address))
	supply := sdkmath.NewIntWithDecimal(1000, 18).BigInt()

	// accupy nonce 0
	_ = suite.DeployTestContract(suite.T(), address, supply)

	// do an EthCall/EstimateGas with nonce 0
	ctorArgs, err := types.ERC20Contract.ABI.Pack("", address, supply)
	suite.Require().NoError(err)

	data := types.ERC20Contract.Bin
	data = append(data, ctorArgs...)
	args, err := json.Marshal(&types.TransactionArgs{
		From: &address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)
	proposerAddress := suite.network.GetContext().BlockHeader().ProposerAddress
	_, err = suite.network.GetEvmClient().EstimateGas(suite.network.GetContext(), &types.EthCallRequest{
		Args:            args,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: proposerAddress,
	})
	suite.Require().NoError(err)

	_, err = suite.network.GetEvmClient().EthCall(suite.network.GetContext(), &types.EthCallRequest{
		Args:            args,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: proposerAddress,
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestQueryBaseFee() {
	var (
		aux    sdkmath.Int
		expRes *types.QueryBaseFeeResponse
	)

	testCases := []struct {
		name            string
		malleate        func()
		expPass         bool
		enableFeemarket bool
		enableLondonHF  bool
	}{
		{
			"pass - default Base Fee",
			func() {
				initialBaseFee := sdkmath.NewInt(ethparams.InitialBaseFee)
				expRes = &types.QueryBaseFeeResponse{BaseFee: &initialBaseFee}
			},
			true, true, true,
		},
		{
			"pass - non-nil Base Fee",
			func() {
				baseFee := sdkmath.OneInt().BigInt()
				suite.network.App.FeeMarketKeeper.SetBaseFee(suite.network.GetContext(), baseFee)

				aux = sdkmath.NewIntFromBigInt(baseFee)
				expRes = &types.QueryBaseFeeResponse{BaseFee: &aux}
			},
			true, true, true,
		},
		{
			"pass - nil Base Fee when london hardfork not activated",
			func() {
				baseFee := sdkmath.OneInt().BigInt()
				suite.network.App.FeeMarketKeeper.SetBaseFee(suite.network.GetContext(), baseFee)

				expRes = &types.QueryBaseFeeResponse{}
			},
			true, true, false,
		},
		{
			"pass - zero Base Fee when feemarket not activated",
			func() {
				baseFee := sdkmath.ZeroInt()
				expRes = &types.QueryBaseFeeResponse{BaseFee: &baseFee}
			},
			true, false, true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest()

			tc.malleate()

			res, err := suite.network.GetEvmClient().BaseFee(suite.network.GetContext().Context(), &types.QueryBaseFeeRequest{})
			if tc.expPass {
				suite.Require().NotNil(res)
				suite.Require().Equal(expRes, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *KeeperTestSuite) TestEthCall() {
	var req *types.EthCallRequest

	address := utiltx.GenerateAddress()
	suite.Require().Equal(uint64(0), suite.network.App.EvmKeeper.GetNonce(suite.network.GetContext(), address))
	supply := sdkmath.NewIntWithDecimal(1000, 18).BigInt()

	hexBigInt := hexutil.Big(*big.NewInt(1))
	ctorArgs, err := types.ERC20Contract.ABI.Pack("", address, supply)
	suite.Require().NoError(err)

	data := types.ERC20Contract.Bin
	data = append(data, ctorArgs...)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid args",
			func() {
				req = &types.EthCallRequest{Args: []byte("invalid args"), GasCap: config.DefaultGasCap}
			},
			false,
		},
		{
			"invalid args - specified both gasPrice and maxFeePerGas",
			func() {
				args, err := json.Marshal(&types.TransactionArgs{
					From:         &address,
					Data:         (*hexutil.Bytes)(&data),
					GasPrice:     &hexBigInt,
					MaxFeePerGas: &hexBigInt,
				})

				suite.Require().NoError(err)
				req = &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}
			},
			false,
		},
		{
			"set param EnableCreate = false",
			func() {
				args, err := json.Marshal(&types.TransactionArgs{
					From: &address,
					Data: (*hexutil.Bytes)(&data),
				})

				suite.Require().NoError(err)
				req = &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}

				params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				params.EnableCreate = false
				err = suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), params)
				suite.Require().NoError(err)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			res, err := suite.network.GetEvmClient().EthCall(suite.network.GetContext(), req)
			if tc.expPass {
				suite.Require().NotNil(res)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEmptyRequest() {
	k := suite.network.App.EvmKeeper

	testCases := []struct {
		name      string
		queryFunc func() (interface{}, error)
	}{
		{
			"Account method",
			func() (interface{}, error) {
				return k.Account(suite.network.GetContext(), nil)
			},
		},
		{
			"CosmosAccount method",
			func() (interface{}, error) {
				return k.CosmosAccount(suite.network.GetContext(), nil)
			},
		},
		{
			"ValidatorAccount method",
			func() (interface{}, error) {
				return k.ValidatorAccount(suite.network.GetContext(), nil)
			},
		},
		{
			"Balance method",
			func() (interface{}, error) {
				return k.Balance(suite.network.GetContext(), nil)
			},
		},
		{
			"Storage method",
			func() (interface{}, error) {
				return k.Storage(suite.network.GetContext(), nil)
			},
		},
		{
			"Code method",
			func() (interface{}, error) {
				return k.Code(suite.network.GetContext(), nil)
			},
		},
		{
			"EthCall method",
			func() (interface{}, error) {
				return k.EthCall(suite.network.GetContext(), nil)
			},
		},
		{
			"EstimateGas method",
			func() (interface{}, error) {
				return k.EstimateGas(suite.network.GetContext(), nil)
			},
		},
		{
			"TraceTx method",
			func() (interface{}, error) {
				return k.TraceTx(suite.network.GetContext(), nil)
			},
		},
		{
			"TraceBlock method",
			func() (interface{}, error) {
				return k.TraceBlock(suite.network.GetContext(), nil)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			_, err := tc.queryFunc()
			suite.Require().Error(err)
		})
	}
}
