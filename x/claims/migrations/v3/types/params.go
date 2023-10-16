// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
	"time"

	"github.com/evmos/evmos/v15/utils"
	"github.com/evmos/evmos/v15/x/claims/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
)

var _ types.LegacyParams = &V3Params{}

var (
	// DefaultClaimsDenom is aevmos
	DefaultClaimsDenom = utils.BaseDenom
	// DefaultDurationUntilDecay is 1 month = 30.4375 days
	DefaultDurationUntilDecay = 2629800 * time.Second
	// DefaultDurationOfDecay is 2 months
	DefaultDurationOfDecay = 2 * DefaultDurationUntilDecay
	// DefaultAuthorizedChannels  defines the list of default IBC authorized channels that can perform
	// IBC address attestations in order to migrate claimable amounts. By default
	// only Osmosis and Cosmos Hub channels are authorized
	DefaultAuthorizedChannels = []string{
		"channel-0", // Osmosis
		"channel-3", // Cosmos Hub
	}
	DefaultEVMChannels = []string{
		"channel-2", // Injective
	}
	DefaultEnableClaims     = true
	DefaultAirdropStartTime = time.Time{}
)

// Parameter store key
var (
	ParamsKey                       = []byte("Params")
	ParamStoreKeyEnableClaims       = []byte("EnableClaims")
	ParamStoreKeyAirdropStartTime   = []byte("AirdropStartTime")
	ParamStoreKeyDurationUntilDecay = []byte("DurationUntilDecay")
	ParamStoreKeyDurationOfDecay    = []byte("DurationOfDecay")
	ParamStoreKeyClaimsDenom        = []byte("ClaimsDenom")
	ParamStoreKeyAuthorizedChannels = []byte("AuthorizedChannels")
	ParamStoreKeyEVMChannels        = []byte("EVMChannels")
)

var _ paramtypes.ParamSet = &V3Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&V3Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *V3Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableClaims, &p.EnableClaims, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyAirdropStartTime, &p.AirdropStartTime, validateStartDate),
		paramtypes.NewParamSetPair(ParamStoreKeyDurationUntilDecay, &p.DurationUntilDecay, validateDuration),
		paramtypes.NewParamSetPair(ParamStoreKeyDurationOfDecay, &p.DurationOfDecay, validateDuration),
		paramtypes.NewParamSetPair(ParamStoreKeyClaimsDenom, &p.ClaimsDenom, validateDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyAuthorizedChannels, &p.AuthorizedChannels, ValidateChannels),
		paramtypes.NewParamSetPair(ParamStoreKeyEVMChannels, &p.EVMChannels, ValidateChannels),
	}
}

// NewParams creates a new Params object
func NewParams(
	enableClaim bool,
	claimsDenom string,
	airdropStartTime time.Time,
	durationUntilDecay,
	durationOfDecay time.Duration,
	authorizedChannels,
	evmChannels []string,
) V3Params {
	return V3Params{
		EnableClaims:       enableClaim,
		ClaimsDenom:        claimsDenom,
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: durationUntilDecay,
		DurationOfDecay:    durationOfDecay,
		AuthorizedChannels: authorizedChannels,
		EVMChannels:        evmChannels,
	}
}

// DefaultParams creates a parameter instance with default values
// for the claims module.
func DefaultParams() V3Params {
	return V3Params{
		EnableClaims:       DefaultEnableClaims,
		ClaimsDenom:        DefaultClaimsDenom,
		AirdropStartTime:   DefaultAirdropStartTime,
		DurationUntilDecay: DefaultDurationUntilDecay,
		DurationOfDecay:    DefaultDurationOfDecay,
		AuthorizedChannels: DefaultAuthorizedChannels,
		EVMChannels:        DefaultEVMChannels,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateStartDate(i interface{}) error {
	_, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("duration must be positive: %s", v)
	}

	return nil
}

func validateDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

// ValidateChannels checks if channels ids are valid
func ValidateChannels(i interface{}) error {
	channels, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, channel := range channels {
		if err := host.ChannelIdentifierValidator(channel); err != nil {
			return errorsmod.Wrap(
				channeltypes.ErrInvalidChannelIdentifier, err.Error(),
			)
		}
	}

	return nil
}

func (p V3Params) Validate() error {
	if p.DurationOfDecay <= 0 {
		return fmt.Errorf("duration of decay must be positive: %d", p.DurationOfDecay)
	}
	if p.DurationUntilDecay <= 0 {
		return fmt.Errorf("duration until decay must be positive: %d", p.DurationOfDecay)
	}
	if err := sdk.ValidateDenom(p.ClaimsDenom); err != nil {
		return err
	}
	if err := ValidateChannels(p.AuthorizedChannels); err != nil {
		return err
	}
	return ValidateChannels(p.EVMChannels)
}

// DecayStartTime returns the time at which the Decay period starts
func (p V3Params) DecayStartTime() time.Time {
	return p.AirdropStartTime.Add(p.DurationUntilDecay)
}

// AirdropEndTime returns the time at which no further claims will be processed.
func (p V3Params) AirdropEndTime() time.Time {
	return p.AirdropStartTime.Add(p.DurationUntilDecay).Add(p.DurationOfDecay)
}

// IsClaimsActive returns true if the claiming process is active:
// - claims are enabled AND
// - block time is equal or after the airdrop start time AND
// - block time is before or equal the airdrop end time
func (p V3Params) IsClaimsActive(blockTime time.Time) bool {
	if !p.EnableClaims || blockTime.Before(p.AirdropStartTime) || blockTime.After(p.AirdropEndTime()) {
		return false
	}
	return true
}

// IsAuthorizedChannel returns true if the channel provided is in the list of
// authorized channels
func (p V3Params) IsAuthorizedChannel(channel string) bool {
	for _, authorizedChannel := range p.AuthorizedChannels {
		if channel == authorizedChannel {
			return true
		}
	}

	return false
}

// IsEVMChannel returns true if the channel provided is in the list of
// EVM channels
func (p V3Params) IsEVMChannel(channel string) bool {
	for _, evmChannel := range p.EVMChannels {
		if channel == evmChannel {
			return true
		}
	}

	return false
}
