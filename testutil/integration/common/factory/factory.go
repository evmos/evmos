// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package factory

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"

	errorsmod "cosmossdk.io/errors"
)

const (
	GasAdjustment = float64(1.7)
)

// TxFactory is the interface that wraps the common methods to build and broadcast transactions
// within cosmos chains
type TxFactory interface {
	// BuildCosmosTx builds a Cosmos tx with the provided private key and txArgs
	BuildCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (signing.Tx, error)
	// ExecuteCosmosTx builds, signs and broadcasts a Cosmos tx with the provided private key and txArgs
	ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error)
}

var _ TxFactory = (*IntegrationTxFactory)(nil)

// IntegrationTxFactory is a helper struct to build and broadcast transactions
// to the network on integration tests. This is to simulate the behavior of a real user.
type IntegrationTxFactory struct {
	grpcHandler grpc.Handler
	network     network.Network
	ec          *testutiltypes.TestEncodingConfig
}

// New creates a new IntegrationTxFactory instance
func New(
	network network.Network,
	grpcHandler grpc.Handler,
	ec *testutiltypes.TestEncodingConfig,
) *IntegrationTxFactory {
	return &IntegrationTxFactory{
		grpcHandler: grpcHandler,
		network:     network,
		ec:          ec,
	}
}

func (tf *IntegrationTxFactory) BuildCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (signing.Tx, error) {
	txBuilder, err := tf.buildTx(privKey, txArgs)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to build tx")
	}
	return txBuilder.GetTx(), nil
}

// ExecuteCosmosTx creates, signs and broadcasts a Cosmos transaction
func (tf *IntegrationTxFactory) ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error) {
	signedTx, err := tf.BuildCosmosTx(privKey, txArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to generate tx")
	}

	txBytes, err := tf.encodeTx(signedTx)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to encode tx")
	}

	return tf.network.BroadcastTxSync(txBytes)
}
