package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	evmos "github.com/evmos/evmos/v6/types"
)

// GetTransferSenderRecipient returns the sender and recipient sdk.AccAddresses
// from an ICS20 FungibleTokenPacketData as well as the original sender bech32
// address from the packet data. This function fails if:
//  - the packet data is not FungibleTokenPacketData
//  - sender address is invalid
//  - recipient address is invalid
func GetTransferSenderRecipient(packet channeltypes.Packet) (
	sender, recipient sdk.AccAddress,
	senderBech32, recipientBech32 string,
	err error,
) {
	// unmarshal packet data to obtain the sender and recipient
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return nil, nil, "", "", sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
	}

	// validate the sender bech32 address from the counterparty chain
	// and change the bech32 human readable prefix (HRP) of the sender to `evmos`
	sender, err = evmos.GetEvmosAddressFromBech32(data.Sender)
	if err != nil {
		return nil, nil, "", "", sdkerrors.Wrap(err, "invalid sender")
	}

	// validate the recipient bech32 address from the counterparty chain
	// and change the bech32 human readable prefix (HRP) of the recipient to `evmos`
	recipient, err = evmos.GetEvmosAddressFromBech32(data.Receiver)
	if err != nil {
		return nil, nil, "", "", sdkerrors.Wrap(err, "invalid recipient")
	}

	return sender, recipient, data.Sender, data.Receiver, nil
}

// GetTransferAmount returns the amount from an ICS20 FungibleTokenPacketData.
func GetTransferAmount(packet channeltypes.Packet) (string, error) {
	// unmarshal packet data to obtain the sender and recipient
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return "", sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
	}

	if data.Amount == "" {
		return "", sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "empty amount")
	}

	if _, ok := sdk.NewIntFromString(data.Amount); !ok {
		return "", sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount")
	}

	return data.Amount, nil
}
