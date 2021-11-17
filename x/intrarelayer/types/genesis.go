package types

import "fmt"

func NewGenesisState(params Params, pairs []TokenPair) GenesisState {
	return GenesisState{
		Params:     params,
		TokenPairs: pairs,
	}
}

func (gs GenesisState) Validate() error {
	seenErc20 := make(map[string]bool)
	seenDenom := make(map[string]bool)

	for _, b := range gs.TokenPairs {
		if seenErc20[b.Erc20Address] {
			return fmt.Errorf("token ERC20 contract duplicated on genesis '%s'", b.Erc20Address)
		}
		if seenDenom[b.Denom] {
			return fmt.Errorf("coin denomination duplicated on genesis: '%s'", b.Denom)
		}

		if err := b.Validate(); err != nil {
			return err
		}

		seenErc20[b.Erc20Address] = true
		seenDenom[b.Denom] = true
	}

	return gs.Params.Validate()
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}
