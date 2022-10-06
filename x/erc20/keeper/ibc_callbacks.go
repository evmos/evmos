package keeper

import (
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/evmos/evmos/v9/ibc"
	evmos "github.com/evmos/evmos/v9/types"
	"github.com/evmos/evmos/v9/x/erc20/types"
)

// TODO: overview
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {

	// Register the logger for the keeper
	logger := k.Logger(ctx)
	// Get the ERC20 parameters from the ERC20 Keeper
	erc20Params := k.GetParams(ctx)

	// Check and return original ACK if ERC20s and simplified conversions (ERC20s -> Cosmos Coins) are not enabled
	// Don't want to prevent unconverted IBC transfers in the case that ERC20s are disabled
	if !erc20Params.GetEnableERC20() || !erc20Params.GetEnableEVMHook() {
		return ack
	}

	// Get addresses in `evmos1` and the original bech32 format
	sender, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Return error ACK for blocked sender and recipient addresses
	if k.bankKeeper.BlockedAddr(sender) || k.bankKeeper.BlockedAddr(recipient) {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				types.ErrBlockedAddress,
				"sender (%s) or recipient (%s) address are in the deny list for sending and receiving transfers",
				senderBech32, recipientBech32,
			),
		)
	}

	// Get the denomination of the packet's IBC transfer
	denom, err := ibc.GetTransferDenomination(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Case 1: Denomination of IBC Transfer is base denomination of chain, noop
	if strings.Compare(sdk.GetBaseDenom(), denom) == 0 {
		return ack
	}

	// Case 2: Denomination of IBC Transfer is "erc20/{bytes}", noop if recognized (else error)
	if strings.HasPrefix(denom, "erc20/") {
		contractString := (strings.SplitN(denom, "/", 2))[0]
		if _, err = k.MintingEnabled(ctx, sender, recipient, contractString); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		return ack
	}

	// TODO: Case 3: Denomination of IBC Transfer is "ibc/{bytes}", conversion

	// Finally, check whether the denomination of the IBC transfer is of the form "ibc/"
	// Check if MintingEnabledd 
	// If so, convert the coins and reconstruct the packet, then continue
	// convert the tracedenom to a coin w/ appropriate amount
	// convert coins to the new one
	// use the new coin to fill in new packet
	
	// mockpacket.Data = transfertypes.ModuleCdc.MustMarshalJSON(
	// 	&transfertypes.FungibleTokenPacketData{
	// 		Sender:   "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v",
	// 		Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
	// 	},
	// )

	//https://pkg.go.dev/github.com/cosmos/ibc-go@v1.5.0/modules/apps/transfer/types#ParseDenomTrace

	return ack
}