package osmosis

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/ics20"
)

const (
	// OsmosisXCSContract defines the contract address for the Osmosis XCS contract
	// OsmosisXCSContract = "osmo1xcsjj7g9qf6qy8w4xg2j3q4q3k6x5q2x9k5x2e"
	// SwapMethod defines the ABI method name for the Osmosis Swap function
	SwapMethod = "swap"
)

// Constants used during memo creation
const (
	slippage_percentage = "5"
	window_seconds = 10
)

// Swap swaps the given base denom for the given target denom on Osmosis and returns
// the newly swapped tokens to the receiver address.
func (p Precompile) Swap(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, input, output, amount, receiver, err := ParseSwapPacketData(args)
	if err != nil {
		return nil, err
	}

	inputTokenPairID := p.erc20Keeper.GetERC20Map(ctx, input)
	inputTokenPair, found := p.erc20Keeper.GetTokenPair(ctx, inputTokenPairID)
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, input)
	}
	inputDenom := inputTokenPair.Denom

	outputTokenPairID := p.erc20Keeper.GetERC20Map(ctx, output)
	outputTokenPair, found := p.erc20Keeper.GetTokenPair(ctx, outputTokenPairID)
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, output)
	}
	outputDenom := outputTokenPair.Denom

	bondDenom := p.stakingKeeper.GetParams(ctx).BondDenom

	err = ValidateSwap(ctx, p.portID, p.channelID, inputDenom, outputDenom, bondDenom)
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

	// Create the memo field for the Swap from the JSON file
	memo, err := CreateMemo(outputDenom, receiver, p.osmosisXCSContract, slippage_percentage, window_seconds)
	if err != nil {
		return nil, err
	}

	coin := sdk.Coin{Denom: inputDenom, Amount: sdk.NewIntFromBigInt(amount)}

	// Create the IBC Transfer message
	msg, err := NewMsgTransfer(p.portID, p.channelID, sender.String(), receiver, memo, coin)
	if err != nil {
		return nil, err
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	accept, expiration, err := ics20.CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	// Send the IBC Transfer message
	res, err := p.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err = ics20.UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, accept); err != nil {
		return nil, err
	}

	// Emit the ICS20 Transfer Event
	if err := ics20.EmitIBCTransferEvent(ctx, stateDB, p.ABI.Events, sender, p.Address(), msg); err != nil {
		return nil, err
	}

	// Emit the Osmosis Swap Event
	if err := p.EmitSwapEvent(ctx, stateDB, sender, input, output, amount, receiver); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Sequence, true)
}

