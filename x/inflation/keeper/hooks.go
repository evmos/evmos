package keeper

import (
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/evmos/evmos/v6/x/epochs/types"
	"github.com/evmos/evmos/v6/x/inflation/types"
)

// BeforeEpochStart: noop, We don't need to do anything here
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
}

// AfterEpochEnd mints and allocates coins at the end of each epoch end
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	skippedEpochs := k.GetSkippedEpochs(ctx)

	// Skip inflation if it is disabled and increment number of skipped epochs
	if !params.EnableInflation {
		// check if the epochIdentifier is "day" before incrementing.
		if epochIdentifier != epochstypes.DayEpochID {
			return
		}
		skippedEpochs++

		k.SetSkippedEpochs(ctx, skippedEpochs)
		k.Logger(ctx).Debug(
			"skipping inflation mint and allocation",
			"height", ctx.BlockHeight(),
			"epoch-id", epochIdentifier,
			"epoch-number", epochNumber,
			"skipped-epochs", skippedEpochs,
		)
		return
	}

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
	staking, incentives, communityPool, err := k.MintAndAllocateInflation(ctx, mintedCoin)
	if err != nil {
		panic(err)
	}

	period := k.GetPeriod(ctx)
	epochsPerPeriod := k.GetEpochsPerPeriod(ctx)
	newProvision := epochMintProvision

	// If period is passed, update the period and epochMintProvision. A period is
	// passed if the current epoch number surpasses the epochsPerPeriod for the
	// current period. Skipped epochs are subtracted to only account for epochs
	// where inflation minted tokens.
	//
	// Examples:
	// Given, epochNumber = 1, period = 0, epochPerPeriod = 365, skippedEpochs = 0
	//   => 1 - 365 * 0 - 0 < 365 --- nothing to do here
	// Given, epochNumber = 741, period = 1, epochPerPeriod = 365, skippedEpochs = 10
	//   => 741 - 1 * 365 - 10 > 365 --- a period has passed! we change the epochMintProvision and set a new period
	if epochNumber-epochsPerPeriod*int64(period)-int64(skippedEpochs) > epochsPerPeriod {
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

	defer func() {
		if mintedCoin.Amount.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "total"},
				float32(mintedCoin.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}
		if staking.AmountOf(mintedCoin.Denom).IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "staking", "total"},
				float32(staking.AmountOf(mintedCoin.Denom).Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}
		if incentives.AmountOf(mintedCoin.Denom).IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "incentives", "total"},
				float32(incentives.AmountOf(mintedCoin.Denom).Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}
		if communityPool.AmountOf(mintedCoin.Denom).IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "community_pool", "total"},
				float32(communityPool.AmountOf(mintedCoin.Denom).Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}
	}()

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
