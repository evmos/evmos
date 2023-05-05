package keeper

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
)

// ValidatedIBCChannelParams does validation of a newly created dIBC channel. A transfer
// channel must be UNORDERED, use the correct port (a smart contract), and use the current
// supported version. Only 2^32 channels are allowed to be created.
func (k Keeper) ValidatedIBCChannelParams(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// NOTE: for escrow address security only 2^32 channels are allowed to be created
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/7737
	channelSequence, err := channeltypes.ParseChannelSequence(channelID)
	if err != nil {
		return err
	}

	if channelSequence > uint64(math.MaxUint32) {
		return errorsmod.Wrapf(transfertypes.ErrMaxTransferChannels, "channel sequence %d is greater than max allowed transfer channels %d", channelSequence, uint64(math.MaxUint32))
	}

	// TODO: it should have a separator between contract addresses?
	hexAddr := portID

	if !common.IsHexAddress(hexAddr) {
		return errorsmod.Wrapf(
			porttypes.ErrInvalidPort,
			"port is not a valid contract hex address: %s", hexAddr,
		)
	}

	address := common.HexToAddress(hexAddr)

	account := k.evmKeeper.GetAccountWithoutBalance(ctx, address)
	if account == nil {
		return errorsmod.Wrapf(
			sdkerrors.ErrUnknownAddress,
			"port identifier's contract was not found: %s", hexAddr,
		)
	}

	if !account.IsContract() {
		return errorsmod.Wrapf(
			porttypes.ErrInvalidPort,
			"port is not a contract: %s", hexAddr,
		)
	}

	return nil
}
