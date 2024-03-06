package factory

import (
	errorsmod "cosmossdk.io/errors"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
)

// BaseTxFactory is the interface that wraps the common methods to build and broadcast transactions
// within cosmos chains
type BaseTxFactory interface {
	// BuildCosmosTx builds a Cosmos tx with the provided private key and txArgs
	BuildCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (authsigning.Tx, error)
	// SignCosmosTx signs a Cosmos transaction with the provided
	// private key and tx builder
	SignCosmosTx(privKey cryptotypes.PrivKey, txBuilder client.TxBuilder) error
	// ExecuteCosmosTx builds, signs and broadcasts a Cosmos tx with the provided private key and txArgs
	ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ExecTxResult, error)
	// EncodeTx encodes the provided transaction
	EncodeTx(tx sdktypes.Tx) ([]byte, error)
}

// baseTxFactory is the struct of the basic tx factory
// to build and broadcast transactions.
// This is to simulate the behavior of a real user.
type baseTxFactory struct {
	grpcHandler grpc.Handler
	network     network.Network
	ec          *testutiltypes.TestEncodingConfig
}

// newBaseTxFactory instantiates a new baseTxFactory
func newBaseTxFactory(
	network network.Network,
	grpcHandler grpc.Handler,
	ec *testutiltypes.TestEncodingConfig,
) BaseTxFactory {
	return &baseTxFactory{
		grpcHandler: grpcHandler,
		network:     network,
		ec:          ec,
	}
}

func (tf *baseTxFactory) BuildCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (authsigning.Tx, error) {
	txBuilder, err := tf.buildTx(privKey, txArgs)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to build tx")
	}
	return txBuilder.GetTx(), nil
}

// ExecuteCosmosTx creates, signs and broadcasts a Cosmos transaction
func (tf *baseTxFactory) ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ExecTxResult, error) {
	signedTx, err := tf.BuildCosmosTx(privKey, txArgs)
	if err != nil {
		return abcitypes.ExecTxResult{}, errorsmod.Wrap(err, "failed to build tx")
	}

	txBytes, err := tf.EncodeTx(signedTx)
	if err != nil {
		return abcitypes.ExecTxResult{}, errorsmod.Wrap(err, "failed to encode tx")
	}

	return tf.network.BroadcastTxSync(txBytes)
}

// SignCosmosTx is a helper function that signs a Cosmos transaction
// with the provided private key and transaction builder
func (tf *baseTxFactory) SignCosmosTx(privKey cryptotypes.PrivKey, txBuilder client.TxBuilder) error {
	txConfig := tf.ec.TxConfig
	signMode, err := authsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return errorsmod.Wrap(err, "invalid sign mode")
	}
	signerData, err := tf.setSignatures(privKey, txBuilder, signMode)
	if err != nil {
		return errorsmod.Wrap(err, "failed to set tx signatures")
	}

	return tf.signWithPrivKey(privKey, txBuilder, signerData, signMode)
}
