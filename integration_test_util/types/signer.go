package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
)

var _ keyring.Signer = &signer{}

// signer defines a type that is used on testing for signing MsgEthereumTx
type signer struct {
	privKey cryptotypes.PrivKey
}

func NewSigner(sk cryptotypes.PrivKey) keyring.Signer {
	if sk.Type() != ethsecp256k1.KeyType {
		panic(fmt.Sprintf(
			"require key type %s, got %s",
			ethsecp256k1.KeyType,
			sk.Type(),
		))
	}

	return &signer{
		privKey: sk,
	}
}

// Sign signs the message using the underlying private key
func (s signer) Sign(_ string, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	sig, err := s.privKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, s.privKey.PubKey(), nil
}

// SignByAddress sign byte messages with a user key providing the address.
func (s signer) SignByAddress(address sdk.Address, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	signerAddress := sdk.AccAddress(s.privKey.PubKey().Address())
	if !signerAddress.Equals(address) {
		return nil, nil, fmt.Errorf("address mismatch: signer %s ≠ given address %s", signerAddress, address)
	}

	return s.Sign("", msg)
}
