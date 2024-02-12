package evm_test

import (
	"math/big"

	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *AnteTestSuite) TestSignatures() {
	suite.WithFeemarketEnabled(false)
	suite.SetupTest() // reset

	privKey := suite.GetKeyring().GetPrivKey(0)
	to := utiltx.GenerateAddress()

	txArgs := evmtypes.EvmTxArgs{
		ChainID:  suite.GetNetwork().App.EvmKeeper.ChainID(),
		Nonce:    0,
		To:       &to,
		Amount:   big.NewInt(10),
		GasLimit: 100000,
		GasPrice: big.NewInt(1),
	}

	// CreateTestTx will sign the msgEthereumTx but not sign the cosmos tx since we have signCosmosTx as false
	tx := suite.CreateTxBuilder(privKey, txArgs).GetTx()
	sigs, err := tx.GetSignaturesV2()
	suite.Require().NoError(err)

	// signatures of cosmos tx should be empty
	suite.Require().Equal(len(sigs), 0)

	msg := tx.GetMsgs()[0]
	msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
	suite.Require().True(ok)
	txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
	suite.Require().NoError(err)

	msgV, msgR, msgS := txData.GetRawSignatureValues()

	ethTx := msgEthTx.AsTransaction()
	ethV, ethR, ethS := ethTx.RawSignatureValues()

	// The signatures of MsgehtereumTx should be the same with the corresponding eth tx
	suite.Require().Equal(msgV, ethV)
	suite.Require().Equal(msgR, ethR)
	suite.Require().Equal(msgS, ethS)
}
