package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/incentives/types"
)

// Distribute transfers the allocated rewards to the participants of a given
// incentive.
func (k Keeper) DistributeIncentives(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	// for incentive in incentives:
	//	total_gas = GetTotalGas(incentive)
	// 	for gasmeter in IncentiveGasMeters(incentive)
	//    incentiveShare =
	//		transferRewards

	// Allocation
	var coinsAllocated map[common.Address]sdk.Coins
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			var coins sdk.Coins
			for _, allocation := range incentive.Allocations {
				moduleBalance :=
				coinAllocated := moduleBalance * allocation.Amount / 100
				coin := sdk.NewCoin(allocation.Denom, coinAllocated)
				coins.Add(coin)
			}
			contract := common.HexToAddress(incentive.Contract)
			coinsAllocated[contract] = coins
			return false
		},
	)

	// Distribution
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			contract := common.HexToAddress(incentive.Contract)

			// Get total cummulative gas per contract
			totalGas := k.GetTotalGas(ctx, incentive)

			// iterate over the gas meters per contract
			k.IterateIncentiveGasMeters(
				ctx,
				contract,
				func(gm types.GasMeter) (stop bool) {

					// reward
					coins := sdk.Coins{}
					for _, allocation := range incentive.Allocations {

						reward := allocatedCoin * gm.CummulativeGas / totalGas
						coin := sdk.Coin{Denom: allocation.Denom, Amount: reward}
						coinsAllocated.Coins.Add(coin)
					}
					participant := common.HexToAddress(gm.Participant)
					err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, participant.Bytes(), coins)
					if err != nil {
						logger.Debug(
							"failed to distribute incentive",
							"address", gm.Participant,
							"amount", coins.String(),
							"incentive", gm.Contract,
							"error", err.Error(),
						)
						return true // break iteration
					}

					// remove gas meter
					k.DeleteGasMeter(ctx, gm)
					// remuve cummulative gas meter per contract
					return false
				})

			incentive.Epochs--

			// remove incentive from epoch if already finalized
			if !incentive.IsActive() {
				k.DeleteIncentive(ctx, incentive)
				logger.Info(
					"incentive finalized",
					"contract", incentive.Contract,
				)
			}

			return false
		})

	return nil
}