package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/x/withdraw/types"
)

// OnRecvPacket performs an IBC receive callback.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)

	// TODO: get params for list of enabled channels

	params := k.GetParams(ctx)
	// check channels from this chain (i.e destination)
	if !params.EnableWithdraw || !params.IsChannelEnabled(packet.DestinationChannel) {
		// return original ACK if withdraw is disabled globally or per channel
		return ack
	}

	// unmarshal packet data to obtain the sender and recipient
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// validate the sender bech32 address from the counterparty chain
	bech32Prefix := strings.Split(data.Sender, "1")[0]
	if bech32Prefix == data.Sender {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender: %s", data.Sender).Error(),
		)
	}

	senderBz, err := sdk.GetFromBech32(data.Sender, bech32Prefix)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender %s, %s", data.Sender, err.Error()).Error(),
		)
	}

	// change the bech32 human readable prefix (HRP) of the sender to `evmos1`
	sender := sdk.AccAddress(senderBz)

	// obtain the evmos recipient address
	recipient, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address %s", err.Error()).Error(),
		)
	}

	if !sender.Equals(recipient) {
		// return the original success acknowledgement since the address are different

		// NOTE: here we can't know if the recipient is a 'secp256k1' address. If it is,
		// then the user will need to withdraw the tokens using this same logic.
		return ack
	}

	balances := k.bankKeeper.GetAllBalances(ctx, recipient)

	// NOTE: since the balance is 0 we can reject the transaction
	// We return error since the sender and recipient addresses are the same
	if balances.IsZero() {
		logger.Debug(
			"rejected IBC transfer to 'secp256k1' key address",
			"sender", data.Sender, "recipient", data.Receiver,
		)

		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(types.ErrKeyTypeNotSupported, "receiver address %s", err.Error()).Error(),
		)
	}

	// transfer the balance back to the sender address

	srcPort := packet.DestinationPort
	srcChannel := packet.DestinationChannel

	// swap the sender and recipient
	sender = recipient
	recipientStr := data.Sender

	for _, coin := range balances {
		// withdraw the tokens back to sender
		if err := k.transferKeeper.SendTransfer(
			ctx,
			srcPort,                  // packet destination port is now the source
			srcChannel,               // packet destination channel is now the source
			coin,                     // balances + transfer amount
			sender,                   // transfer recipient is now the sender
			recipientStr,             // transfer sender is now the recipient
			clienttypes.ZeroHeight(), // timeout height disabled
			0,                        // timeout timestamp disabled
		); err != nil {
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(
					err,
					"failed to withdraw '%s' back to sender %s", coin.Denom, recipientStr,
				).Error(),
			)
		}
	}

	logger.Debug(
		"balances withdrawn to sender address",
		"sender", data.Sender,
		"receiver", data.Receiver,
		"balances", balances.String(),
		"source-port", packet.SourcePort,
		"source-channel", packet.SourceChannel,
	)

	// return error acknowledgement so that the counterparty chain can revert the
	// transfer
	return channeltypes.NewErrorAcknowledgement(
		sdkerrors.Wrapf(
			types.ErrKeyTypeNotSupported,
			"reverted IBC transfer from %s (%s) to recipient %s",
			data.Sender, sender, data.Receiver,
		).Error(),
	)
}
