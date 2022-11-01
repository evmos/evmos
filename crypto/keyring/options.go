package keyring

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/evoblockchain/ethermint/crypto/hd"
)

var (
	// SupportedAlgorithms defines the list of signing algorithms used on Evoblock:
	//  - eth_secp256k1 (Ethereum)
	SupportedAlgorithms = keyring.SigningAlgoList{hd.EthSecp256k1}
	// SupportedAlgorithmsLedger defines the list of signing algorithms used on Evoblock for the Ledger device:
	//  - eth_secp256k1 (Ethereum)
	SupportedAlgorithmsLedger = keyring.SigningAlgoList{hd.EthSecp256k1}
)

// EthSecp256k1Option defines a function keys options for the ethereum Secp256k1 curve.
// It supports eth_secp256k1 keys for accounts.
func Option() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = SupportedAlgorithms
		options.SupportedAlgosLedger = SupportedAlgorithmsLedger
	}
}
