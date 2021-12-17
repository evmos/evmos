package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/incentives/types"
)

// Distribute transfers the allocated rewards to the participants of a given
// incentive.
//  - allocates the amount to be distribbuted from the inflaction pool
//  - distributes the rewards to all paricpants
//  - deletes all gas meters
//  - sets the cumulative totalGas to zero
//  - updates the remaining epochs of each incentive
func (k Keeper) DistributeIncentives(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	// Allocate rewards
	coinsAllocated := k.allocateCoins(ctx)

	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			// Distribute rewards and reset total gasMeter
			k.rewardParticipants(ctx, incentive, coinsAllocated)
			k.ResetTotalGas(ctx, incentive)

			// Update Epoche and remove incentive from epoch if already finalized
			incentive.Epochs--
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

// Allocate amount of coins to be ditributed for each incentive
func (k Keeper) allocateCoins(ctx sdk.Context) map[common.Address]sdk.Coins {
	var coinsAllocated map[common.Address]sdk.Coins
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
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
	return coinsAllocated
}

// Reward Participants of a given Incentive and delete their gas meters
func (k Keeper) rewardParticipants(
	ctx sdk.Context,
	incentive types.Incentive,
	coinsAllocated map[common.Address]sdk.Coins,
) {
	logger := k.Logger(ctx)

	contract := common.HexToAddress(incentive.Contract)
	totalGas := k.GetTotalGas(ctx, incentive)

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
}
