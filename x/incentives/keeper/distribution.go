package keeper

import (
	"math/big"
	"strconv"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evoblockchain/evoblock/v8/x/incentives/types"
)

// DistributeRewards transfers the allocated rewards to the participants of a given
// incentive.
//   - allocates the amount to be distributed from the inflation pool
//   - distributes the rewards to all participants
//   - deletes all gas meters
//   - updates the remaining epochs of each incentive
//   - sets the cumulative totalGas to zero
func (k Keeper) DistributeRewards(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	rewardAllocations, totalRewards, err := k.rewardAllocations(ctx)
	if err != nil {
		return err
	}

	k.IterateIncentives(ctx, func(incentive types.Incentive) (stop bool) {
		rewards, participants := k.rewardParticipants(ctx, incentive, rewardAllocations)

		incentive.Epochs--

		// Update Incentive and reset its total gas count. Remove incentive if it
		// has no remaining epochs left.
		if incentive.IsActive() {
			k.SetIncentive(ctx, incentive)
			k.SetIncentiveTotalGas(ctx, incentive, 0)
		} else {
			k.DeleteIncentiveAndUpdateAllocationMeters(ctx, incentive)
			logger.Info(
				"incentive finalized",
				"contract", incentive.Contract,
			)
		}

		defer func() {
			if !rewards.IsZero() {
				telemetry.IncrCounterWithLabels(
					[]string{types.ModuleName, "distribute", "participant", "total"},
					float32(participants),
					[]metrics.Label{
						telemetry.NewLabel("contract", incentive.Contract),
					},
				)
			}
		}()

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

	defer func() {
		for _, r := range totalRewards {
			if r.Amount.IsInt64() {
				telemetry.IncrCounterWithLabels(
					[]string{types.ModuleName, "distribute", "reward", "total"},
					float32(r.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", r.Denom)},
				)
			}
		}
	}()

	return nil
}

// rewardAllocations returns a map of each incentive's reward allocation
//   - Iterate over all the registered and active incentives
//   - create an allocation (module account) from escrow balance to be distributed to the contract address
//   - check that escrow balance is sufficient
func (k Keeper) rewardAllocations(
	ctx sdk.Context,
) (map[common.Address]sdk.Coins, sdk.Coins, error) {
	// Get balances on incentive module account
	denomBalances := make(map[string]sdk.Int)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	escrow := sdk.Coins{}

	// iterate over the module account balance insert elements to the denom -> amount
	// lookup map
	k.bankKeeper.IterateAccountBalances(ctx, moduleAddr, func(coin sdk.Coin) bool {
		denomBalances[coin.Denom] = coin.Amount
		// NOTE: all coins have different denomination so we can safely append instead
		// of using Add
		escrow = append(escrow, coin)
		return false
	})

	rewardAllocations := make(map[common.Address]sdk.Coins)
	rewards := sdk.Coins{}

	// iterate over all the incentives to define the allocation
	// amount for each contract
	k.IterateIncentives(
		ctx,
		func(incentive types.Incentive) (stop bool) {
			coins := sdk.Coins{}
			contract := common.HexToAddress(incentive.Contract)

			// calculate allocation for the incentivized contract
			for _, al := range incentive.Allocations {
				// Check if a balance to allocate exists
				if _, ok := denomBalances[al.Denom]; !ok {
					continue
				}

				// allocation for the contract is the amount escrowed * the allocation %
				coinAllocated := denomBalances[al.Denom].ToDec().Mul(al.Amount)
				amount := coinAllocated.TruncateInt()

				// NOTE: safety check, shouldn't occur since the allocation and balance
				// are > 0
				if !amount.IsPositive() {
					continue
				}

				coin := sdk.Coin{Denom: al.Denom, Amount: amount}
				coins = coins.Add(coin)
			}

			rewardAllocations[contract] = coins
			rewards = rewards.Add(coins...)

			return false
		},
	)

	// checks if module account has sufficient balance for allocation
	if rewards.IsAnyGT(escrow) {
		return nil, nil, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"escrowed balance < total coins allocated (%s < %s)",
			escrow, rewards,
		)
	}

	return rewardAllocations, rewards, nil
}

// rewardParticipants reward participants of a given Incentive, delete their gas
// meters and returns a count of all gas meters
//   - Check if participants spent gas on interacting with incentive
//   - Iterate over the incentive participants' gas meters
//   - Allocate rewards according to participants gasRatio and cap them at 100% of their gas spent on interaction with incentive
//   - Send rewards to participants
//   - Delete gas meter
func (k Keeper) rewardParticipants(
	ctx sdk.Context,
	incentive types.Incentive,
	coinsAllocated map[common.Address]sdk.Coins,
) (rewards sdk.Coins, count uint64) {
	logger := k.Logger(ctx)

	// Check if coin allocation was successful
	contract := common.HexToAddress(incentive.Contract)
	contractAllocation, ok := coinsAllocated[contract]
	if !ok {
		logger.Debug(
			"contract allocation coins not found",
			"contract", incentive.Contract,
		)
		return sdk.Coins{}, 0
	}

	// Check if participants spent gas on interacting with incentive
	totalGas := incentive.TotalGas
	if totalGas == 0 {
		logger.Debug(
			"no gas spent on incentive during epoch",
			"contract", incentive.Contract,
		)
		return sdk.Coins{}, 0
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

				// Cap rewards in mint denom (i.e. aEVO) to receive only up to 100% of
				// the participant's gas spent and prevent gaming
				if mintDenom == allocation.Denom {
					rewardCap := cumulativeGas.Mul(rewardScaler)
					reward = sdk.MinDec(reward, rewardCap)
				}

				// NOTE: ignore denom validation
				coin := sdk.Coin{Denom: allocation.Denom, Amount: reward.TruncateInt()}
				coins = coins.Add(coin)
			}

			rewards = rewards.Add(coins...)

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
			count++

			return false
		},
	)

	return rewards, count
}
