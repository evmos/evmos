package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/claim/types"
)

var (
	_ transfertypes.ICS4Wrapper = Hooks{}
	_ evmtypes.EvmHooks         = Hooks{}
	_ govtypes.GovHooks         = Hooks{}
	_ stakingtypes.StakingHooks = Hooks{}
)

// Hooks wrapper struct for the claim keeper
type Hooks struct {
	k Keeper
}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	claimRecord, found := k.GetClaimRecord(ctx, voterAddr)
	if !found {
		// no-op
		return
	}

	params := k.GetParams(ctx)

	_, err := k.ClaimCoinsForAction(ctx, voterAddr, claimRecord, types.ActionVote, params)
	if err != nil {
		panic(err.Error())
	}
}

func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	claimRecord, found := k.GetClaimRecord(ctx, delAddr)
	if !found {
		// no-op
		return
	}

	params := k.GetParams(ctx)

	_, err := k.ClaimCoinsForAction(ctx, delAddr, claimRecord, types.ActionDelegate, params)
	if err != nil {
		panic(err.Error())
	}
}

func (k Keeper) AfterEVMStateTransition(ctx sdk.Context, from common.Address, to *common.Address, receipt *ethtypes.Receipt) error {
	fromAddr := sdk.AccAddress(from.Bytes())

	claimRecord, found := k.GetClaimRecord(ctx, fromAddr)
	if !found {
		// no-op
		return nil
	}

	params := k.GetParams(ctx)

	_, err := k.ClaimCoinsForAction(ctx, fromAddr, claimRecord, types.ActionEVM, params)
	if err != nil {
		return err
	}

	return nil
}

// ________________________________________________________________________________________

// IBC callbacks and transfer handlers

// SendPacket implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying SendPacket function directly to move down the middleware stack.
func (h Hooks) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return h.k.ics4Wrapper.SendPacket(ctx, channelCap, packet)
}

// evm hook
func (h Hooks) PostTxProcessing(ctx sdk.Context, from common.Address, to *common.Address, receipt *ethtypes.Receipt) error {
	return h.k.AfterEVMStateTransition(ctx, from, to, receipt)
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
