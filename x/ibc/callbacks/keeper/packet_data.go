package keeper

import (
	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// DecodeTransferPacketData decodes the packet data from a FungibleToken transfer.
func (k Keeper) DecodeTransferPacketData(packetData []byte) (transfertypes.FungibleTokenPacketData, error) {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetData, &data); err != nil {
		// NOTE: shouldn't happen as the packet has already
		// been decoded on ICS20 transfer logic
		err = errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return transfertypes.FungibleTokenPacketData{}, err
	}

	return data, nil
}
