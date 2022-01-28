package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/tharsis/evmos/x/epochs/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)

	// mint coins, update supply
	epochMintProvision, found := k.GetEpochMintProvision(ctx)
	if !found {
		panic("the epochMintProvision was not found")
	}

	mintedCoin := sdk.NewCoin(params.MintDenom, epochMintProvision.TruncateInt())
	if err := k.MintAndAllocateInflation(ctx, mintedCoin); err != nil {
		panic(err)
	}

	// check if a period is over
	period := k.GetPeriod(ctx)
	epochsPerPeriod := k.GetEpochsPerPeriod(ctx)
	if epochNumber-epochsPerPeriod*int64(period) > epochsPerPeriod {
		period++
		k.SetPeriod(ctx, period)
		newProvision := types.CalculateEpochMintProvision(params, k.GetPeriod(ctx), epochsPerPeriod)
		k.SetEpochMintProvision(ctx, newProvision)
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
			// sdk.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
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
