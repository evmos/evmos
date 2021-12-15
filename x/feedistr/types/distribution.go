package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (d Distribution) Validate() error {
	if d.ContractRewards.IsNegative() || d.ContractRewards.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid contract rewards value: %s", d.ProposerReward)
	}
	if d.ProposerReward.IsNegative() || d.ProposerReward.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid proposer reward value: %s", d.ProposerReward)
	}

	total := d.ProposerReward.Add(d.ContractRewards)
	if total.GT(sdk.OneDec()) {
		return fmt.Errorf("contract rewards + proposer reward cannot be > 1: %s", total)
	}
	return nil
}
