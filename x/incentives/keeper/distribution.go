package keeper

import (
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v2/x/incentives/types"
)

// Distribute transfers the allocated rewards to the participants of a given
// incentive.
//  - allocates the amount to be distributed from the inflation pool
//  - distributes the rewards to all particpants
//  - deletes all gas meters
//  - updates the remaining epochs of each incentive
//  - sets the cumulative totalGas to zero
func (k Keeper) DistributeIncentives(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	// Allocate rewards for each Incentive
	coinsAllocated, err := k.allocateCoins(ctx)
	if err != nil {
		return err
	}

	// Iterate over each incentive and distribute allocated rewards
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			// Distribute rewards
			k.rewardParticipants(ctx, incentive, coinsAllocated)

			// Update epoch
			incentive.Epochs--

			// Update Incentive and reset its total gas count
			if incentive.IsActive() {
				k.SetIncentive(ctx, incentive)
				k.SetIncentiveTotalGas(ctx, incentive, 0)
			} else {
				// Remove incentive if it has no remaining epochs
				k.DeleteIncentiveAndUpdateAllocationMeters(ctx, incentive)
				logger.Info(
					"incentive finalized",
					"contract", incentive.Contract,
				)
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeDistributeIncentives,
					sdk.NewAttribute(types.AttributeKeyContract, incentive.Contract),
					sdk.NewAttribute(
						types.AttributeKeyEpochs,
						strconv.FormatUint(uint64(incentive.Epochs), 10),
					),
				),
			)
			return false
		})

	return nil
}

// Allocate amount of coins to be distributed for each incentive
//  - Iterate over all the registered and active incentives
//  - create an allocation (module account) from escrow balance to be distributed to the contract address
//  - check that escrow balance is sufficient
func (k Keeper) allocateCoins(ctx sdk.Context) (map[common.Address]sdk.Coins, error) {
	// Get balances on incentive module account
	denomBalances := make(map[string]sdk.Int)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	escrowedCoins := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	for _, coin := range escrowedCoins {
		if !coin.Amount.IsPositive() {
			continue
		}
		denomBalances[coin.Denom] = coin.Amount
	}

	// Iterate over each incentive's allocations to create map off allocated coins
	coinsAllocated := make(map[common.Address]sdk.Coins)
	totalAllocated := sdk.Coins{}
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			coins := sdk.Coins{}
			contract := common.HexToAddress(incentive.Contract)

			for _, al := range incentive.Allocations {
				// Check if a balance to allocate exists
				if _, ok := denomBalances[al.Denom]; !ok {
					continue
				}

				// Create escrow balance for allocation
				coinAllocated := denomBalances[al.Denom].ToDec().Mul(al.Amount)
				amount := coinAllocated.TruncateInt()
				coin := sdk.Coin{Denom: al.Denom, Amount: amount}
				coins = coins.Add(coin)
			}

			coinsAllocated[contract] = coins
			totalAllocated = totalAllocated.Add(coins...)

			return false
		},
	)

	// checks if escrow balance has sufficient balance for allocation
	if totalAllocated.IsAnyGT(escrowedCoins) {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"escrowed balance < total coins allocated (%s < %s)",
			escrowedCoins, totalAllocated,
		)
	}

	return coinsAllocated, nil
}

// Reward Participants of a given Incentive and delete their gas meters
//  - Check if participants spent gas on interacting with incentive
//  - Iterate over the incentive participants' gas meters
//    - Allocate rewards according to participants gasRatio and cap them at 100% of their gas spent on interaction with incentive
//    - Send rewards to participants
//    - Delete gas meter
func (k Keeper) rewardParticipants(
	ctx sdk.Context,
	incentive types.Incentive,
	coinsAllocated map[common.Address]sdk.Coins,
) {
	logger := k.Logger(ctx)

	// Check if coin allocation was successful
	contract := common.HexToAddress(incentive.Contract)
	contractAllocation, ok := coinsAllocated[contract]
	if !ok {
		logger.Debug(
			"contract allocation coins not found",
			"contract", incentive.Contract,
		)
		return
	}

	// Check if participants spent gas on interacting with incentive
	totalGas := incentive.TotalGas
	if totalGas == 0 {
		logger.Debug(
			"no gas spent on incentive during epoch",
			"contract", incentive.Contract,
		)
		return
	}

	totalGasDec := sdk.NewDecFromBigInt(new(big.Int).SetUint64(totalGas))
	mintDenom := k.evmKeeper.GetParams(ctx).EvmDenom
	rewardScaler := k.GetParams(ctx).RewardScaler

	// Iterate over the incentive's gas meters and distribute rewards
	k.IterateIncentiveGasMeters(
		ctx,
		contract,
		func(gm types.GasMeter) (stop bool) {
			// Get participant's ratio of `gas spent / total gas spent`
			cumulativeGas := sdk.NewDecFromBigInt(new(big.Int).SetUint64(gm.CumulativeGas))
			gasRatio := cumulativeGas.Quo(totalGasDec)
			coins := sdk.Coins{}

			// Allocate rewards according to gasRatio
			for _, allocation := range incentive.Allocations {
				coinAllocated := contractAllocation.AmountOf(allocation.Denom)
				reward := gasRatio.MulInt(coinAllocated)
				if !reward.IsPositive() {
					continue
				}

				// Cap rewards in mint denom (i.e. aevmos) to receive only up to 100% of
				// the participant's gas spent and prevent gaming
				if mintDenom == allocation.Denom {
					rewardCap := cumulativeGas.Mul(rewardScaler)
					reward = sdk.MinDec(reward, rewardCap)
				}

				// NOTE: ignore denom validation
				coin := sdk.Coin{Denom: allocation.Denom, Amount: reward.TruncateInt()}
				coins = coins.Add(coin)
			}

			// Send rewards to participant
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

			// Remove gas meter once the rewards are distributed
			k.DeleteGasMeter(ctx, gm)

			return false
		})
}
