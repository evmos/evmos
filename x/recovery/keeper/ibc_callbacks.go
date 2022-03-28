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
// - sends back IBC tokens which originated from the source chain
// - sends over all Evmos native tokens
// Second transfer from a different authorized source chain:
// - only sends back IBC tokens which originated from the source chain
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)

	params := k.GetParams(ctx)
	claimsParams := k.claimsKeeper.GetParams(ctx)

	// check channels from this chain (i.e destination)
	if !params.EnableRecovery ||
		!claimsParams.IsAuthorizedChannel(packet.DestinationChannel) ||
		claimsParams.IsEVMChannel(packet.DestinationChannel) {
		// return original ACK if:
		// - recovery is disabled globally
		// - channel is not authorized
		// - channel is an EVM channel
		return ack
	}

	sender, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// return error ACK if the address is in the deny list
	if k.bankKeeper.BlockedAddr(sender) || k.bankKeeper.BlockedAddr(recipient) {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				types.ErrBlockedAddress,
				"sender (%s) or recipient (%s) address are in the deny list for sending and receiving transfers",
				senderBech32, recipientBech32,
			).Error(),
		)
	}

	// case 1: sender ≠ recipient.
	// Recovery is only possible for addresses in which the sender = recipient
	// (i.e transferring to your own account in Evmos).
	if !sender.Equals(recipient) {
		// Continue to the next IBC middleware by returning the original ACK.
		return ack
	}

	// get the sender account
	account := k.accountKeeper.GetAccount(ctx, sender)

	// recovery is not supported for vesting or module accounts
	_, isVestingAcc := account.(vestexported.VestingAccount)
	if isVestingAcc {
		return ack
	}

	_, isModuleAccount := account.(authtypes.ModuleAccountI)
	if isModuleAccount {
		return ack
	}

	// Case 2. sender pubkey is a supported key (eth_secp256k1, amino multisig, ed25519)
	// ==> Continue and return success ACK as the funds are not stuck on chain
	if account != nil && evmos.IsSupportedKey(account.GetPubKey()) {
		return ack
	}

	// NOTE: Since destination channel is authorized and not from an EVM chain, we know that
	// only secp256k1 keys are supported in the source chain. This means that we can now
	// initiate the recovery logic

	// transfer the balance back to the sender address
	destPort := packet.DestinationPort
	destChannel := packet.DestinationChannel
	balances := sdk.Coins{}

	// iterate over all the tokens owned by the address (i.e sender balance) and
	// transfer them to the original sender address in the source chain (if
	// applicable, see cases for IBC vouchers below).
	k.bankKeeper.IterateAccountBalances(ctx, sender, func(coin sdk.Coin) (stop bool) {
		if coin.IsZero() {
			// safety check: continue
			return false
		}

		if strings.HasPrefix(coin.Denom, "ibc/") {
			// IBC vouchers, obtain the source port and channel from the denom path
			destPort, destChannel, err = k.GetIBCDenomDestinationIdentifiers(ctx, coin.Denom, senderBech32)
			if err != nil {
				logger.Error(
					"failed to get the IBC full denom path of source chain",
					"error", err.Error(),
				)
				return true // stop iteration
			}

			// NOTE: only recover the IBC tokens from the source chain connected through our
			// authorized destination channel
			if packet.DestinationPort != destPort || packet.DestinationChannel != destChannel {
				// continue
				return false
			}
		} else {
			// Native tokens, use the source port and channel to transfer the EVMOS and
			// other converted ERC20 coin denoms to the authorized source chain
			destPort = packet.DestinationPort
			destChannel = packet.DestinationChannel
		}

		// NOTE: Don't use the consensus state because it may become unreliable if updates slow down
		timeout := uint64(ctx.BlockTime().Add(params.PacketTimeoutDuration).UnixNano())

		// Recovery the tokens to the bech32 prefixed address of the source chain
		err = k.transferKeeper.SendTransfer(
			ctx,
			packet.DestinationPort,    // packet destination port is now the source
			packet.DestinationChannel, // packet destination channel is now the source
			coin,                      // balance of the coin
			sender,                    // sender is the address in the Evmos chain
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
			sdk.NewAttribute(channeltypes.AttributeKeySrcPort, packet.SourcePort),
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
