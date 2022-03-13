package keeper

import (
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"

	evmos "github.com/tharsis/evmos/v2/types"
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

	// get the recipient account
	account := k.accountKeeper.GetAccount(ctx, recipient)

	// case 2.a: recipient pubkey not secp256k1 -> don't withdraw, success ACK
	// case 2.b: no recipient account, no pubkey,

	// balance == 0  -> don't withdraw, success ACK

	// case 1: sender ≠ recipient
	if !sender.Equals(recipient) {
		switch {
		// 1.a: no recipient account or no pubkey -> no way to determine the type of address
		// ==> return success ACK
		case account == nil,
			account.GetPubKey() == nil,
			// 1.b: pubkey is eth_secp256k1 (valid ethereum address) or other supported keys
			// ==> return success ACK
			account.GetPubKey().Type() == ethsecp256k1.KeyType,
			account.GetPubKey().Type() == (&multisig.LegacyAminoPubKey{}).Type(),
			account.GetPubKey().Type() == (&ed25519.PubKey{}).Type():
			return ack
		default:
			// 1.c: pubkey not eth_secp256k1 ==> error ACK
			logger.Debug(
				"rejected IBC transfer to 'secp256k1' key address",
				"sender", data.Sender, "recipient", data.Receiver,
			)
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(evmos.ErrKeyTypeNotSupported, "receiver address %s is not a valid ethereum address", data.Receiver).Error(),
			)
		}
	}

	// case 2: sender address = recipient address

	// NOTE: here we need to check if the recipient is an ethereum account, if it is,
	// then we need to return the original ACK
	if account != nil &&
		account.GetPubKey() != nil &&
		(account.GetPubKey().Type() == ethsecp256k1.KeyType ||
			account.GetPubKey().Type() == (&multisig.LegacyAminoPubKey{}).Type() ||
			account.GetPubKey().Type() == (&ed25519.PubKey{}).Type()) {
		return ack
	}

	// transfer the balance back to the sender address
	srcPort := packet.SourcePort
	srcChannel := packet.SourceChannel
	balances := sdk.Coins{}

	k.bankKeeper.IterateAccountBalances(ctx, recipient, func(coin sdk.Coin) (stop bool) {
		if coin.IsZero() {
			// continue
			return false
		}

		// we only transfer IBC tokens back to their respective source chains
		if strings.HasPrefix(coin.Denom, "ibc/") {
			srcPort, srcChannel, err = k.GetIBCDenomSource(ctx, coin.Denom, data.Sender)
			if err != nil {
				logger.Error(
					"failed to get the IBC full denom path of source chain",
					"error", err.Error(),
				)
				return true // stop iteration
			}

			// NOTE: only withdraw the IBC tokens from the source channel
			if packet.SourcePort != srcPort || packet.SourceChannel != srcChannel {
				// reset to the original values
				srcPort = packet.SourcePort
				srcChannel = packet.SourceChannel
				// continue
				return false
			}
		}

		// Native tokens will be transferred to the authorized source chain to unstuck them

		// Withdraw the tokens to the bech32 prefixed address of the source chain
		err = k.transferKeeper.SendTransfer(
			ctx,
			srcPort,                  // packet destination port is now the source
			srcChannel,               // packet destination channel is now the source
			coin,                     // balances + transfer amount
			recipient,                // transfer recipient is now the sender
			data.Sender,              // transfer sender is now the recipient
			clienttypes.ZeroHeight(), // timeout height disabled
			uint64(ctx.BlockTime().Add(time.Hour).UnixNano()), // timeout timestamp is one hour // TODO: get timestamp from consensus state
		)

		if err != nil {
			return true // stop iteration
		}

		balances = balances.Add(coin)
		return false
	})

	if err != nil {
		logger.Error(
			"failed to withdraw IBC vouchers",
			"sender", data.Sender,
			"receiver", data.Receiver,
			"source-port", packet.SourcePort,
			"source-channel", packet.SourceChannel,
			"error", err.Error(),
		)

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
			evmos.ErrKeyTypeNotSupported,
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
