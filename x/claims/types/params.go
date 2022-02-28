package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultClaimsDenom        = "aevmos"
	DefaultDurationUntilDecay = 2629800 * time.Second         // 1 month = 30.4375 days
	DefaultDurationOfDecay    = 2 * DefaultDurationUntilDecay // 2 months
)

// Parameter store key
var (
	ParamStoreKeyEnableClaims       = []byte("EnableClaims")
	ParamStoreKeyAirdropStartTime   = []byte("AirdropStartTime")
	ParamStoreKeyDurationUntilDecay = []byte("DurationUntilDecay")
	ParamStoreKeyDurationOfDecay    = []byte("DurationOfDecay")
	ParamStoreKeyClaimsDenom        = []byte("ClaimsDenom")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableClaims, &p.EnableClaims, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyAirdropStartTime, &p.AirdropStartTime, validateStartDate),
		paramtypes.NewParamSetPair(ParamStoreKeyDurationUntilDecay, &p.DurationUntilDecay, validateDuration),
		paramtypes.NewParamSetPair(ParamStoreKeyDurationOfDecay, &p.DurationOfDecay, validateDuration),
		paramtypes.NewParamSetPair(ParamStoreKeyClaimsDenom, &p.ClaimsDenom, validateDenom),
	}
}

// NewParams creates a new Params object
func NewParams(
	enableClaim bool,
	airdropStartTime time.Time,
	claimsDenom string,
	durationUntilDecay,
	durationOfDecay time.Duration,
) Params {
	return Params{
		EnableClaims:       enableClaim,
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: durationUntilDecay,
		DurationOfDecay:    durationOfDecay,
		ClaimsDenom:        claimsDenom,
	}
}

// DefaultParams creates a parameter instance with default values
// for the claims module.
func DefaultParams() Params {
	return Params{
		EnableClaims:       true,
		AirdropStartTime:   time.Time{},
		DurationUntilDecay: DefaultDurationUntilDecay,
		DurationOfDecay:    DefaultDurationOfDecay,
		ClaimsDenom:        DefaultClaimsDenom,
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

func (p Params) Validate() error {
	if p.DurationOfDecay <= 0 {
		return fmt.Errorf("duration of decay must be positive: %d", p.DurationOfDecay)
	}
	if p.DurationUntilDecay <= 0 {
		return fmt.Errorf("duration until decay must be positive: %d", p.DurationOfDecay)
	}
	return sdk.ValidateDenom(p.ClaimsDenom)
}

// DecayStartTime returns the time at which the Decay period starts
func (p Params) DecayStartTime() time.Time {
	return p.AirdropStartTime.Add(p.DurationUntilDecay)
}

// AirdropEndTime returns the time at which no further claims will be processed.
func (p Params) AirdropEndTime() time.Time {
	return p.AirdropStartTime.Add(p.DurationUntilDecay).Add(p.DurationOfDecay)
}

// IsClaimsActive returns true if the claiming process is active, i.e block time
// is after the airdrop start time and claims are enabled.
func (p Params) IsClaimsActive(blockTime time.Time) bool {
	if !p.EnableClaims || blockTime.Before(p.AirdropStartTime) {
		return false
	}
	return true
}
