package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

// OnRecvPacket performs an IBC receive callback.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)

	params := k.GetParams(ctx)
	claimsParams := k.claimsKeeper.GetParams(ctx)

	// check channels from this chain (i.e destination)
	if !params.EnableWithdraw || !claimsParams.IsAuthorizedChannel(packet.DestinationChannel) {
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

		// return the an error acknowledgement since the address the same
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(types.ErrKeyTypeNotSupported, "receiver address %s is not a valid ethereum address", data.Receiver).Error(),
		)
	}

	// swap the sender and recipient
	sender = recipient

	// transfer the balance back to the sender address
	var (
		recipientStr        string
		srcPort, srcChannel string
	)

	for _, coin := range balances {
		// we only transfer IBC tokens back to their respective source chains
		if strings.HasPrefix(coin.Denom, "ibc/") {
			// TODO: recipient should have the bech32 prefix from the chain
			recipientStr = data.Sender

			srcPort, srcChannel, err = k.GetIBCDenomSource(ctx, coin.Denom, data.Sender)
			if err != nil {
				return channeltypes.NewErrorAcknowledgement(err.Error())
			}

		} else {
			// send Evmos native tokens to the source port and channel
			recipientStr = data.Sender

			srcPort = packet.SourcePort
			srcChannel = packet.SourceChannel
		}

		// TODO: get the correct bech32 address of the source chain
		recipientStr = data.Sender

		// TODO: coin needs

		// Withdraw the tokens to the bech32 prefixed address of the source chain
		err = k.transferKeeper.SendTransfer(
			ctx,
			srcPort,                  // packet destination port is now the source
			srcChannel,               // packet destination channel is now the source
			coin,                     // balances + transfer amount
			sender,                   // transfer recipient is now the sender
			recipientStr,             // transfer sender is now the recipient
			clienttypes.ZeroHeight(), // timeout height disabled
			0,                        // timeout timestamp disabled
		)
	}

	if err != nil {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				err,
				"failed to withdraw IBC vouchers back to sender '%s' in the corresponding IBC chain", data.Sender,
			).Error(),
		)
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

func (k Keeper) GetIBCDenomSource(ctx sdk.Context, denom, sender string) (srcPort, srcChannel string, err error) {
	ibcHexHash := strings.SplitN(denom, "/", 2)[1]
	hash, err := transfertypes.ParseHexHash(ibcHexHash)
	if err != nil {
		return "", "", sdkerrors.Wrapf(
			err,
			"failed to withdraw IBC vouchers back to sender '%s' in the corresponding IBC chain", sender,
		)
	}

	denomTrace, found := k.transferKeeper.GetDenomTrace(ctx, hash)
	if !found {
		return "", "", sdkerrors.Wrapf(
			transfertypes.ErrTraceNotFound,
			"failed to withdraw IBC vouchers back to sender '%s' in the corresponding IBC chain", sender,
		)
	}

	path := strings.Split(denomTrace.Path, "/")
	if len(path)%2 != 0 {
		return "", "", sdkerrors.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid denom (%s) trace path %s", denomTrace.BaseDenom, denomTrace.Path,
		)
	}

	counterpartyPortID := path[0]
	counterpartyChannelID := path[1]

	channel, found := k.channelKeeper.GetChannel(ctx, counterpartyPortID, counterpartyChannelID)
	if !found {
		return "", "", sdkerrors.Wrapf(
			channeltypes.ErrChannelNotFound,
			"port ID %s, channel ID %s", counterpartyPortID, counterpartyChannelID,
		)
	}

	srcPort = channel.Counterparty.PortId
	srcChannel = channel.Counterparty.ChannelId

	return srcPort, srcChannel, nil
}
