package keeper

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	evmos "github.com/tharsis/evmos/v2/types"
)

// OnRecvPacket performs an IBC receive callback. It returns the tokens that
// users transferred to their Cosmos secp256k1 address instead of the Ethereum
// ethsecp256k1 address.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)

	params := k.GetParams(ctx)
	claimsParams := k.claimsKeeper.GetParams(ctx)

	// check channels from this chain (i.e destination)
	if !params.EnableWithdraw ||
		!claimsParams.IsAuthorizedChannel(packet.DestinationChannel) ||
		claimsParams.IsEVMChannel(packet.DestinationChannel) {
		// return original ACK if:
		// - withdraw is disabled globally
		// - channel is not authorized
		// - channel is an EVM channel
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

	// case 1: sender ≠ recipient.
	// Withdraw is only possible for addresses in which the sender = recipient.
	// Continue to the next IBC middleware by returning the original ACK.
	if !sender.Equals(recipient) {
		return ack
	}

	// get the recipient account
	account := k.accountKeeper.GetAccount(ctx, recipient)

	// Case 2. recipient pubkey is a supported key (eth_secp256k1, amino multisig, ed25519)
	// ==> Continue and return success ACK as the funds are not stuck on chain
	if account != nil &&
		account.GetPubKey() != nil &&
		evmos.IsSupportedKey(account.GetPubKey()) {
		return ack
	}

	// NOTE: Since destination channel is authorized and not from an EVM chain, we know that
	// only secp256k1 keys are supported in the source chain. This means that we can now
	// initiate the withdraw logic

	// transfer the balance back to the sender address
	destPort := packet.DestinationPort
	destChannel := packet.DestinationChannel
	balances := sdk.Coins{}

	k.bankKeeper.IterateAccountBalances(ctx, recipient, func(coin sdk.Coin) (stop bool) {
		if coin.IsZero() {
			// continue
			return false
		}

		switch strings.HasPrefix(coin.Denom, "ibc/") {
		case true:
			// IBC vouchers, obtain the source port and channel from the denom path
			destPort, destChannel, err = k.GetIBCDenomSelfIdentifiers(ctx, coin.Denom, data.Sender)
		default:
			// Native tokens, use the source port and channel to transfer the EVMOS and
			// other converted ERC20 coin denoms to the authorized source chain
			destPort = packet.DestinationPort
			destChannel = packet.DestinationChannel
		}

		if err != nil {
			logger.Error(
				"failed to get the IBC full denom path of source chain",
				"error", err.Error(),
			)
			return true // stop iteration
		}

		// NOTE: only withdraw the IBC tokens from the enabled destination channel
		if packet.DestinationPort != destPort || packet.DestinationChannel != destChannel {
			// continue
			return false
		}

		// NOTE: Don't use the consensus state because it may become unreliable if updates slow down
		timeout := uint64(ctx.BlockTime().Add(4 * time.Hour).UnixNano())

		// Withdraw the tokens to the bech32 prefixed address of the source chain
		err = k.transferKeeper.SendTransfer(
			ctx,
			packet.DestinationPort,    // packet destination port is now the source
			packet.DestinationChannel, // packet destination channel is now the source
			coin,                      // balance of the coin
			recipient,                 // transfer recipient is now the sender
			data.Sender,               // transfer sender is now the recipient
			clienttypes.ZeroHeight(),  // timeout height disabled
			timeout,                   // timeout timestamp is 4 hours from now
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
