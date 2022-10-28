package app

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

var _ authante.SignatureVerificationGasConsumer = SigVerificationGasConsumer

const (
	secp256k1VerifyCost uint64 = 21000
)

// SigVerificationGasConsumer is the Evmos implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
// The types of keys supported are:
//
// - ethsecp256k1 (Ethereum keys)
//
// - ed25519 (Validators)
//
// - multisig (Cosmos SDK multisigs)
func SigVerificationGasConsumer(
	meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	pubkey := sig.PubKey
	switch pubkey := pubkey.(type) {

	case *ethsecp256k1.PubKey:
		// Ethereum keys
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil
	case *ed25519.PubKey:
		// Validator keys
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return sdkerrors.Wrap(errortypes.ErrInvalidPubKey, "ED25519 public keys are unsupported")

	case multisig.PubKey:
		// Multisig keys
		multisignature, ok := sig.Data.(*signing.MultiSignatureData)
		if !ok {
			return fmt.Errorf("expected %T, got, %T", &signing.MultiSignatureData{}, sig.Data)
		}
		return ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params, sig.Sequence)

	default:
		return sdkerrors.Wrapf(errortypes.ErrInvalidPubKey, "unrecognized/unsupported public key type: %T", pubkey)
	}
}

// ConsumeMultisignatureVerificationGas consumes gas from a GasMeter for verifying a multisig pubkey signature
func ConsumeMultisignatureVerificationGas(
	meter sdk.GasMeter, sig *signing.MultiSignatureData, pubkey multisig.PubKey,
	params authtypes.Params, accSeq uint64,
) error {
	size := sig.BitArray.Count()
	sigIndex := 0

	for i := 0; i < size; i++ {
		if !sig.BitArray.GetIndex(i) {
			continue
		}
		sigV2 := signing.SignatureV2{
			PubKey:   pubkey.GetPubKeys()[i],
			Data:     sig.Signatures[sigIndex],
			Sequence: accSeq,
		}
		err := SigVerificationGasConsumer(meter, sigV2, params)
		if err != nil {
			return err
		}
		sigIndex++
	}

	return nil
}
