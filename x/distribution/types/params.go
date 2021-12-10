package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyFeeDistribution       = []byte("FeeDistribution")
	ParamStoreKeyEnableWithdrawAddress = []byte("WithdrawAddress")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableWithdrawAddress bool,
	distribution Distribution,
) Params {
	return Params{
		WithdrawAddrEnabled: enableWithdrawAddress,
		FeeDistribution:     distribution,
	}
}

func DefaultParams() Params {
	return Params{
		WithdrawAddrEnabled: true,
		FeeDistribution: Distribution{
			ProposerReward:  sdk.NewDecWithPrec(5, 1), // 50%,
			ContractRewards: sdk.NewDecWithPrec(5, 1), // 50%,
		},
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDistribution(i interface{}) error {
	distr, ok := i.(Distribution)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return distr.Validate()
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyFeeDistribution, &p.FeeDistribution, validateDistribution),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableWithdrawAddress, &p.WithdrawAddrEnabled, validateBool),
	}
}

func (p Params) Validate() error {
	return p.FeeDistribution.Validate()
}
