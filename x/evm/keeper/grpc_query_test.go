package keeper_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/evmos/evmos/v20/x/evm/keeper/testdata"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"
	ethlogger "github.com/evmos/evmos/v20/x/evm/core/logger"
	"github.com/evmos/evmos/v20/x/evm/core/vm"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v20/server/config"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	"github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

// Not valid Ethereum address
const invalidAddress = "0x0000"

func (suite *KeeperTestSuite) TestQueryAccount() {
	baseDenom := types.GetEVMCoinDenom()
	testCases := []struct {
		msg         string
		getReq      func() *types.QueryAccountRequest
		expResponse *types.QueryAccountResponse
		expPass     bool
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
				amt := sdk.Coins{sdk.NewInt64Coin(baseDenom, 100)}

				// Add new unfunded key
				index := suite.keyring.AddKey()
				addr := suite.keyring.GetAddr(index)

				err := suite.network.App.BankKeeper.MintCoins(
					suite.network.GetContext(),
					types.ModuleName,
					amt,
				)
				suite.Require().NoError(err)

				err = suite.network.App.BankKeeper.SendCoinsFromModuleToAccount(
					suite.network.GetContext(),
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
			expectedResponse := tc.expResponse

			ctx := suite.network.GetContext()
			// Function under test
			res, err := suite.network.GetEvmClient().Account(ctx, req)

			suite.Require().Equal(expectedResponse, res)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCosmosAccount() {
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
				key := suite.keyring.GetKey(0)
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
				index := suite.keyring.AddKey()
				newKey := suite.keyring.GetKey(index)
				accountNumber := uint64(100)
				acc := suite.network.App.AccountKeeper.NewAccountWithAddress(
					suite.network.GetContext(),
					newKey.AccAddr,
				)

				suite.Require().NoError(acc.SetSequence(10))
				suite.Require().NoError(acc.SetAccountNumber(accountNumber))
				suite.network.App.AccountKeeper.SetAccount(suite.network.GetContext(), acc)

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

			ctx := suite.network.GetContext()

			// Function under test
			res, err := suite.network.GetEvmClient().CosmosAccount(ctx, req)

			suite.Require().Equal(expectedResponse, res)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryBalance() {
	baseDenom := types.GetEVMCoinDenom()

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
				newIndex := suite.keyring.AddKey()
				addr := suite.keyring.GetAddr(newIndex)

				balance := int64(100)
				amt := sdk.Coins{sdk.NewInt64Coin(baseDenom, balance)}

				err := suite.network.App.BankKeeper.MintCoins(suite.network.GetContext(), types.ModuleName, amt)
				suite.Require().NoError(err)
				err = suite.network.App.BankKeeper.SendCoinsFromModuleToAccount(suite.network.GetContext(), types.ModuleName, addr.Bytes(), amt)
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

			ctx := suite.network.GetContext()
			res, err := suite.network.GetEvmClient().Balance(ctx, req)

			suite.Require().Equal(resp, res)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryStorage() {
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

				newIndex := suite.keyring.AddKey()
				addr := suite.keyring.GetAddr(newIndex)

				suite.network.App.EvmKeeper.SetState(
					suite.network.GetContext(),
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

			ctx := suite.network.GetContext()
			res, err := suite.network.GetEvmClient().Storage(ctx, req)

			suite.Require().Equal(expectedResp, res)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCode() {
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
				newIndex := suite.keyring.AddKey()
				addr := suite.keyring.GetAddr(newIndex)

				expCode = []byte("code")
				stateDB := suite.network.GetStateDB()
				stateDB.SetCode(addr, expCode)
				suite.Require().NoError(stateDB.Commit())

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

			ctx := suite.network.GetContext()
			res, err := suite.network.GetEvmClient().Code(ctx, req)

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
	expLogs := []*types.Log{}
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
			txCfg := statedb.NewTxConfig(
				common.BytesToHash(suite.network.GetContext().HeaderHash()),
				txHash,
				txIndex,
				logIndex,
			)
			vmdb := statedb.New(
				suite.network.GetContext(),
				suite.network.App.EvmKeeper,
				txCfg,
			)

			tc.malleate(vmdb)
			suite.Require().NoError(vmdb.Commit())

			logs := vmdb.Logs()
			suite.Require().Equal(expLogs, types.NewLogsFromEth(logs))
		})
	}
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := suite.network.GetContext()
	expParams := types.DefaultParams()

	res, err := suite.network.GetEvmClient().Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

func (suite *KeeperTestSuite) TestQueryValidatorAccount() {
	testCases := []struct {
		msg           string
		getReqAndResp func() (*types.QueryValidatorAccountRequest, *types.QueryValidatorAccountResponse)
		expPass       bool
	}{
		{
			"invalid address",
			func() (*types.QueryValidatorAccountRequest, *types.QueryValidatorAccountResponse) {
				req := &types.QueryValidatorAccountRequest{
					ConsAddress: "",
				}
				return req, nil
			},
			false,
		},
		{
			"success",
			func() (*types.QueryValidatorAccountRequest, *types.QueryValidatorAccountResponse) {
				val := suite.network.GetValidators()[0]
				consAddr, err := val.GetConsAddr()
				suite.Require().NoError(err)

				req := &types.QueryValidatorAccountRequest{
					ConsAddress: sdk.ConsAddress(consAddr).String(),
				}

				addrBz, err := suite.network.App.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
				suite.Require().NoError(err)

				resp := &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(addrBz).String(),
					Sequence:       0,
					AccountNumber:  2,
				}

				return req, resp
			},
			true,
		},
		{
			"success with seq and account number",
			func() (*types.QueryValidatorAccountRequest, *types.QueryValidatorAccountResponse) {
				val := suite.network.GetValidators()[0]
				consAddr, err := val.GetConsAddr()
				suite.Require().NoError(err)

				// Create validator account and set sequence and account number
				accNumber := uint64(100)
				accSeq := uint64(10)

				addrBz, err := suite.network.App.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
				suite.Require().NoError(err)

				accAddrStr := sdk.AccAddress(addrBz).String()

				baseAcc := &authtypes.BaseAccount{Address: accAddrStr}
				acc := suite.network.App.AccountKeeper.NewAccount(suite.network.GetContext(), baseAcc)
				suite.Require().NoError(acc.SetSequence(accSeq))
				suite.Require().NoError(acc.SetAccountNumber(accNumber))
				suite.network.App.AccountKeeper.SetAccount(suite.network.GetContext(), acc)

				resp := &types.QueryValidatorAccountResponse{
					AccountAddress: accAddrStr,
					Sequence:       accSeq,
					AccountNumber:  accNumber,
				}
				req := &types.QueryValidatorAccountRequest{
					ConsAddress: sdk.ConsAddress(consAddr).String(),
				}

				return req, resp
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			req, resp := tc.getReqAndResp()
			ctx := suite.network.GetContext()
			res, err := suite.network.GetEvmClient().ValidatorAccount(ctx, req)

			suite.Require().Equal(resp, res)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateGas() {
	gasHelper := hexutil.Uint64(20000)
	higherGas := hexutil.Uint64(25000)
	// Hardcode recipient address to avoid non determinism in tests
	hardcodedRecipient := common.HexToAddress("0xC6Fe5D33615a1C52c08018c47E8Bc53646A0E101")

	erc20Contract, err := testdata.LoadERC20Contract()
	suite.Require().NoError(err)

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
				addr := suite.keyring.GetAddr(0)
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
		{
			"fail - not enough balance w/ gas fee cap",
			func() types.TransactionArgs {
				addr := suite.keyring.GetAddr(0)
				hexBigInt := hexutil.Big(*big.NewInt(1))
				balance := suite.network.App.BankKeeper.GetBalance(suite.network.GetContext(), sdk.AccAddress(addr.Bytes()), types.GetEVMCoinDenom())
				value := balance.Amount.Add(sdkmath.NewInt(1))
				return types.TransactionArgs{
					To:           &common.Address{},
					From:         &addr,
					Value:        (*hexutil.Big)(value.BigInt()),
					MaxFeePerGas: &hexBigInt,
				}
			},
			false,
			0,
			false,
			config.DefaultGasCap,
		},
		{
			"fail - insufficient funds for gas * price + value w/ gas fee cap",
			func() types.TransactionArgs {
				addr := suite.keyring.GetAddr(0)
				hexBigInt := hexutil.Big(*big.NewInt(1))
				balance := suite.network.App.BankKeeper.GetBalance(suite.network.GetContext(), sdk.AccAddress(addr.Bytes()), types.GetEVMCoinDenom())
				value := balance.Amount.Sub(sdkmath.NewInt(1))
				return types.TransactionArgs{
					To:           &common.Address{},
					From:         &addr,
					Value:        (*hexutil.Big)(value.BigInt()),
					MaxFeePerGas: &hexBigInt,
				}
			},
			false,
			0,
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
				ctorArgs, err := erc20Contract.ABI.Pack(
					"",
					&hardcodedRecipient,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				)
				suite.Require().NoError(err)
				data := erc20Contract.Bin
				data = append(data, ctorArgs...)

				addr := suite.keyring.GetAddr(0)
				return types.TransactionArgs{
					Data: (*hexutil.Bytes)(&data),
					From: &addr,
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
				key := suite.keyring.GetKey(0)
				contractAddr, err := deployErc20Contract(key, suite.factory)
				suite.Require().NoError(err)

				err = suite.network.NextBlock()
				suite.Require().NoError(err)

				transferData, err := erc20Contract.ABI.Pack(
					"transfer",
					hardcodedRecipient,
					big.NewInt(1000),
				)
				suite.Require().NoError(err)
				return types.TransactionArgs{
					To:   &contractAddr,
					Data: (*hexutil.Bytes)(&transferData),
					From: &key.Addr,
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
				addr := suite.keyring.GetAddr(0)
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
				ctorArgs, err := erc20Contract.ABI.Pack(
					"",
					&hardcodedRecipient,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				)
				suite.Require().NoError(err)
				data := erc20Contract.Bin
				data = append(data, ctorArgs...)

				sender := suite.keyring.GetAddr(0)
				return types.TransactionArgs{
					Data: (*hexutil.Bytes)(&data),
					From: &sender,
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
				key := suite.keyring.GetKey(1)

				contractAddr, err := deployErc20Contract(key, suite.factory)
				suite.Require().NoError(err)

				err = suite.network.NextBlock()
				suite.Require().NoError(err)

				transferData, err := erc20Contract.ABI.Pack(
					"transfer",
					hardcodedRecipient,
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
				addr := suite.keyring.GetAddr(0)
				ctorArgs, err := erc20Contract.ABI.Pack(
					"",
					&addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				)
				suite.Require().NoError(err)

				data := erc20Contract.Bin
				data = append(data, ctorArgs...)

				args := types.TransactionArgs{
					From: &addr,
					Data: (*hexutil.Bytes)(&data),
				}
				params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				params.AccessControl = types.AccessControl{
					Create: types.AccessControlType{
						AccessType: types.AccessTypeRestricted,
					},
				}
				err = suite.network.App.EvmKeeper.SetParams(
					suite.network.GetContext(),
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
				hexBigInt := hexutil.Big(*big.NewInt(1))

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
			// Start from a clean state
			suite.Require().NoError(suite.network.NextBlock())

			// Update feemarket params per test
			evmParams := feemarkettypes.DefaultParams()
			if !tc.enableFeemarket {
				evmParams := suite.network.App.FeeMarketKeeper.GetParams(
					suite.network.GetContext(),
				)
				evmParams.NoBaseFee = true
			}

			err := suite.network.App.FeeMarketKeeper.SetParams(
				suite.network.GetContext(),
				evmParams,
			)
			suite.Require().NoError(err)

			// Get call args
			args := tc.getArgs()
			marshalArgs, err := json.Marshal(args)
			suite.Require().NoError(err)

			req := types.EthCallRequest{
				Args:            marshalArgs,
				GasCap:          tc.gasCap,
				ProposerAddress: suite.network.GetContext().BlockHeader().ProposerAddress,
			}

			// Function under test
			rsp, err := suite.network.GetEvmClient().EstimateGas(
				suite.network.GetContext(),
				&req,
			)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(int64(tc.expGas), int64(rsp.Gas)) //#nosec G115
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

func (suite *KeeperTestSuite) TestTraceTx() {
	suite.enableFeemarket = true
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()

	// Hardcode recipient address to avoid non determinism in tests
	hardcodedRecipient := common.HexToAddress("0xC6Fe5D33615a1C52c08018c47E8Bc53646A0E101")

	erc20Contract, err := testdata.LoadERC20Contract()
	suite.Require().NoError(err)

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
				return getDefaultTraceTxRequest(suite.network)
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
				defaultRequest := getDefaultTraceTxRequest(suite.network)
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
				defaultRequest := getDefaultTraceTxRequest(suite.network)
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
				return getDefaultTraceTxRequest(suite.network)
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				// Create predecessor tx
				// Use different address to avoid nonce collision
				senderKey := suite.keyring.GetKey(1)
				contractAddr, err := deployErc20Contract(senderKey, suite.factory)
				suite.Require().NoError(err)

				err = suite.network.NextBlock()
				suite.Require().NoError(err)

				txMsg, err := executeTransferCall(
					transferParams{
						senderKey:     senderKey,
						contractAddr:  contractAddr,
						recipientAddr: hardcodedRecipient,
					},
					suite.factory,
				)
				suite.Require().NoError(err)

				return []*types.MsgEthereumTx{txMsg}
			},
			expPass:       true,
			expectedTrace: "{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "invalid trace config - Negative Limit",
			getRequest: func() types.QueryTraceTxRequest {
				defaultRequest := getDefaultTraceTxRequest(suite.network)
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
				defaultRequest := getDefaultTraceTxRequest(suite.network)
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
				defaultRequest := getDefaultTraceTxRequest(suite.network)
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
				return getDefaultTraceTxRequest(suite.network)
			},
			getPredecessors: func() []*types.MsgEthereumTx {
				// use different address to avoid nonce collision
				senderKey := suite.keyring.GetKey(1)

				constructorArgs := []interface{}{
					senderKey.Addr,
					sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
				}
				compiledContract := erc20Contract
				deploymentData := factory.ContractDeploymentData{
					Contract:        compiledContract,
					ConstructorArgs: constructorArgs,
				}

				txArgs, err := suite.factory.GenerateDeployContractArgs(senderKey.Addr, types.EvmTxArgs{}, deploymentData)
				suite.Require().NoError(err)

				txMsg, err := suite.factory.GenerateMsgEthereumTx(senderKey.Priv, txArgs)
				suite.Require().NoError(err)

				_, err = suite.factory.ExecuteEthTx(
					senderKey.Priv,
					txArgs, // Default values
				)
				suite.Require().NoError(err)

				params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				params.AccessControl = types.AccessControl{
					Create: types.AccessControlType{
						AccessType: types.AccessTypeRestricted,
					},
				}
				err = suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), params)
				suite.Require().NoError(err)
				return []*types.MsgEthereumTx{&txMsg}
			},
			expPass:       true,
			expectedTrace: "{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
			// expFinalGas:   26744, // gas consumed in traceTx setup (GetProposerAddr + CalculateBaseFee) + gas consumed in malleate func
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// Clean up per test
			defaultEvmParams := types.DefaultParams()
			err := suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), defaultEvmParams)
			suite.Require().NoError(err)

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			// ----- Contract Deployment -----
			senderKey := suite.keyring.GetKey(0)
			contractAddr, err := deployErc20Contract(senderKey, suite.factory)
			suite.Require().NoError(err)

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			// --- Add predecessor ---
			predecessors := tc.getPredecessors()

			// Get the message to trace
			msgToTrace, err := executeTransferCall(
				transferParams{
					senderKey:     senderKey,
					contractAddr:  contractAddr,
					recipientAddr: hardcodedRecipient,
				},
				suite.factory,
			)
			suite.Require().NoError(err)

			suite.Require().NoError(suite.network.NextBlock())

			// Get the trace request
			traceReq := tc.getRequest()
			// Add predecessor to trace request
			traceReq.Predecessors = predecessors
			traceReq.Msg = msgToTrace

			// Function under test
			res, err := suite.network.GetEvmClient().TraceTx(
				suite.network.GetContext(),
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
		})
	}
}

func (suite *KeeperTestSuite) TestTraceBlock() {
	suite.enableFeemarket = true
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()

	// Hardcode recipient to make gas estimation deterministic
	hardcodedTransferRecipient := common.HexToAddress("0xC6Fe5D33615a1C52c08018c47E8Bc53646A0E101")

	testCases := []struct {
		msg              string
		getRequest       func() types.QueryTraceBlockRequest
		getAdditionalTxs func() []*types.MsgEthereumTx
		expPass          bool
		traceResponse    string
	}{
		{
			msg: "default trace",
			getRequest: func() types.QueryTraceBlockRequest {
				return getDefaultTraceBlockRequest(suite.network)
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
		},
		{
			msg: "filtered trace",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(suite.network)
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
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
		},
		{
			msg: "javascript tracer",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(suite.network)
				defaultReq.TraceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass:       true,
			traceResponse: "[{\"result\":[]}]",
		},
		{
			msg: "tracer with multiple transactions",
			getRequest: func() types.QueryTraceBlockRequest {
				return getDefaultTraceBlockRequest(suite.network)
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				// Create predecessor tx
				// Use different address to avoid nonce collision
				senderKey := suite.keyring.GetKey(1)
				contractAddr, err := deployErc20Contract(senderKey, suite.factory)
				suite.Require().NoError(err)

				err = suite.network.NextBlock()
				suite.Require().NoError(err)

				firstTransferMessage, err := executeTransferCall(
					transferParams{
						senderKey:     suite.keyring.GetKey(1),
						contractAddr:  contractAddr,
						recipientAddr: hardcodedTransferRecipient,
					},
					suite.factory,
				)
				suite.Require().NoError(err)
				return []*types.MsgEthereumTx{firstTransferMessage}
			},
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34780,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
		},
		{
			msg: "invalid trace config - Negative Limit",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(suite.network)
				defaultReq.TraceConfig = &types.TraceConfig{
					Limit: -1,
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass: false,
		},
		{
			msg: "invalid trace config - Invalid Tracer",
			getRequest: func() types.QueryTraceBlockRequest {
				defaultReq := getDefaultTraceBlockRequest(suite.network)
				defaultReq.TraceConfig = &types.TraceConfig{
					Tracer: "invalid_tracer",
				}
				return defaultReq
			},
			getAdditionalTxs: func() []*types.MsgEthereumTx {
				return nil
			},
			expPass:       true,
			traceResponse: "[{\"error\":\"rpc error: code = Internal desc = tracer not found\"}]",
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// Start from fresh block
			suite.Require().NoError(suite.network.NextBlock())

			// ----- Contract Deployment -----
			senderKey := suite.keyring.GetKey(0)
			contractAddr, err := deployErc20Contract(senderKey, suite.factory)
			suite.Require().NoError(err)

			err = suite.network.NextBlock()
			suite.Require().NoError(err)

			// --- Add predecessor ---
			txs := tc.getAdditionalTxs()

			// --- Contract Call ---
			msgToTrace, err := executeTransferCall(
				transferParams{
					senderKey:     senderKey,
					contractAddr:  contractAddr,
					recipientAddr: hardcodedTransferRecipient,
				},
				suite.factory,
			)
			suite.Require().NoError(err)
			txs = append(txs, msgToTrace)

			suite.Require().NoError(suite.network.NextBlock())

			// Get the trace request
			traceReq := tc.getRequest()
			// Add txs to trace request
			traceReq.Txs = txs

			res, err := suite.network.GetEvmClient().TraceBlock(suite.network.GetContext(), &traceReq)

			if tc.expPass {
				suite.Require().NoError(err)
				// if data is too big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.traceResponse, string(res.Data[:150]))
				} else {
					suite.Require().Equal(tc.traceResponse, string(res.Data))
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestNonceInQuery() {
	suite.enableFeemarket = true
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()

	senderKey := suite.keyring.GetKey(0)
	nonce := suite.network.App.EvmKeeper.GetNonce(
		suite.network.GetContext(),
		senderKey.Addr,
	)
	suite.Require().Equal(uint64(0), nonce)

	// accupy nonce 0
	_, err := deployErc20Contract(suite.keyring.GetKey(0), suite.factory)
	suite.Require().NoError(err)

	erc20Contract, err := testdata.LoadERC20Contract()
	suite.Require().NoError(err, "failed to load erc20 contract")

	// do an EthCall/EstimateGas with nonce 0
	ctorArgs, err := erc20Contract.ABI.Pack("", senderKey.Addr, big.NewInt(1000))
	suite.Require().NoError(err)

	data := erc20Contract.Bin
	data = append(data, ctorArgs...)
	args, err := json.Marshal(&types.TransactionArgs{
		From: &senderKey.Addr,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	proposerAddress := suite.network.GetContext().BlockHeader().ProposerAddress
	_, err = suite.network.GetEvmClient().EstimateGas(
		suite.network.GetContext(),
		&types.EthCallRequest{
			Args:            args,
			GasCap:          config.DefaultGasCap,
			ProposerAddress: proposerAddress,
		},
	)
	suite.Require().NoError(err)

	_, err = suite.network.GetEvmClient().EthCall(
		suite.network.GetContext(),
		&types.EthCallRequest{
			Args:            args,
			GasCap:          config.DefaultGasCap,
			ProposerAddress: proposerAddress,
		},
	)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestQueryBaseFee() {
	suite.enableFeemarket = true
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()

	testCases := []struct {
		name       string
		getExpResp func() *types.QueryBaseFeeResponse
		setParams  func()
		expPass    bool
	}{
		{
			"pass - default Base Fee",
			func() *types.QueryBaseFeeResponse {
				initialBaseFee := sdkmath.NewInt(ethparams.InitialBaseFee)
				return &types.QueryBaseFeeResponse{BaseFee: &initialBaseFee}
			},
			func() {
				feemarketDefault := feemarkettypes.DefaultParams()
				suite.Require().NoError(suite.network.App.FeeMarketKeeper.SetParams(suite.network.GetContext(), feemarketDefault))

				evmDefault := types.DefaultParams()
				suite.Require().NoError(suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), evmDefault))
			},

			true,
		},
		{
			"pass - nil Base Fee when london hardfork not activated",
			func() *types.QueryBaseFeeResponse {
				return &types.QueryBaseFeeResponse{}
			},
			func() {
				feemarketDefault := feemarkettypes.DefaultParams()
				suite.Require().NoError(suite.network.App.FeeMarketKeeper.SetParams(suite.network.GetContext(), feemarketDefault))
				chainConfig := types.DefaultChainConfig(suite.network.GetChainID())
				maxInt := sdkmath.NewInt(math.MaxInt64)
				chainConfig.LondonBlock = &maxInt
				chainConfig.ArrowGlacierBlock = &maxInt
				chainConfig.GrayGlacierBlock = &maxInt
				chainConfig.MergeNetsplitBlock = &maxInt
				chainConfig.ShanghaiBlock = &maxInt
				chainConfig.CancunBlock = &maxInt
				configurator := types.NewEVMConfigurator()
				configurator.ResetTestChainConfig()
				err := configurator.
					WithChainConfig(chainConfig).
					Configure()
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"pass - zero Base Fee when feemarket not activated",
			func() *types.QueryBaseFeeResponse {
				baseFee := sdkmath.ZeroInt()
				return &types.QueryBaseFeeResponse{BaseFee: &baseFee}
			},
			func() {
				feemarketDefault := feemarkettypes.DefaultParams()
				feemarketDefault.NoBaseFee = true
				suite.Require().NoError(suite.network.App.FeeMarketKeeper.SetParams(suite.network.GetContext(), feemarketDefault))

				evmDefault := types.DefaultParams()
				suite.Require().NoError(suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), evmDefault))
			},
			true,
		},
	}

	// Save initial configure to restore it between tests
	denom := types.GetEVMCoinDenom()
	decimals := types.GetEVMCoinDecimals()
	chainConfig := types.DefaultChainConfig(suite.network.GetChainID())

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			// Set necessary params
			tc.setParams()
			// Get the expected response
			expResp := tc.getExpResp()
			// Function under test
			res, err := suite.network.GetEvmClient().BaseFee(
				suite.network.GetContext(),
				&types.QueryBaseFeeRequest{},
			)
			if tc.expPass {
				suite.Require().NotNil(res)
				suite.Require().Equal(expResp, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
			suite.Require().NoError(suite.network.NextBlock())
			configurator := types.NewEVMConfigurator()
			configurator.ResetTestChainConfig()
			err = configurator.
				WithChainConfig(chainConfig).
				WithEVMCoinInfo(denom, uint8(decimals)).
				Configure()
			suite.Require().NoError(err)
		})
	}
}

func (suite *KeeperTestSuite) TestEthCall() {
	suite.SetupTest()

	erc20Contract, err := testdata.LoadERC20Contract()
	suite.Require().NoError(err)

	// Generate common data for requests
	sender := suite.keyring.GetAddr(0)
	supply := sdkmath.NewIntWithDecimal(1000, 18).BigInt()
	ctorArgs, err := erc20Contract.ABI.Pack("", sender, supply)
	suite.Require().NoError(err)
	data := erc20Contract.Bin
	data = append(data, ctorArgs...)

	testCases := []struct {
		name       string
		getReq     func() *types.EthCallRequest
		expVMError bool
	}{
		{
			"invalid args",
			func() *types.EthCallRequest {
				return &types.EthCallRequest{Args: []byte("invalid args"), GasCap: config.DefaultGasCap}
			},
			false,
		},
		{
			"invalid args - specified both gasPrice and maxFeePerGas",
			func() *types.EthCallRequest {
				hexBigInt := hexutil.Big(*big.NewInt(1))
				args, err := json.Marshal(&types.TransactionArgs{
					From:         &sender,
					Data:         (*hexutil.Bytes)(&data),
					GasPrice:     &hexBigInt,
					MaxFeePerGas: &hexBigInt,
				})
				suite.Require().NoError(err)

				return &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}
			},
			false,
		},
		{
			"set param AccessControl - no Access",
			func() *types.EthCallRequest {
				args, err := json.Marshal(&types.TransactionArgs{
					From: &sender,
					Data: (*hexutil.Bytes)(&data),
				})

				suite.Require().NoError(err)
				req := &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}

				params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				params.AccessControl = types.AccessControl{
					Create: types.AccessControlType{
						AccessType: types.AccessTypeRestricted,
					},
				}
				err = suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), params)
				suite.Require().NoError(err)
				return req
			},
			true,
		},
		{
			"set param AccessControl = non whitelist",
			func() *types.EthCallRequest {
				args, err := json.Marshal(&types.TransactionArgs{
					From: &sender,
					Data: (*hexutil.Bytes)(&data),
				})

				suite.Require().NoError(err)
				req := &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}

				params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				params.AccessControl = types.AccessControl{
					Create: types.AccessControlType{
						AccessType: types.AccessTypePermissioned,
					},
				}
				err = suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), params)
				suite.Require().NoError(err)
				return req
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := tc.getReq()

			res, err := suite.network.GetEvmClient().EthCall(suite.network.GetContext(), req)
			if tc.expVMError {
				suite.Require().NotNil(res)
				suite.Require().Contains(res.VmError, "does not have permission to deploy contracts")
			} else {
				suite.Require().Error(err)
			}

			// Reset params
			defaultEvmParams := types.DefaultParams()
			err = suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), defaultEvmParams)
			suite.Require().NoError(err)
		})
	}
}

func (suite *KeeperTestSuite) TestEmptyRequest() {
	suite.SetupTest()
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
			_, err := tc.queryFunc()
			suite.Require().Error(err)
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

func deployErc20Contract(from testkeyring.Key, txFactory factory.TxFactory) (common.Address, error) {
	erc20Contract, err := testdata.LoadERC20Contract()
	if err != nil {
		return common.Address{}, err
	}

	constructorArgs := []interface{}{
		from.Addr,
		sdkmath.NewIntWithDecimal(1000, 18).BigInt(),
	}
	compiledContract := erc20Contract
	contractAddr, err := txFactory.DeployContract(
		from.Priv,
		types.EvmTxArgs{}, // Default values
		factory.ContractDeploymentData{
			Contract:        compiledContract,
			ConstructorArgs: constructorArgs,
		},
	)
	if err != nil {
		return common.Address{}, err
	}
	return contractAddr, nil
}

type transferParams struct {
	senderKey     testkeyring.Key
	contractAddr  common.Address
	recipientAddr common.Address
}

func executeTransferCall(
	transferParams transferParams,
	txFactory factory.TxFactory,
) (msgEthereumTx *types.MsgEthereumTx, err error) {
	erc20Contract, err := testdata.LoadERC20Contract()
	if err != nil {
		return nil, err
	}

	transferArgs := types.EvmTxArgs{
		To: &transferParams.contractAddr,
	}
	callArgs := factory.CallArgs{
		ContractABI: erc20Contract.ABI,
		MethodName:  "transfer",
		Args:        []interface{}{transferParams.recipientAddr, big.NewInt(1000)},
	}

	transferArgs, err = txFactory.GenerateContractCallArgs(transferArgs, callArgs)
	if err != nil {
		return nil, err
	}

	// We need to get access to the message
	firstSignedTX, err := txFactory.GenerateSignedEthTx(transferParams.senderKey.Priv, transferArgs)
	if err != nil {
		return nil, err
	}
	txMsg, ok := firstSignedTX.GetMsgs()[0].(*types.MsgEthereumTx)
	if !ok {
		return nil, fmt.Errorf("invalid type")
	}

	result, err := txFactory.ExecuteContractCall(transferParams.senderKey.Priv, transferArgs, callArgs)
	if err != nil || !result.IsOK() {
		return nil, err
	}
	return txMsg, nil
}
