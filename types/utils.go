package types

import (
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
)

// IsSupportedKey returns true if the pubkey type is supported by the chain
// (i.e eth_secp256k1, amino multisig, ed25519)
func IsSupportedKey(pubkey cryptotypes.PubKey) bool {
	switch pubkey.(type) {
	case *ethsecp256k1.PubKey, *ed25519.PubKey, multisig.PubKey:
		return true
	default:
		return false
	}
}
