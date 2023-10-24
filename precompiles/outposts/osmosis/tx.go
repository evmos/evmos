// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/ics20"

	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// SwapMethod is the name of the swap method
	SwapMethod = "swap"
	// SwapAction is the action name needed in the memo field
	SwapAction = "Swap"
)

// Swap is a transaction that swap tokens on the Osmosis chain using
// an ICS20 transfer with a custom memo field to trigger the XCS V2 contract.
func (p Precompile) Swap(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {

	sender, input, output, amount, slippagePercentage, windowSeconds, receiver, err := ParseSwapPacketData(args)
	if err != nil {
		return nil, err
	}

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err = ics20.CheckOriginAndSender(contract, origin, sender)
	if err != nil {
		return nil, err
	}

	// Given the HEX address of the input, we obtain the denom from the ERC20
	// keeper. We need it to compare the input and output denom.
	inputTokenPairID := p.erc20Keeper.GetERC20Map(ctx, input)
	inputTokenPair, found := p.erc20Keeper.GetTokenPair(ctx, inputTokenPairID)
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, input)
	}
	inputDenom := inputTokenPair.Denom

	// Given the HEX address of the output, we obtain the denom from the ERC20
	// keeper. We need it to compare the input and output denom.
	outputTokenPairID := p.erc20Keeper.GetERC20Map(ctx, output)
	outputTokenPair, found := p.erc20Keeper.GetTokenPair(ctx, outputTokenPairID)
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, output)
	}
	outputDenom := outputTokenPair.Denom

	// We need the bonded denom just for the outpost alpha version where the
	// the only two inputs allowed are aevmos and uosmo.
	bondDenom := p.stakingKeeper.GetParams(ctx).BondDenom

	packet := osmosisoutpost.CreatePacketWithMemo(
		tc.outputDenom, tc.receiver, tc.contract, tc.slippagePercentage, tc.windowSeconds, tc.onFailedDelivery, tc.nextMemo,
	)
	packetString := packet.String()

	return nil, nil
}
