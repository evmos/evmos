package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/incentives/types"
)

// Distribute transfers the allocated rewards to the participants of a given
// incentive.
//  - allocates the amount to be distribbuted from the inflaction pool
//  - distributes the rewards to all paricpants
//  - deletes all gas meters
//  - updates the remaining epochs of each incentive
//  - sets the cumulative totalGas to zero
func (k Keeper) DistributeIncentives(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	// Allocate rewards for each Incentive
	coinsAllocated := k.allocateCoins(ctx)

	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			// Distribute rewards
			k.rewardParticipants(ctx, incentive, coinsAllocated)

			// Update Epoche and remove incentive from epoch if already finalized
			incentive.Epochs--
			if !incentive.IsActive() {
				k.DeleteIncentive(ctx, incentive)
				logger.Info(
					"incentive finalized",
					"contract", incentive.Contract,
				)
			} else {
				k.SetIncentive(ctx, incentive)
				// reset incentive's total gas count
				k.SetIncentiveTotalGas(ctx, incentive, 0)
			}

			return false
		})

	return nil
}

// Allocate amount of coins to be distributed for each incentive
func (k Keeper) allocateCoins(ctx sdk.Context) map[common.Address]sdk.Coins {
	coinsAllocated := make(map[common.Address]sdk.Coins)

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			coins := sdk.Coins{}
			contract := common.HexToAddress(incentive.Contract)

			for _, allocation := range incentive.Allocations {
				balance := k.bankKeeper.GetBalance(ctx, moduleAddr, allocation.Denom)
				if !balance.IsPositive() {
					continue
				}
				coinAllocated := balance.Amount.ToDec().Mul(allocation.Amount)
				amount := coinAllocated.TruncateInt()
				coin := sdk.Coin{Denom: allocation.Denom, Amount: amount}
				coins = coins.Add(coin)
			}

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
	contractAllocation, ok := coinsAllocated[contract]
	if !ok {
		logger.Debug(
			"contract allocation coins not found",
			"contract", incentive.Contract,
		)
		return
	}

	totalGas := k.GetIncentiveTotalGas(ctx, incentive)
	if totalGas == 0 {
		logger.Debug(
			"no gas spent on incentive during epoch",
			"contract", incentive.Contract,
		)
		return
	}
	totalGasDec := sdk.NewDecFromBigInt(new(big.Int).SetUint64(totalGas))

	k.IterateIncentiveGasMeters(
		ctx,
		contract,
		func(gm types.GasMeter) (stop bool) {
			// get the participant ratio of their gas spent / total gas
			cumulativeGas := sdk.NewDecFromBigInt(new(big.Int).SetUint64(gm.CumulativeGas))
			gasRatio := cumulativeGas.Quo(totalGasDec)
			coins := sdk.Coins{}

			// allocate the coins corresponding to the ratio of gas spent
			for _, allocation := range incentive.Allocations {
				coinAllocated := contractAllocation.AmountOf(allocation.Denom)
				reward := gasRatio.MulInt(coinAllocated)
				if !reward.IsPositive() {
					continue
				}

				// NOTE: ignore denom validation
				coin := sdk.Coin{Denom: allocation.Denom, Amount: reward.TruncateInt()}
				coins = coins.Add(coin)
			}

			participant := common.HexToAddress(gm.Participant)
			err := k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				types.ModuleName,
				sdk.AccAddress(participant.Bytes()),
				coins,
			)

			if err != nil {
				logger.Debug(
					"failed to distribute incentive",
					"address", gm.Participant,
					"allocation", coins.String(),
					"incentive", gm.Contract,
					"error", err.Error(),
				)
				return true // break iteration
			}

			// remove gas meter once the incentives are allocated to the user
			k.DeleteGasMeter(ctx, gm)

			return false
		})
}
