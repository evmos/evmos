package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/tharsis/evmos/v3/x/claims/types"
)

// EndBlocker checks if the airdrop claiming period has ended in order to
// process the clawback of unclaimed tokens
func (k Keeper) EndBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// NOTE: ignore end of airdrop period check if claiming is disabled
	if !params.EnableClaims {
		return
	}

	// check if the time to claim airdrop tokens has passed
	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)
	if elapsedAirdropTime <= params.DurationUntilDecay+params.DurationOfDecay {
		return
	}

	if err := k.EndAirdrop(ctx, params); err != nil {
		panic(err)
	}
}

// EndAirdrop transfers the unclaimed tokens from the airdrop to the community
// pool, removes all claims records from state and disables the claims.
func (k Keeper) EndAirdrop(ctx sdk.Context, params types.Params) error {
	logger := k.Logger(ctx)
	logger.Info("beginning EndAirdrop logic")

	if err := k.ClawbackEscrowedTokens(ctx); err != nil {
		return err
	}

	// transfer unclaimed tokens from accounts to community pool and clean up the
	// claims record state
	k.ClawbackEmptyAccounts(ctx, params.ClaimsDenom)

	// set the EnableClaims param to false so that we don't have to compute
	// duration every block
	params.EnableClaims = false
	k.SetParams(ctx, params)
	logger.Info("end EndAirdrop logic")
	return nil
}

// ClawbackEscrowedTokens transfers all the escrowed airdrop tokens on the
// ModuleAccount to the community pool
func (k Keeper) ClawbackEscrowedTokens(ctx sdk.Context) error {
	logger := k.Logger(ctx)

	moduleAccAddr := k.GetModuleAccountAddress()
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

// ClawbackEmptyAccounts performs the clawback of all allocated tokens
// from airdrop recipient accounts with a sequence number of 0 (i.e the account
// hasn't performed a single tx during the claim window).
// Once the account is clawbacked, the claims record is deleted from state.
func (k Keeper) ClawbackEmptyAccounts(ctx sdk.Context, claimsDenom string) {
	totalClawback := sdk.Coins{}
	logger := k.Logger(ctx)

	accPruned := int64(0)
	accClawbacked := int64(0)

	k.IterateClaimsRecords(ctx, func(addr sdk.AccAddress, _ types.ClaimsRecord) (stop bool) {
		// delete claims record once the account balance is clawed back
		defer k.DeleteClaimsRecord(ctx, addr)

		acc := k.accountKeeper.GetAccount(ctx, addr)
		if acc == nil {
			logger.Debug(
				"airdrop account not found during clawback",
				"address", addr.String(),
			)
			return false
		}

		// ignore vesting accounts since some of the funds might be locked
		if _, isVesting := acc.(vestexported.VestingAccount); isVesting {
			return false
		}

		seq, err := k.accountKeeper.GetSequence(ctx, addr)
		if err != nil {
			logger.Debug(
				"airdrop account nonce not found during clawback",
				"address", addr.String(),
			)
			return false
		}

		if seq != 0 {
			return false
		}

		balances := k.bankKeeper.GetAllBalances(ctx, addr)

		// prune empty accounts from the airdrop
		if balances == nil || balances.IsZero() {
			k.accountKeeper.RemoveAccount(ctx, acc)
			accPruned++
			return false
		}

		clawbackCoin := sdk.Coin{Denom: claimsDenom, Amount: balances.AmountOfNoDenomValidation(claimsDenom)}
		if !clawbackCoin.IsPositive() {
			return false
		}

		// Send all unclaimed airdropped coins back to the community pool
		// and prune those inactive wallets from current state.
		// "Unclaimed" tokens are defined as being in wallets which have a sequence
		// number = 0, which means the address has NOT performed a single action
		// during the airdrop claim window.
		if err := k.distrKeeper.FundCommunityPool(ctx, sdk.Coins{clawbackCoin}, addr); err != nil {
			logger.Debug(
				"not enough balance to clawback account",
				"address", addr.String(),
				"amount", clawbackCoin.String(),
			)
			return false
		}

		totalClawback = totalClawback.Add(clawbackCoin)
		accClawbacked++

		return false
	})

	logger.Info(
		"clawed back funds into community pool",
		"total", totalClawback.String(),
		"clawbacked-accounts", strconv.FormatInt(accClawbacked, 10),
		"pruned-accounts", strconv.FormatInt(accPruned, 10),
	)
}
