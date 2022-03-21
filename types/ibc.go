package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// GetTransferSenderRecipient returns the sender and recipient sdk.AccAddresses
// from an ICS20 FungibleTokenPacketData as well as the original sender bech32
// address from the packet data. This function fails if:
// - the packet data is not FungibleTokenPacketData
// - sender address is invalid
// - recipient address is invalid
func GetTransferSenderRecipient(packet channeltypes.Packet) (sender, recipient sdk.AccAddress, senderBech32 string, err error) {
	// unmarshal packet data to obtain the sender and recipient
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return nil, nil, "", sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
	}

	// validate the sender bech32 address from the counterparty chain
	bech32Prefix := strings.Split(data.Sender, "1")[0]
	if bech32Prefix == data.Sender {
		return nil, nil, "", sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender: %s", data.Sender)
	}

	senderBz, err := sdk.GetFromBech32(data.Sender, bech32Prefix)
	if err != nil {
		return nil, nil, "", sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender %s, %s", data.Sender, err.Error())
	}

	// change the bech32 human readable prefix (HRP) of the sender to `evmos1`
	sender = sdk.AccAddress(senderBz)

	// obtain the evmos recipient address
	recipient, err = sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return nil, nil, "", sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address %s", err.Error())
	}

	return sender, recipient, data.Sender, nil
}
