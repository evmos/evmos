package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/v3/ibc"
	evmos "github.com/tharsis/evmos/v3/types"
	"github.com/tharsis/evmos/v3/x/recovery/types"
)

// OnRecvPacket performs an IBC receive callback. It returns the tokens that
// users transferred to their Cosmos secp256k1 address instead of the Ethereum
// ethsecp256k1 address. The expected behavior is as follows:
//
// First transfer from authorized source chain:
//  - sends back IBC tokens which originated from the source chain
//  - sends over all Evmos native tokens
// Second transfer from a different authorized source chain:
//  - only sends back IBC tokens which originated from the source chain
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)

	params := k.GetParams(ctx)
	claimsParams := k.claimsKeeper.GetParams(ctx)

	// Check and return original ACK if:
	//  - recovery is disabled globally
	//  - channel is not authorized
	//  - channel is an EVM channel
	if !params.EnableRecovery ||
		!claimsParams.IsAuthorizedChannel(packet.DestinationChannel) ||
		claimsParams.IsEVMChannel(packet.DestinationChannel) {
		return ack
	}

	// Get addresses in `evmos1` and the original bech32 format
	sender, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// return error ACK if the address is on the deny list
	if k.bankKeeper.BlockedAddr(sender) || k.bankKeeper.BlockedAddr(recipient) {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				types.ErrBlockedAddress,
				"sender (%s) or recipient (%s) address are in the deny list for sending and receiving transfers",
				senderBech32, recipientBech32,
			).Error(),
		)
	}

	// Check if sender != recipient, as recovery is only possible for transfers to
	// a sender's own account on Evmos (sender == recipient)
	if !sender.Equals(recipient) {
		// Continue to the next IBC middleware by returning the original ACK.
		return ack
	}

	// get the recipient/sender account
	account := k.accountKeeper.GetAccount(ctx, recipient)

	// recovery is not supported for vesting or module accounts
	if _, isVestingAcc := account.(vestexported.VestingAccount); isVestingAcc {
		return ack
	}

	if _, isModuleAccount := account.(authtypes.ModuleAccountI); isModuleAccount {
		return ack
	}

	// Check if recipient pubkey is a supported key (eth_secp256k1, amino multisig,
	// ed25519). Continue and return success ACK as the funds are not stuck on
	// chain for supported keys
	if account != nil && evmos.IsSupportedKey(account.GetPubKey()) {
		return ack
	}

	// Perform recovery to transfer the balance back to the sender bech32 address.
	// NOTE: Since destination channel is authorized and not from an EVM chain, we
	// know that only secp256k1 keys are supported in the source chain.
	var destPort, destChannel string
	balances := sdk.Coins{}

	// iterate over all tokens owned by the address (i.e recipient balance) and
	// transfer them to the original sender address in the source chain (if
	// applicable, see cases for IBC vouchers below).
	k.bankKeeper.IterateAccountBalances(ctx, recipient, func(coin sdk.Coin) (stop bool) {
		if coin.IsZero() {
			// safety check: continue
			return false
		}

		if strings.HasPrefix(coin.Denom, "ibc/") {
			// IBC vouchers, obtain the destination port and channel from the denom path
			destPort, destChannel, err = k.GetIBCDenomDestinationIdentifiers(ctx, coin.Denom, senderBech32)
			if err != nil {
				logger.Error(
					"failed to get the IBC full denom path of source chain",
					"error", err.Error(),
				)
				return true // stop iteration
			}

			// NOTE: only recover the IBC tokens from the source chain connected
			// through our authorized destination channel
			if packet.DestinationPort != destPort || packet.DestinationChannel != destChannel {
				// continue
				return false
			}
		}

		// NOTE: Don't use the consensus state because it may become unreliable if updates slow down
		timeout := uint64(ctx.BlockTime().Add(params.PacketTimeoutDuration).UnixNano())

		// Recover the tokens to the bech32 prefixed address of the source chain
		err = k.transferKeeper.SendTransfer(
			ctx,
			packet.DestinationPort,    // packet destination port is now the source
			packet.DestinationChannel, // packet destination channel is now the source
			coin,                      // balance of the coin
			recipient,                 // recipient is the address in the Evmos chain
			senderBech32,              // transfer to your own account address on the source chain
			clienttypes.ZeroHeight(),  // timeout height disabled
			timeout,                   // timeout timestamp is 4 hours from now
		)

		if err != nil {
			return true // stop iteration
		}

		balances = balances.Add(coin)
		return false
	})

	// check error from the iteration above
	if err != nil {
		logger.Error(
			"failed to recover IBC vouchers",
			"sender", senderBech32,
			"receiver", recipientBech32,
			"source-port", packet.SourcePort,
			"source-channel", packet.SourceChannel,
			"error", err.Error(),
		)

		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				err,
				"failed to recover IBC vouchers back to sender '%s' in the corresponding IBC chain", senderBech32,
			).Error(),
		)
	}

	if balances.IsZero() {
		// short circuit in case the user doesn't have any balance
		return ack
	}

	amtStr := balances.String()

	logger.Info(
		"balances recovered to sender address",
		"sender", senderBech32,
		"receiver", recipientBech32,
		"amount", amtStr,
		"source-port", packet.SourcePort,
		"source-channel", packet.SourceChannel,
		"dest-port", packet.DestinationPort,
		"dest-channel", packet.DestinationChannel,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecovery,
			sdk.NewAttribute(sdk.AttributeKeySender, senderBech32),
			sdk.NewAttribute(transfertypes.AttributeKeyReceiver, recipientBech32),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amtStr),
			sdk.NewAttribute(channeltypes.AttributeKeySrcChannel, packet.SourceChannel),
			sdk.NewAttribute(channeltypes.AttributeKeySrcPort, packet.SourcePort),
			sdk.NewAttribute(channeltypes.AttributeKeyDstPort, packet.DestinationPort),
			sdk.NewAttribute(channeltypes.AttributeKeyDstChannel, packet.DestinationChannel),
		),
	)

	// return original acknowledgement
	return ack
}

