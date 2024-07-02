package types

import "fmt"

var ParamsKey = []byte("Params")

func NewParams(
	auctionEnabled bool,
) Params {
	return Params{
		EnableAuction: auctionEnabled,
	}
}

func DefaultParams() Params {
	return Params{
		EnableAuction: true,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	return validateBool(p.EnableAuction)
}
