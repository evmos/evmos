package types

import (
	fmt "fmt"
	"time"

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
	ParamStoreKeyAirdropStartTime      = []byte("AirdropStartTime")
	ParamStoreKeyEnableErc20           = []byte("EnableErc20")
	ParamStoreKeyTokenPairVotingPeriod = []byte("TokenPairVotingPeriod")
	ParamStoreKeyEnableEVMHook         = []byte("EnableEVMHook")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	airdropStartTime time.Time,
	enableClaim bool,
	claimDenom string,
	durationUntilDecay,
	durationOfDecay time.Duration,
) Params {
	return Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: durationUntilDecay,
		DurationOfDecay:    durationOfDecay,
		ClaimDenom:         claimDenom,
	}
}

func DefaultParams() Params {
	return Params{}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validatePeriod(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("voting period must be positive: %s", v)
	}

	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		// paramtypes.NewParamSetPair(ParamStoreKeyEnableErc20, &p.EnableErc20, validateBool),
		// paramtypes.NewParamSetPair(ParamStoreKeyTokenPairVotingPeriod, &p.TokenPairVotingPeriod, validatePeriod),
		// paramtypes.NewParamSetPair(ParamStoreKeyEnableEVMHook, &p.EnableEVMHook, validateBool),
	}
}

func (p Params) Validate() error {
	// if p.TokenPairVotingPeriod <= 0 {
	// 	return fmt.Errorf("voting period must be positive: %d", p.TokenPairVotingPeriod)
	// }

	return nil
}
