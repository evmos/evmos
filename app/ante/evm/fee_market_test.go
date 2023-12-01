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
	dec := evm.NewGasWantedDecorator(suite.app.EvmKeeper, suite.app.FeeMarketKeeper)
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
				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, nil)
				return suite.CreateTestTx(msg, fromPrivKey, 1, false)
			},
			true,
		},
		{
			"Ethereum Access List Tx",
			TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, &emptyAccessList)
				return suite.CreateTestTx(msg, fromPrivKey, 1, false)
			},
			true,
		},
		{
			"Ethereum Dynamic Fee Tx (EIP1559)",
			TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), big.NewInt(100), big.NewInt(50), &emptyAccessList)
				return suite.CreateTestTx(msg, fromPrivKey, 1, false)
			},
			true,
		},
		{
			"EIP712 message",
			200000,
			func() sdk.Tx {
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20)))
				gas := uint64(200000)
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, from.Bytes())
				suite.Require().NoError(acc.SetSequence(1))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
				builder, err := suite.CreateTestEIP712TxBuilderMsgSend(acc.GetAddress(), fromPrivKey, suite.ctx.ChainID(), gas, amount)
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
				limit := types.BlockGasLimit(suite.ctx)
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
			_, err := dec.AnteHandle(suite.ctx, tc.malleate(), false, testutil.NextFn)
			if tc.expPass {
				suite.Require().NoError(err)

				gasWanted := suite.app.FeeMarketKeeper.GetTransientGasWanted(suite.ctx)
				expectedGasWanted += tc.expectedGasWanted
				suite.Require().Equal(expectedGasWanted, gasWanted)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
