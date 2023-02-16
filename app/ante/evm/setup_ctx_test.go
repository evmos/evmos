package evm_test

import (
	"math/big"

	evmante "github.com/evmos/evmos/v11/app/ante/evm"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
)

func (suite *AnteTestSuite) TestEthSetupContextDecorator() {
	dec := evmante.NewEthSetUpContextDecorator(suite.app.EvmKeeper)
	ethTxContractParams := &evmtypes.EvmTxParams{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	tx := evmtypes.NewTxContract(ethTxContractParams)

	testCases := []struct {
		name    string
		tx      sdk.Tx
		expPass bool
	}{
		{"invalid transaction type - does not implement GasTx", &invalidTx{}, false},
		{
			"success - transaction implement GasTx",
			tx,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			ctx, err := dec.AnteHandle(suite.ctx, tc.tx, false, NextFn)

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
