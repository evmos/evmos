package v14rc2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingkeeper "github.com/evmos/evmos/v13/x/vesting/keeper"
	vestingtypes "github.com/evmos/evmos/v13/x/vesting/types"
)

// UpdateVestingFunders updates the vesting funders for accounts managed by the team
// to the new dedicated multisig address.
func UpdateVestingFunders(ctx sdk.Context, vk vestingkeeper.Keeper) error {
	// Update account created by funder 1
	if _, err := UpdateVestingFunder(ctx, vk, VestingAddrByFunder1, OldFunder1); err != nil {
		return err
	}

	// Update accounts created by funder 2
	for _, address := range VestingAddrsByFunder2 {
		if _, err := UpdateVestingFunder(ctx, vk, address, OldFunder2); err != nil {
			return err
		}
	}
	return nil
}

// UpdateVestingFunder updates the vesting funder for a single vesting account when address and the previous funder
// are given as strings.
func UpdateVestingFunder(ctx sdk.Context, k vestingkeeper.Keeper, address, oldFunder string) (*vestingtypes.MsgUpdateVestingFunderResponse, error) {
	vestingAcc := sdk.MustAccAddressFromBech32(address)
	oldFunderAcc := sdk.MustAccAddressFromBech32(oldFunder)
	msgUpdate := vestingtypes.NewMsgUpdateVestingFunder(oldFunderAcc, NewTeamPremintWalletAcc, vestingAcc)

	return k.UpdateVestingFunder(ctx, msgUpdate)
}