// GetIBCDenomDestinationIdentifiers returns the destination port and channel of
// the IBC denomination, i.e port and channel on Evmos for the voucher. It
// returns an error if:
//  - the denomination is invalid
//  - the denom trace is not found on the store
//  - destination port or channel ID are invalid
func (k Keeper) GetIBCDenomDestinationIdentifiers(ctx sdk.Context, denom, sender string) (destinationPort, destinationChannel string, err error) {
	ibcDenom := strings.SplitN(denom, "/", 2)
	if len(ibcDenom) < 2 {
		return "", "", sdkerrors.Wrap(transfertypes.ErrInvalidDenomForTransfer, denom)
	}

	hash, err := transfertypes.ParseHexHash(ibcDenom[1])
	if err != nil {
		return "", "", sdkerrors.Wrapf(
			err,
			"failed to recover IBC vouchers back to sender '%s' in the corresponding IBC chain", sender,
		)
	}

	denomTrace, found := k.transferKeeper.GetDenomTrace(ctx, hash)
	if !found {
		return "", "", sdkerrors.Wrapf(
			transfertypes.ErrTraceNotFound,
			"failed to recover IBC vouchers back to sender '%s' in the corresponding IBC chain", sender,
		)
	}

	path := strings.Split(denomTrace.Path, "/")
	if len(path)%2 != 0 {
		// safety check: shouldn't occur
		return "", "", sdkerrors.Wrapf(
			transfertypes.ErrInvalidDenomForTransfer,
			"invalid denom (%s) trace path %s", denomTrace.BaseDenom, denomTrace.Path,
		)
	}

	destinationPort = path[0]
	destinationChannel = path[1]

	_, found = k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return "", "", sdkerrors.Wrapf(
			channeltypes.ErrChannelNotFound,
			"port ID %s, channel ID %s", destinationPort, destinationChannel,
		)
	}

	// NOTE: optimistic handshakes could cause unforeseen issues.
	// Safety check: verify that the destination port and channel are valid
	if err := host.PortIdentifierValidator(destinationPort); err != nil {
		// shouldn't occur
		return "", "", sdkerrors.Wrapf(
			host.ErrInvalidID,
			"invalid port ID '%s': %s", destinationPort, err.Error(),
		)
	}

	if err := host.ChannelIdentifierValidator(destinationChannel); err != nil {
		// shouldn't occur
		return "", "", sdkerrors.Wrapf(
			channeltypes.ErrInvalidChannelIdentifier,
			"channel ID '%s': %s", destinationChannel, err.Error(),
		)
	}

	return destinationPort, destinationChannel, nil
}
