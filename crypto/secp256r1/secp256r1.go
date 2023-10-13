// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package secp256r1

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	tmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// // PrivKeySize defines the size of the PrivKey bytes
	// PrivKeySize = 66

	// KeyType is the string constant for the Secp256r1 algorithm
	KeyType = "secp256r1"
)

// Amino encoding names
const (
	// PrivKeyName defines the amino encoding name for the Secp256r1 private key
	PrivKeyName = "evmos/PrivKeySecp256r1"
	// PubKeyName defines the amino encoding name for the Secp256r1 public key
	PubKeyName = "evmos/PubKeyEthSecp256r1"
)

// ----------------------------------------------------------------------------
// secp256r1 Private Key

var (
	_ cryptotypes.PrivKey  = &PrivKey{}
	_ codec.AminoMarshaler = &PrivKey{}
)

// GenerateKey generates a new random, uncompressed, P256 private key. It returns an error upon
// failure.
func GenerateKey() (*PrivKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	bz, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	return &PrivKey{
		Key: bz,
	}, nil
}

// Bytes returns the byte representation of the P256 Private Key.
func (privKey PrivKey) Bytes() []byte {
	bz := make([]byte, len(privKey.Key))
	copy(bz, privKey.Key)

	return bz
}

// PubKey returns the P256 private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() cryptotypes.PubKey {
	// priv, err := x509.ParseECPrivateKey(privKey.Key)
	// if err != nil {
	// 	return nil
	// }

	return &PubKey{}
}

// Equals returns true if two P256 private keys are equal and false otherwise.
func (privKey PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	privA, err := ecdh.P256().NewPrivateKey(privKey.Key)
	if err != nil {
		return false
	}

	privB, err := ecdh.P256().NewPrivateKey(other.Bytes())
	if err != nil {
		return false
	}

	return privA.Equal(privB)
}

// Type returns eth_secp256k1
func (privKey PrivKey) Type() string {
	return KeyType
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Bytes(), nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	_, err := ecdh.P256().NewPrivateKey(bz)
	if err != nil {
		return err
	}

	privKey.Key = bz
	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// Sign creates a recoverable ECDSA signature on the secp256k1 curve over the
// provided hash of the message. The produced signature is 65 bytes
// where the last byte contains the recovery ID.
func (privKey PrivKey) Sign(digestBz []byte) ([]byte, error) {
	priv, err := x509.ParseECPrivateKey(privKey.Key)
	if err != nil {
		return nil, err
	}

	return ecdsa.SignASN1(rand.Reader, priv, digestBz)
}

// ----------------------------------------------------------------------------
// secp256r1 Public Key

var (
	_ cryptotypes.PubKey   = &PubKey{}
	_ codec.AminoMarshaler = &PubKey{}
)

// Address returns the address of the P256 public key.
// The function will return an empty address if the public key is invalid.
func (pubKey PubKey) Address() tmcrypto.Address {
	return tmcrypto.Address(common.BytesToAddress(crypto.Keccak256(pubKey.Key[1:])[12:]).Bytes())
}

// Bytes returns the raw bytes of the ECDSA public key.
func (pubKey PubKey) Bytes() []byte {
	bz := make([]byte, len(pubKey.Key))
	copy(bz, pubKey.Key)

	return bz
}

// String implements the fmt.Stringer interface.
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeySecp256r1{%X}", pubKey.Key)
}

// Type returns secp256r1
func (pubKey PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}

// VerifySignature verifies that the ECDSA public key created a given signature over
// the provided message.
//
// CONTRACT: The signature should be in [R || S] format.
func (pubKey PubKey) VerifySignature(hash, sig []byte) bool {
	// Parse the public key bytes
	block, _ := pem.Decode(pubKey.Key)
	if block == nil {
		return false
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	// Verify if the parsed public key is an ECDSA public key
	pub, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return false
	}

	return ecdsa.VerifyASN1(pub, hash, sig)
}
