// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/evmos/evmos/v15/utils"
)

var (
	// DefaultClaimsDenom is aevmos
	DefaultClaimsDenom = utils.BaseDenom
	// DefaultDurationUntilDecay is 1 month = 30.4375 days
	DefaultDurationUntilDecay = 2629800 * time.Second
	// DefaultDurationOfDecay is 2 months
	DefaultDurationOfDecay = 2 * DefaultDurationUntilDecay
	// DefaultAuthorizedChannels defines the list of default IBC authorized channels that can perform
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

// ParamsKey store key for params
var ParamsKey = []byte("Params")

// NewParams creates a new Params object
func NewParams(
	enableClaim bool,
	claimsDenom string,
	airdropStartTime time.Time,
	durationUntilDecay,
	durationOfDecay time.Duration,
	authorizedChannels,
	evmChannels []string,
) Params {
	return Params{
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
func DefaultParams() Params {
	return Params{
		EnableClaims:       DefaultEnableClaims,
		ClaimsDenom:        DefaultClaimsDenom,
		AirdropStartTime:   DefaultAirdropStartTime,
		DurationUntilDecay: DefaultDurationUntilDecay,
		DurationOfDecay:    DefaultDurationOfDecay,
		AuthorizedChannels: DefaultAuthorizedChannels,
		EVMChannels:        DefaultEVMChannels,
	}
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

func (p Params) Validate() error {
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
func (p Params) DecayStartTime() time.Time {
	return p.AirdropStartTime.Add(p.DurationUntilDecay)
}

// AirdropEndTime returns the time at which no further claims will be processed.
func (p Params) AirdropEndTime() time.Time {
	return p.AirdropStartTime.Add(p.DurationUntilDecay).Add(p.DurationOfDecay)
}

// IsClaimsActive returns true if the claiming process is active:
// - claims are enabled AND
// - block time is equal or after the airdrop start time AND
// - block time is before or equal the airdrop end time
func (p Params) IsClaimsActive(blockTime time.Time) bool {
	if !p.EnableClaims || blockTime.Before(p.AirdropStartTime) || blockTime.After(p.AirdropEndTime()) {
		return false
	}
	return true
}

// IsAuthorizedChannel returns true if the channel provided is in the list of
// authorized channels
func (p Params) IsAuthorizedChannel(channel string) bool {
	for _, authorizedChannel := range p.AuthorizedChannels {
		if channel == authorizedChannel {
			return true
		}
	}

	return false
}

// IsEVMChannel returns true if the channel provided is in the list of
// EVM channels
func (p Params) IsEVMChannel(channel string) bool {
	for _, evmChannel := range p.EVMChannels {
		if channel == evmChannel {
			return true
		}
	}

	return false
}
