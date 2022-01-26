package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/tharsis/evmos/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// TODO daily epoch logic
	// return
	// // check if epochIdentifier signal equals the identifier in the params
	// if epochIdentifier != params.EpochIdentifier {
	// 	return
	// }

	// // not distribute rewards if it's not time yet for rewards allocation
	// if epochNumber < params.MintingRewardsAllocationStartEpoch {
	// 	return
	// } else if epochNumber == params.MintingRewardsAllocationStartEpoch {
	// 	k.SetLastHalvenEpochNum(ctx, epochNumber)
	// }
	// // fetch stored minter & params
	// minter := k.GetMinter(ctx)
	// // params := k.GetParams(ctx)

	// // Check if we have hit an epoch where we update the inflation parameter.
	// // Since epochs only update based on BFT time data, it is safe to store the
	// // "halvening period time" in terms of the number of epochs that have
	// // transpired.
	// if epochNumber >= k.GetParams(ctx).ReductionPeriodInEpochs+k.GetLastHalvenEpochNum(ctx) {
	// 	// Halven the reward per halven period
	// 	minter.EpochProvisions = minter.NextEpochProvisions(params)
	// 	k.SetMinter(ctx, minter)
	// 	k.SetLastHalvenEpochNum(ctx, epochNumber)
	// }

	// mint coins, update supply
	// epochMintProvision, found := k.GetEpochMintProvision(ctx)
	// if found {
	// 	panic("the epochMintProvision has was not found")
	// }

	// mintedCoin := sdk.NewCoin(params.MintDenom, epochMintProvision.TruncateInt())
	// // We over-allocate by the developer vesting portion, and burn this later
	// if err := k.MintAndAllocateInflation(ctx, mintedCoin); err != nil {
	// 	panic(err)
	// }

	// if mintedCoin.Amount.IsInt64() {
	// 	defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	// }

	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		types.EventTypeMint,
	// 		sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
	// 		// sdk.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
	// 		sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
	// 	),
	// )
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
