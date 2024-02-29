package v17_test

import banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

func CreateFullMetadata(denom, symbol, name string) banktypes.Metadata {
	return banktypes.Metadata{
		Description: "desc",
		Base:        denom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: 0,
			},
			{
				Denom:    symbol,
				Exponent: uint32(18),
			},
		},
		Name:    name,
		Symbol:  symbol,
		Display: denom,
	}
}
