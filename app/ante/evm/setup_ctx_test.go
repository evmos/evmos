package evm_test

import (
	"math/big"

	evmante "github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/testutil"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
)

func (suite *AnteTestSuite) TestEthSetupContextDecorator() {
	dec := evmante.NewEthSetUpContextDecorator(suite.app.EvmKeeper)
	tx := evmtypes.NewTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil, nil, nil)

	testCases := []struct {
		name    string
		tx      sdk.Tx
		expPass bool
	}{
		{"invalid transaction type - does not implement GasTx", &testutiltx.InvalidTx{}, false},
		{
			"success - transaction implement GasTx",
			tx,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			ctx, err := dec.AnteHandle(suite.ctx, tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Equal(storetypes.GasConfig{}, ctx.KVGasConfig())
				suite.Equal(storetypes.GasConfig{}, ctx.TransientKVGasConfig())
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
