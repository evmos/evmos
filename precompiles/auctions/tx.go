// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

const (
	// BidMethod defines the ABI method name for the auctions
	// Bid transaction.
	BidMethod = "bid"
	// DepositCoinMethod defines the ABI method name for the auctions
	// DepositCoin transaction.
	DepositCoinMethod = "depositCoin"
)

// Bid bids on the current auction with a specified Evmos amount that must be higher than the highest bid.
func (p *Precompile) Bid(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	_ *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, msgBid, err := NewMsgBid(args)
	if err != nil {
		return nil, err
	}

	var (
		// isCallerOrigin is true when the contract caller is the same as the origin
		isCallerOrigin = contract.CallerAddress == origin
		// isCallerSender is true when the contract caller is the same as the sender
		isCallerSender = contract.CallerAddress == sender
	)

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	if isCallerSender {
		sender = origin
	} else if origin != sender {
		return nil, fmt.Errorf(ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	// TODO: Do we need a generic Authz or a custom one here?

	msgBid.Sender = sdk.AccAddress(sender.Bytes()).String()
	_, err = p.auctionsKeeper.Bid(ctx, msgBid)
	if err != nil {
		return nil, err
	}

	currentRound := p.auctionsKeeper.GetRound(ctx)
	// emits an event for the Bid transaction.
	if err := p.EmitBidEvent(ctx, stateDB, sender, currentRound, msgBid.Amount.Amount.BigInt()); err != nil {
		return nil, err
	}

	if !isCallerOrigin {
		// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
		// when calling the precompile from a smart contract
		// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(sender, msgBid.Amount.Amount.BigInt(), cmn.Sub))
	}

	return cmn.TrueValue, nil
}

// DepositCoin deposits coins into the auction collector module to be used in the following auction.
func (p *Precompile) DepositCoin(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	_ *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, tokenAddress, msgDepositCoin, err := NewMsgDepositCoin(args, ctx, p.erc20Keeper)
	if err != nil {
		return nil, err
	}

	var (
		// isCallerOrigin is true when the contract caller is the same as the origin
		isCallerOrigin = contract.CallerAddress == origin
		// isCallerSender is true when the contract caller is the same as the sender
		isCallerSender = contract.CallerAddress == sender
	)

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	if isCallerSender {
		sender = origin
	} else if origin != sender {
		return nil, fmt.Errorf(ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	msgDepositCoin.Sender = sdk.AccAddress(sender.Bytes()).String()
	_, err = p.auctionsKeeper.DepositCoin(ctx, msgDepositCoin)
	if err != nil {
		return nil, err
	}

	currentRound := p.auctionsKeeper.GetRound(ctx)
	// emits an event for the DepositCoin transaction.
	if err := p.EmitDepositCoinEvent(ctx, stateDB, sender, currentRound, tokenAddress, msgDepositCoin.Amount.Amount.BigInt()); err != nil {
		return nil, err
	}

	if !isCallerOrigin {
		// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
		// when calling the precompile from a smart contract
		// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
		p.SetBalanceChangeEntries(cmn.NewBalanceChangeEntry(sender, msgDepositCoin.Amount.Amount.BigInt(), cmn.Sub))
	}

	return cmn.TrueValue, nil
}
