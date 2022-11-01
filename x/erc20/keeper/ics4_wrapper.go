package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/ethereum/go-ethereum/common"
	evmos "github.com/evmos/evmos/v9/types"
	"github.com/evmos/evmos/v9/x/erc20/types"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
)

// SendPacket implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying SendPacket function directly to move down the middleware stack.
func (k Keeper) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI) error {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// NOTE: shouldn't happen as the packet has already
		// been decoded on ICS20 transfer logic
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
	}

	// check if IBC transfer denom is a valid Ethereum contract address
	if !common.IsHexAddress(data.Denom) {
		// no-op
		return nil
	}

	// Return acknowledgement and continue with the next layer of the IBC middleware
	// stack if if:
	// - ERC20s are disabled
	// - The ERC20 contract is not registered as Cosmos coin
	erc20Params := k.GetParams(ctx)
	if !erc20Params.EnableErc20 {
		// no-op
		return nil
	}

	contractAddr := common.HexToAddress(data.Denom)
	if !k.IsERC20Registered(ctx, contractAddr) {
		// no-op
		return nil
	}

	sender, err := evmos.GetEvmosAddressFromBech32(data.Sender)
	if err != nil {
		// NOTE: shouldn't happen as the receiving address has already
		// been validated on ICS20 transfer logic
		return sdkerrors.Wrap(err, "invalid recipient")
	}

	// NOTE: Denom and amount are already validated
	amount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidType, "cannot convert amount to int")
	}

	msg := types.NewMsgConvertERC20(amount, sender, contractAddr, common.BytesToAddress(sender.Bytes()))

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	// Use MsgConvertERC20 to convert the ERC20 to a Cosmos IBC Coin
	if _, err := k.ConvertERC20(sdk.WrapSDKContext(ctx), msg); err != nil {
		return err
	}

	// Continue to the transfer ICS20 logic
	return k.ics4Wrapper.SendPacket(ctx, channelCap, packet)
}

// WriteAcknowledgement implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying WriteAcknowledgement function directly to move down the middleware stack.
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, channelCap, packet, ack)
}

// GetAppVersion returns the underlying application version.
func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
