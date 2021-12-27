package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tharsis/evmos/x/claim/types"
)

// EndBlocker checks if the airdrop claiming period has ended
func (k Keeper) EndBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// NOTE: ignore end of airdrop period check if claiming is disabled
	if !params.EnableClaim {
		return
	}

	// check if the time to claim the airdrop tokens has passed
	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)
	if elapsedAirdropTime <= params.DurationUntilDecay+params.DurationOfDecay {
		return
	}

	// clawback all the remaining tokens to the community pool and
	if err := k.EndAirdrop(ctx, params); err != nil {
		panic(err)
	}
}

// EndAirdrop transfers the unclaimed tokens from the airdrop to the community pool,
// then clears the state by removing all the entries
func (k Keeper) EndAirdrop(ctx sdk.Context, params types.Params) error {
	logger := k.Logger(ctx)
	logger.Info("beginning EndAirdrop logic")

	if err := k.ClawbackEscrowedTokens(ctx); err != nil {
		return err
	}

	if err := k.ClawbackEmptyAccounts(ctx, params.ClaimDenom); err != nil {
		return err
	}

	logger.Info("clearing claim record state entries")
	k.DeleteClaimRecords(ctx)

	// set the EnableClaim param to false so that we don't have to compute duration every block
	params.EnableClaim = false
	k.SetParams(ctx, params)

	return nil
}

func (k Keeper) ClawbackEscrowedTokens(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAccAddr)

	if balances.IsZero() {
		logger.Debug("clawback aborted, airdrop escrow account is empty")
		return nil
	}

	if err := k.distrKeeper.FundCommunityPool(ctx, balances, moduleAccAddr); err != nil {
		return sdkerrors.Wrap(err, "failed to transfer escrowed airdrop tokens")
	}

	logger.Info(
		"clawback of funds to community pool treasury",
		"total", balances.String(),
	)

	return nil
}

// ClawbackEmptyAccounts performs the a claw back off all the EVMOS tokens from airdrop
// recipient accounts with a sequence number of 0.
func (k Keeper) ClawbackEmptyAccounts(ctx sdk.Context, claimDenom string) error {
	totalClawback := sdk.Coins{}
	logger := k.Logger(ctx)

	accPruned := int64(0)

	for _, bechAddr := range types.AirdropAddrs {
		addr, err := sdk.AccAddressFromBech32(bechAddr)
		if err != nil {
			return err
		}

		acc := k.accountKeeper.GetAccount(ctx, addr)
		if acc == nil {
			logger.Debug("airdrop account not found during clawback", "address", addr.String())
			continue
		}

		seq, err := k.accountKeeper.GetSequence(ctx, addr)
		if err != nil {
			return err
		}

		if seq != 0 {
			continue
		}

		balances := k.bankKeeper.GetAllBalances(ctx, addr)

		// prune empty accounts from the airdrop
		if balances == nil || balances.IsZero() {
			k.accountKeeper.RemoveAccount(ctx, acc)
			// TODO: update bank module to allow clearing the empty balance state
			accPruned++
			continue
		}

		clawbackCoin := sdk.Coin{Denom: claimDenom, Amount: balances.AmountOfNoDenomValidation(claimDenom)}
		if !clawbackCoin.IsPositive() {
			continue
		}

		// When sequence number is 0, _and_ from an airdrop account,
		// clawback all the aevmos to community pool.
		//
		// ***Reminder***
		// 'Unclaimed' tokens are defined as being in wallets which have a Sequence Number = 0,
		// which means the address has NOT performed a single action during the airdrop claim window.

		// ******CLAWBACK PROPOSED FRAMEWORK******
		// TLDR -- Send ALL unclaimed EVMOS back to the community pool
		// and prune those inactive wallets from current state.

		if err := k.distrKeeper.FundCommunityPool(ctx, sdk.Coins{clawbackCoin}, addr); err != nil {
			return err
		}

		totalClawback = totalClawback.Add(clawbackCoin)
	}

	logger.Info(
		"clawed back funds into community pool",
		"total", totalClawback.String(),
		"pruned-accounts", strconv.FormatInt(accPruned, 64),
	)
	return nil
}
