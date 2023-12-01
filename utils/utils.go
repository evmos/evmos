// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"
	"strings"

	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// MainnetChainID defines the Evmos EIP155 chain ID for mainnet
	MainnetChainID = "evmos_9001"
	// TestnetChainID defines the Evmos EIP155 chain ID for testnet
	TestnetChainID = "evmos_9000"
	// BaseDenom defines the Evmos mainnet denomination
	BaseDenom = "aevmos"
)

// IsMainnet returns true if the chain-id has the Evmos mainnet EIP155 chain prefix.
func IsMainnet(chainID string) bool {
	return strings.HasPrefix(chainID, MainnetChainID)
}

// IsTestnet returns true if the chain-id has the Evmos testnet EIP155 chain prefix.
func IsTestnet(chainID string) bool {
	return strings.HasPrefix(chainID, TestnetChainID)
}

// IsSupportedKey returns true if the pubkey type is supported by the chain
// (i.e eth_secp256k1, amino multisig, ed25519).
// NOTE: Nested multisigs are not supported.
func IsSupportedKey(pubkey cryptotypes.PubKey) bool {
	switch pubkey := pubkey.(type) {
	case *ethsecp256k1.PubKey, *ed25519.PubKey:
		return true
	case multisig.PubKey:
		if len(pubkey.GetPubKeys()) == 0 {
			return false
		}

		for _, pk := range pubkey.GetPubKeys() {
			switch pk.(type) {
			case *ethsecp256k1.PubKey, *ed25519.PubKey:
				continue
			default:
				// Nested multisigs are unsupported
				return false
			}
		}

		return true
	default:
		return false
	}
}

// GetEvmosAddressFromBech32 returns the sdk.Account address of given address,
// while also changing bech32 human readable prefix (HRP) to the value set on
// the global sdk.Config (eg: `evmos`).
// The function fails if the provided bech32 address is invalid.
func GetEvmosAddressFromBech32(address string) (sdk.AccAddress, error) {
	bech32Prefix := strings.SplitN(address, "1", 2)[0]
	if bech32Prefix == address {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid bech32 address: %s", address)
	}

	addressBz, err := sdk.GetFromBech32(address, bech32Prefix)
	if err != nil {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid address %s, %s", address, err.Error())
	}

	// safety check: shouldn't happen
	if err := sdk.VerifyAddressFormat(addressBz); err != nil {
		return nil, err
	}

	return sdk.AccAddress(addressBz), nil
}

// CreateAccAddressFromBech32 creates an AccAddress from a Bech32 string.
func CreateAccAddressFromBech32(address string, bech32prefix string) (addr sdk.AccAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return sdk.AccAddress{}, fmt.Errorf("empty address string is not allowed")
	}

	bz, err := sdk.GetFromBech32(address, bech32prefix)
	if err != nil {
		return nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddress(bz), nil
}

// GetIBCDenomAddress returns the address from the hash of the ICS20's DenomTrace Path.
func GetIBCDenomAddress(denom string) (common.Address, error) {
	if !strings.HasPrefix(denom, "ibc/") {
		return common.Address{}, ibctransfertypes.ErrInvalidDenomForTransfer.Wrapf("coin %s does not have 'ibc/' prefix", denom)
	}

	if len(denom) < 5 || strings.TrimSpace(denom[4:]) == "" {
		return common.Address{}, ibctransfertypes.ErrInvalidDenomForTransfer.Wrapf("coin %s does not a valid IBC voucher hash", denom)
	}

	// Get the address from the hash of the ICS20's DenomTrace Path
	bz, err := ibctransfertypes.ParseHexHash(denom[4:])
	if err != nil {
		return common.Address{}, ibctransfertypes.ErrInvalidDenomForTransfer.Wrap(err.Error())
	}

	return common.BytesToAddress(bz), nil
}

// ComputeIBCDenomTrace compute the ibc voucher denom trace associated with
// the portID, channelID, and the given a token denomination.
func ComputeIBCDenomTrace(
	portID, channelID,
	denom string,
) ibctransfertypes.DenomTrace {
	denomTrace := ibctransfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: denom,
	}

	return denomTrace
}

// ComputeIBCDenom compute the ibc voucher denom associated to
// the portID, channelID, and the given a token denomination.
func ComputeIBCDenom(
	portID, channelID,
	denom string,
) string {
	return ComputeIBCDenomTrace(portID, channelID, denom).IBCDenom()
}
