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
		if seenContractGm[gm.Contract] && seenParticipant[gm.Participant] {
			return fmt.Errorf(
				"gasmeter duplicated on genesis cotract: '%s',  participant: '%s'",
				gm.Contract,
				gm.Participant,
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
