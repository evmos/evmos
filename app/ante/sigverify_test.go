package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/app/ante"
	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v18/encoding"
)

func TestConsumeSignatureVerificationGas(t *testing.T) {
	params := authtypes.DefaultParams()
	msg := []byte{1, 2, 3, 4}

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	cdc := encodingConfig.Amino

	p := authtypes.DefaultParams()
	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, msg, false)
	multisigKey1 := kmultisig.NewLegacyAminoPubKey(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)

	for i := 0; i < len(pkSet1); i++ {
		// using nolint:all because the staticcheck nolint is not working as expected
		stdSig := legacytx.StdSignature{PubKey: pkSet1[i], Signature: sigSet1[i]} //nolint:all
		sigV2, err := legacytx.StdSignatureToSignatureV2(cdc, stdSig)
		require.NoError(t, err)
		err = multisig.AddSignatureV2(multisignature1, sigV2, pkSet1)
		require.NoError(t, err)
	}

	ethsecKey, _ := ethsecp256k1.GenerateKey()
	skR1, _ := secp256r1.GenPrivKey()

	type args struct {
		meter  storetypes.GasMeter
		sig    signing.SignatureData
		pubkey cryptotypes.PubKey
		params authtypes.Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		shouldErr   bool
	}{
		{
			"PubKeyEd25519",
			args{sdk.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params},
			p.SigVerifyCostED25519,
			true,
		},
		{
			"PubKeyEthsecp256k1",
			args{sdk.NewInfiniteGasMeter(), nil, ethsecKey.PubKey(), params},
			ante.Secp256k1VerifyCost,
			false,
		},
		{
			"PubKeySecp256k1",
			args{sdk.NewInfiniteGasMeter(), nil, secp256k1.GenPrivKey().PubKey(), params},
			p.SigVerifyCostSecp256k1,
			true,
		},
		{
			"PubKeySecp256r1",
			args{sdk.NewInfiniteGasMeter(), nil, skR1.PubKey(), params},
			p.SigVerifyCostSecp256r1(),
			true,
		},
		{
			"Multisig",
			args{sdk.NewInfiniteGasMeter(), multisignature1, multisigKey1, params},
			expectedCost1,
			false,
		},
		{
			"unknown key",
			args{sdk.NewInfiniteGasMeter(), nil, nil, params},
			0,
			true,
		},
	}
	for _, tt := range tests {
		sigV2 := signing.SignatureV2{
			PubKey:   tt.args.pubkey,
			Data:     tt.args.sig,
			Sequence: 0, // Arbitrary account sequence
		}
		err := ante.SigVerificationGasConsumer(tt.args.meter, sigV2, tt.args.params)

		if tt.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.gasConsumed, tt.args.meter.GasConsumed())
		}
	}
}
