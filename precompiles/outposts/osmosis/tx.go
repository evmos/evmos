package osmosis

import (
	"embed"
	"fmt"
	"slices"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	"github.com/evmos/evmos/v14/precompiles/ics20"
)

// Embed memo json file to the executable binary. Needed when importing as dependency.
//
//go:embed memo.json
var memoF embed.FS

const (
	// OsmosisXCSContract defines the contract address for the Osmosis XCS contract
	// OsmosisXCSContract = "osmo1xcsjj7g9qf6qy8w4xg2j3q4q3k6x5q2x9k5x2e"
	// SwapMethod defines the ABI method name for the Osmosis Swap function
	SwapMethod = "swap"
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
		return nil, fmt.Errorf("token pair for input address %s not found", input)
	}
	inputDenom := inputTokenPair.Denom

	outputTokenPairID := p.erc20Keeper.GetERC20Map(ctx, output)
	outputTokenPair, found := p.erc20Keeper.GetTokenPair(ctx, outputTokenPairID)
	if !found {
		return nil, fmt.Errorf("token pair for output address %s not found", output)
	}
	outputDenom := outputTokenPair.Denom

	err = p.validateSwap(ctx, inputDenom, outputDenom)
	if err != nil {
		return nil, err
	}

	// NOTE: substitute this logic with `ics20.CheckOriginAndSender`
	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	if contract.CallerAddress == sender {
		sender = origin
	} else if origin != sender {
		return nil, fmt.Errorf(ics20.ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	// Create the memo field for the Swap from the JSON file
	memo, err := createSwapMemo(output, receiver)
	if err != nil {
		return nil, err
	}

	// Create the IBC Transfer message
	msg, err := NewMsgTransfer(input, memo, amount, origin)
	if err != nil {
		return nil, err
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	var (
		expiration *time.Time
		auth       authz.Authorization
		resp       *authz.AcceptResponse
	)
	if contract.CallerAddress != origin {
		// check if authorization exists
		auth, expiration, err = authorization.CheckAuthzExists(ctx, p.AuthzKeeper, contract.CallerAddress, origin, ics20.TransferMsgURL)
		if err != nil {
			return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
		}

		// Accept the grant and return an error if the grant is not accepted
		resp, err = ics20.AcceptGrant(ctx, contract.CallerAddress, origin, msg, auth)
		if err != nil {
			return nil, err
		}
	}

	// Send the IBC Transfer message
	_, err = p.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed
	if contract.CallerAddress != origin {
		// accepts and updates the grant adjusting the spending limit
		if err = ics20.UpdateGrant(ctx, p.AuthzKeeper, contract.CallerAddress, origin, expiration, resp); err != nil {
			return nil, err
		}
	}

	// Emit the ICS20 Transfer Event
	if err := ics20.EmitIBCTransferEvent(ctx, stateDB, p.ABI.Events, sender, p.Address(), msg); err != nil {
		return nil, err
	}

	// Emit the Osmosis Swap Event
	if err := p.EmitSwapEvent(ctx, stateDB, sender, common.BytesToAddress(receiverAccAddr), amount, inputDenom, outputDenom, prefix); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// validateSwap performs validation on input and output denom.
func (p Precompile) validateSwap(
	ctx sdk.Context,
	input, output string,
) (err error) {

	// input and output cannot be equal
	if input == output {
		return fmt.Errorf("input and output token cannot be the same: %s", input)
	}

	// We have to compute the ibc voucher string for the osmo coin
	osmoTrace := ibctransfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", p.portID, p.channelID),
		BaseDenom: "uosmo",
	}
	osmoIBCDenom := osmoTrace.IBCDenom()
	// We need to get evmDenom from Params to have the code valid also in testnet
	evmDenom := p.evmKeeper.GetParams(ctx).EvmDenom

	// Check that the input token is evmos or osmo. This constraint will be removed in future
	validInput := []string{evmDenom, osmoIBCDenom}
	if !slices.Contains(validInput, input) {
		return fmt.Errorf("supported only the following input tokens: %v", validInput)
	}

	return nil
}

func (p Precompile) createMemo() string {

	osmosisSwap := OsmosisSwap{}
	// Convert the struct to a JSON string
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Print the JSON string
	fmt.Println(string(jsonBytes))
	return string(jsonBytes)
}

// createSwapMemo creates a memo for the swap transaction
func createSwapMemo(outputDenom, receiverAddress string) (string, error) {
	// Read the JSON memo from the file
	data, err := memoF.ReadFile("memo.json")
	if err != nil {
		return "", fmt.Errorf("failed to read JSON memo: %v", err)
	}

	return fmt.Sprintf(string(data), OsmosisXCSContract, outputDenom, receiverAddress), nil
}
