package evm_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v16/app/ante/evm"
	"github.com/evmos/evmos/v16/testutil"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/types"
	"github.com/evmos/evmos/v16/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *AnteTestSuite) TestGasWantedDecorator() {
	suite.enableFeemarket = true
	suite.SetupTest()
	ctx := suite.network.GetContext()
	dec := evm.NewGasWantedDecorator(suite.network.App.EvmKeeper, suite.network.App.FeeMarketKeeper)
	from, fromPrivKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	testCases := []struct {
		name              string
		expectedGasWanted uint64
		malleate          func() sdk.Tx
		expPass           bool
	}{
		{
			"Cosmos Tx",
			TestGasLimit,
			func() sdk.Tx {
				denom := evmtypes.DefaultEVMDenom
				testMsg := banktypes.MsgSend{
					FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), utils.BaseDenom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
		},
		{
			"Ethereum Legacy Tx",
			TestGasLimit,
			func() sdk.Tx {
				txArgs := evmtypes.EvmTxArgs{
					To:       &to,
					GasPrice: big.NewInt(0),
					GasLimit: TestGasLimit,
				}
				return suite.CreateTxBuilder(fromPrivKey, txArgs).GetTx()
			},
			true,
		},
		{
			"Ethereum Access List Tx",
			TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				txArgs := evmtypes.EvmTxArgs{
					To:       &to,
					GasPrice: big.NewInt(0),
					GasLimit: TestGasLimit,
					Accesses: &emptyAccessList,
				}
				return suite.CreateTxBuilder(fromPrivKey, txArgs).GetTx()
			},
			true,
		},
		{
			"Ethereum Dynamic Fee Tx (EIP1559)",
			TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				txArgs := evmtypes.EvmTxArgs{
					To:        &to,
					GasPrice:  big.NewInt(0),
					GasFeeCap: big.NewInt(100),
					GasLimit:  TestGasLimit,
					GasTipCap: big.NewInt(50),
					Accesses:  &emptyAccessList,
				}
				return suite.CreateTxBuilder(fromPrivKey, txArgs).GetTx()
			},
			true,
		},
		{
			"EIP712 message",
			200000,
			func() sdk.Tx {
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20)))
				gas := uint64(200000)
				acc := suite.network.App.AccountKeeper.NewAccountWithAddress(ctx, from.Bytes())
				suite.Require().NoError(acc.SetSequence(1))
				suite.network.App.AccountKeeper.SetAccount(ctx, acc)
				builder, err := suite.CreateTestEIP712TxBuilderMsgSend(acc.GetAddress(), fromPrivKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return builder.GetTx()
			},
			true,
		},
		{
			"Cosmos Tx - gasWanted > max block gas",
			TestGasLimit,
			func() sdk.Tx {
				denom := evmtypes.DefaultEVMDenom
				testMsg := banktypes.MsgSend{
					FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), utils.BaseDenom, &testMsg)
				limit := types.BlockGasLimit(ctx)
				txBuilder.SetGasLimit(limit + 5)
				return txBuilder.GetTx()
			},
			false,
		},
	}

	// cumulative gas wanted from all test transactions in the same block
	var expectedGasWanted uint64

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := dec.AnteHandle(ctx, tc.malleate(), false, testutil.NextFn)
			if tc.expPass {
				suite.Require().NoError(err)

				gasWanted := suite.network.App.FeeMarketKeeper.GetTransientGasWanted(ctx)
				expectedGasWanted += tc.expectedGasWanted
				suite.Require().Equal(expectedGasWanted, gasWanted)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
