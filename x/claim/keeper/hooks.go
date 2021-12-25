package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/claim/types"
)

// TODO: add hooks for EVM txs and IBC

func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	_, err := k.ClaimCoinsForAction(ctx, voterAddr, types.ActionVote)
	if err != nil {
		panic(err.Error())
	}
}

func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	_, err := k.ClaimCoinsForAction(ctx, delAddr, types.ActionDelegate)
	if err != nil {
		panic(err.Error())
	}
}

// ________________________________________________________________________________________

// Hooks wrapper struct for the claim keeper
type Hooks struct {
	k Keeper
}

// TODO: add EVM hooks
var (
	_ evmtypes.EvmHooks         = Hooks{}
	_ govtypes.GovHooks         = Hooks{}
	_ stakingtypes.StakingHooks = Hooks{}
)

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// evm hook
func (h Hooks) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	return nil
}

// gov hooks
func (h Hooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
}

func (h Hooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
}

// governance hooks
func (h Hooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {}

func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
}

func (h Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	h.k.AfterProposalVote(ctx, proposalID, voterAddr)
}

func (h Hooks) AfterProposalInactive(ctx sdk.Context, proposalID uint64) {}
func (h Hooks) AfterProposalActive(ctx sdk.Context, proposalID uint64)   {}

// staking hooks
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)   {}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}

func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}

func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {}
