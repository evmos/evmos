package types

import (
	fmt "fmt"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyEnableIntrarelayer    = []byte("EnableIntrarelayer")
	ParamStoreKeyTokenPairVotingPeriod = []byte("TokenPairVotingPeriod")
	ParamStoreKeyEnableEVMHook         = []byte("EnableEVMHook")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableIntrarelayer bool,
	votingPeriod time.Duration,
	enableEVMHook bool,
) Params {
	return Params{
		EnableIntrarelayer:    enableIntrarelayer,
		TokenPairVotingPeriod: votingPeriod,
		EnableEVMHook:         enableEVMHook,
	}
}

func DefaultParams() Params {
	return Params{
		EnableIntrarelayer:    true,
		TokenPairVotingPeriod: govtypes.DefaultPeriod,
		EnableEVMHook:         true,
	}
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
		paramtypes.NewParamSetPair(ParamStoreKeyEnableIntrarelayer, &p.EnableIntrarelayer, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyTokenPairVotingPeriod, &p.TokenPairVotingPeriod, validatePeriod),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableEVMHook, &p.EnableEVMHook, validateBool),
	}
}

func (p Params) Validate() error {
	if p.TokenPairVotingPeriod <= 0 {
		return fmt.Errorf("voting period must be positive: %d", p.TokenPairVotingPeriod)
	}

	return nil
}
