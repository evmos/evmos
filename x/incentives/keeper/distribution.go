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
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Allocate amount of coins to be ditributed for each incentive
	var coinsAllocated map[common.Address]sdk.Coins
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			var coins sdk.Coins
			for _, allocation := range incentive.Allocations {
				balance := k.bankKeeper.GetBalance(ctx, moduleAddr, allocation.Denom)
				if !balance.IsPositive() {
					continue
				}
				coinAllocated := balance.Amount.ToDec().Mul(allocation.Amount)
				amount := coinAllocated.TruncateInt()
				coin := sdk.NewCoin(allocation.Denom, amount)
				coins.Add(coin)
			}
			contract := common.HexToAddress(incentive.Contract)
			coinsAllocated[contract] = coins
			return false
		},
	)

	// Distribute rewards for each incentive
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			contract := common.HexToAddress(incentive.Contract)
			totalGas := k.GetTotalGas(ctx, incentive)

			// iterate over the gas meters per contract
			k.IterateIncentiveGasMeters(
				ctx,
				contract,
				func(gm types.GasMeter) (stop bool) {
					// reward
					coins := sdk.Coins{}
					for _, allocation := range incentive.Allocations {
						coinAllocated := coinsAllocated[contract].AmountOf(allocation.Denom)
						reward := coinAllocated.MulRaw(int64(gm.CummulativeGas / totalGas))
						coin := sdk.Coin{Denom: allocation.Denom, Amount: reward}
						coins.Add(coin)
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

					return false
				})

			// remuve cummulative gas meter per contract
			k.ResetTotalGas(ctx, incentive)

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
