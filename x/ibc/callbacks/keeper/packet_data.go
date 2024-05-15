package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ICS20EVMPacketData struct {
	Denom    string
	Amount   *big.Int
	Sender   common.Address
	Receiver string
	Memo     string
}

// DecodeTransferPacketData decodes the packet data from a FungibleToken transfer.
func (k Keeper) DecodeTransferPacketData(packetData []byte) (ICS20EVMPacketData, error) {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetData, &data); err != nil {
		err = errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return ICS20EVMPacketData{}, err
	}

	amount := new(big.Int)
	amount.SetString(data.Amount, 10)

	hexAddrBytes, err := sdk.GetFromBech32(data.Sender, "evmos")
	if err != nil {
		fmt.Println("the error in conversion is: ", err)
		return ICS20EVMPacketData{}, err
	}

	return ICS20EVMPacketData{
		Denom:    data.Denom,
		Amount:   amount,
		Sender:   common.BytesToAddress(hexAddrBytes),
		Receiver: data.Receiver,
		Memo:     data.Memo,
	}, nil
}
