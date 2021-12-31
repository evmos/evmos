package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ethermint "github.com/tharsis/ethermint/types"
)

var (
	DefaultClaimDenom         = ethermint.AttoPhoton
	DefaultDurationUntilDecay = time.Hour
	DefaultDurationOfDecay    = time.Hour * 5
)

// Parameter store key
var (
	ParamStoreKeyEnableClaim        = []byte("EnableClaim")
	ParamStoreKeyAirdropStartTime   = []byte("AirdropStartTime")
	ParamStoreKeyDurationUntilDecay = []byte("DurationUntilDecay")
	ParamStoreKeyDurationOfDecay    = []byte("DurationOfDecay")
	ParamStoreKeyClaimDenom         = []byte("ClaimDenom")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableClaim, &p.EnableClaim, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyAirdropStartTime, &p.AirdropStartTime, validateStartDate),
		paramtypes.NewParamSetPair(ParamStoreKeyDurationUntilDecay, &p.DurationUntilDecay, validateDuration),
		paramtypes.NewParamSetPair(ParamStoreKeyDurationOfDecay, &p.DurationOfDecay, validateDuration),
		paramtypes.NewParamSetPair(ParamStoreKeyClaimDenom, &p.ClaimDenom, validateDenom),
	}
}

// NewParams creates a new Params object
func NewParams(
	enableClaim bool,
	airdropStartTime time.Time,
	claimDenom string,
	durationUntilDecay,
	durationOfDecay time.Duration,
) Params {
	return Params{
		EnableClaim:        enableClaim,
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: durationUntilDecay,
		DurationOfDecay:    durationOfDecay,
		ClaimDenom:         claimDenom,
	}
}

func DefaultParams(airdropStartTime time.Time) Params {
	return Params{
		EnableClaim:        true,
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: DefaultDurationUntilDecay, // 2 month
		DurationOfDecay:    DefaultDurationOfDecay,    // 4 months
		ClaimDenom:         DefaultClaimDenom,         // aphoton
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
	v, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsZero() || v.UnixNano() == 0 {
		return fmt.Errorf("start date cannot be zero: %s", v)
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
	if p.AirdropStartTime.IsZero() || p.AirdropStartTime.UnixNano() == 0 {
		return fmt.Errorf("airdrop start date cannot be zero: %s", p.AirdropStartTime)
	}
	if p.DurationOfDecay <= 0 {
		return fmt.Errorf("duration of decay must be positive: %d", p.DurationOfDecay)
	}
	if p.DurationUntilDecay <= 0 {
		return fmt.Errorf("duration until decay must be positive: %d", p.DurationOfDecay)
	}
	return sdk.ValidateDenom(p.ClaimDenom)
}
