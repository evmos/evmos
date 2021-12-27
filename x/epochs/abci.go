package epochs

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/epochs/keeper"
	"github.com/tharsis/evmos/x/epochs/types"
)

// BeginBlocker of epochs module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		logger := k.Logger(ctx)

		// Has it not started, and is the block time > initial epoch start time
		shouldInitialEpochStart := !epochInfo.EpochCountingStarted && !epochInfo.StartTime.After(ctx.BlockTime())

		epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		shouldEpochStart := ctx.BlockTime().After(epochEndTime) && !shouldInitialEpochStart && !epochInfo.StartTime.After(ctx.BlockTime())

		if shouldInitialEpochStart || shouldEpochStart {
			epochInfo.CurrentEpochStartHeight = ctx.BlockHeight()

			if shouldInitialEpochStart {
				epochInfo.EpochCountingStarted = true
				epochInfo.CurrentEpoch = 1
				epochInfo.CurrentEpochStartTime = epochInfo.StartTime
				logger.Info("starting epoch", "identifier", epochInfo.Identifier)
			} else {
				epochInfo.CurrentEpoch++
				epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
				logger.Info("starting epoch", "identifier", epochInfo.Identifier)
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeEpochEnd,
						sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
					),
				)
				k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
			}
			k.SetEpochInfo(ctx, epochInfo)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEpochStart,
					sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
					sdk.NewAttribute(types.AttributeEpochStartTime, fmt.Sprintf("%d", epochInfo.CurrentEpochStartTime.Unix())),
				),
			)
			k.BeforeEpochStart(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
		}

		return false
	})
}
