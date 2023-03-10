package evm_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethante "github.com/evmos/evmos/v12/app/ante/evm"
	"github.com/evmos/evmos/v12/testutil"
	testutiltx "github.com/evmos/evmos/v12/testutil/tx"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

func (suite *AnteTestSuite) TestEthSigVerificationDecorator() {
	addr, privKey := testutiltx.NewAddrKey()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	signedTx := evmtypes.NewTx(ethContractCreationTxParams)
	signedTx.From = addr.Hex()
	err := signedTx.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	uprotectedEthTxParams := &evmtypes.EvmTxArgs{
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	unprotectedTx := evmtypes.NewTx(uprotectedEthTxParams)
	unprotectedTx.From = addr.Hex()
	err = unprotectedTx.Sign(ethtypes.HomesteadSigner{}, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name                string
		tx                  sdk.Tx
		allowUnprotectedTxs bool
		reCheckTx           bool
		expPass             bool
	}{
		{"ReCheckTx", &testutiltx.InvalidTx{}, false, true, false},
		{"invalid transaction type", &testutiltx.InvalidTx{}, false, false, false},
		{
			"invalid sender",
			evmtypes.NewTx(&evmtypes.EvmTxArgs{
				To:       &addr,
				Nonce:    1,
				Amount:   big.NewInt(10),
				GasLimit: 1000,
				GasPrice: big.NewInt(1),
			}),
			true,
			false,
			false,
		},
		{"successful signature verification", signedTx, false, false, true},
		{"invalid, reject unprotected txs", unprotectedTx, false, false, false},
		{"successful, allow unprotected txs", unprotectedTx, true, false, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.evmParamsOption = func(params *evmtypes.Params) {
				params.AllowUnprotectedTxs = tc.allowUnprotectedTxs
			}
			suite.SetupTest()
			dec := ethante.NewEthSigVerificationDecorator(suite.app.EvmKeeper)
			_, err := dec.AnteHandle(suite.ctx.WithIsReCheckTx(tc.reCheckTx), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.evmParamsOption = nil
}
