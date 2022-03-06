package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/tharsis/evmos/v2/x/epochs/types"
	"github.com/tharsis/evmos/v2/x/inflation/types"
)

// BeforeEpochStart: noop, We don't need to do anything here
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
}

// AfterEpochEnd mints and distributes coins at the end of each epoch end
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	expEpochID := k.GetEpochIdentifier(ctx)
	if epochIdentifier != expEpochID {
		return
	}

	// mint coins, update supply
	epochMintProvision, found := k.GetEpochMintProvision(ctx)
	if !found {
		panic("the epochMintProvision was not found")
	}

	mintedCoin := sdk.NewCoin(params.MintDenom, epochMintProvision.TruncateInt())
	if err := k.MintAndAllocateInflation(ctx, mintedCoin); err != nil {
		panic(err)
	}

	// check if a period is over. If it's completed, update period, and epochMintProvision
	period := k.GetPeriod(ctx)
	epochsPerPeriod := k.GetEpochsPerPeriod(ctx)

	newProvision := epochMintProvision

	// current epoch number needs to be within range for the period
	// Given, epochNumber = 1, period = 0, epochPerPeriod = 365
	// 1 - 365 * 0 < 365 --- nothing to do here
	// Given, epochNumber = 731, period = 1, epochPerPeriod = 365
	// 731 - 1 * 365 > 365 --- a period has passed! we change the epochMintProvision and set a new period
	if epochNumber-epochsPerPeriod*int64(period) > epochsPerPeriod {
		period++
		k.SetPeriod(ctx, period)
		period = k.GetPeriod(ctx)
		bondedRatio := k.BondedRatio(ctx)
		newProvision = types.CalculateEpochMintProvision(
			params,
			period,
			epochsPerPeriod,
			bondedRatio,
		)
		k.SetEpochMintProvision(ctx, newProvision)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
			sdk.NewAttribute(types.AttributeKeyEpochProvisions, newProvision.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
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
