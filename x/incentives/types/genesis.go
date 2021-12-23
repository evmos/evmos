package types

import "fmt"

func NewGenesisState(
	params Params,
	incentives []Incentive,
	gasMeters []GasMeter,
) GenesisState {
	return GenesisState{
		Params:     params,
		Incentives: incentives,
		GasMeters:  gasMeters,
	}
}

func (gs GenesisState) Validate() error {
	seenContractIn := make(map[string]bool)

	for _, in := range gs.Incentives {
		// only one incentive per contract
		if seenContractIn[in.Contract] {
			return fmt.Errorf("contract duplicated on genesis '%s'", in.Contract)
		}

		if err := in.Validate(); err != nil {
			return err
		}

		seenContractIn[in.Contract] = true
	}

	seenContractGm := make(map[string]bool)
	seenParticipant := make(map[string]bool)

	for _, gm := range gs.GasMeters {
		// only 1 participant per gas meter
		if seenContractGm[gm.Contract] && seenParticipant[gm.Participant] {
			return fmt.Errorf(
				"gas meter duplicated on genesis contract: '%s',  participant: '%s'",
				gm.Contract, gm.Participant,
			)
		}

		if err := gm.Validate(); err != nil {
			return err
		}

		seenContractGm[gm.Contract] = true
		seenParticipant[gm.Participant] = true
	}

	return gs.Params.Validate()
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}
