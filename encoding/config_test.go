package encoding_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v20/encoding"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func TestTxEncoding(t *testing.T) {
	addr, key := utiltx.NewAddrKey()
	signer := utiltx.NewSigner(key)

	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:   big.NewInt(1),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(1),
		Input:     []byte{},
	}
	msg := evmtypes.NewTx(&ethTxParams)
	msg.From = addr.Hex()

	ethSigner := ethtypes.LatestSignerForChainID(big.NewInt(1))
	err := msg.Sign(ethSigner, signer)
	require.NoError(t, err)

	cfg := encoding.MakeConfig()

	_, err = cfg.TxConfig.TxEncoder()(msg)
	require.Error(t, err, "encoding failed")

	// FIXME: transaction hashing is hardcoded on Tendermint:
	// See https://github.com/cometbft/cometbft/issues/6539 for reference
	// txHash := msg.AsTransaction().Hash()
	// tmTx := cmttypes.Tx(bz)

	// require.Equal(t, txHash.Bytes(), tmTx.Hash())
}
