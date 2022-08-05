package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v7/x/claims/types"
)

var (
	_ transfertypes.ICS4Wrapper = Keeper{}
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

// AfterProposalVote is a wrapper for calling the Gov AfterProposalVote hook on
// the module keeper
func (h Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	h.k.AfterProposalVote(ctx, proposalID, voterAddr)
}

// AfterProposalVote is called after a vote on a proposal is cast. Once the vote
// is successfully included, the claimable amount for the user's claims record
// vote action is claimed and the transferred to the user address.
func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	params := k.GetParams(ctx)

	claimsRecord, found := k.GetClaimsRecord(ctx, voterAddr)
	if !found {
		return
	}

	_, err := k.ClaimCoinsForAction(ctx, voterAddr, claimsRecord, types.ActionVote, params)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to claim Vote action",
			"address", voterAddr.String(),
			"error", err.Error(),
		)
	}
}

// AfterDelegationModified is a wrapper for calling the Staking AfterDelegationModified
// hook on the module keeper
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}

// AfterDelegationModified is called after a delegation is modified. Once a user
// delegates their EVMOS tokens to a validator, the claimable amount for the
// user's claims record delegation action is claimed and transferred to the user
// address.
func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	params := k.GetParams(ctx)

	claimsRecord, found := k.GetClaimsRecord(ctx, delAddr)
	if !found {
		return
	}

	_, err := k.ClaimCoinsForAction(ctx, delAddr, claimsRecord, types.ActionDelegate, params)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to claim Delegation action",
			"address", delAddr.String(),
			"error", err.Error(),
		)
	}
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements the ethermint evm PostTxProcessing hook.
// After a EVM state transition is successfully processed, the claimable amount
// for the users's claims record evm action is claimed and transferred to the
// user address.
func (k Keeper) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	params := k.GetParams(ctx)
	fromAddr := sdk.AccAddress(msg.From().Bytes())

	claimsRecord, found := k.GetClaimsRecord(ctx, fromAddr)
	if !found {
		return nil
	}

	_, err := k.ClaimCoinsForAction(ctx, fromAddr, claimsRecord, types.ActionEVM, params)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to claim EVM action",
			"address", fromAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

// ________________________________________________________________________________________

// governance hooks
func (h Hooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
}

func (h Hooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
}

func (h Hooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {}

func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
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

func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {}

// IBC callbacks and transfer handlers

// SendPacket implements the ICS4Wrapper interface from the transfer module. It
// calls the underlying SendPacket function directly to move down the middleware
// stack. Without SendPacket, this module would be skipped, when sending packages
// from the transferKeeper to core IBC.
func (k Keeper) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return k.ics4Wrapper.SendPacket(ctx, channelCap, packet)
}

// WriteAcknowledgement implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying WriteAcknowledgement function directly to move down the middleware stack.
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, channelCap, packet, ack)
}
