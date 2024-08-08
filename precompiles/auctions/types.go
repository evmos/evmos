package auctions

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/utils"
	auctionstypes "github.com/evmos/evmos/v19/x/auctions/types"
	"math/big"
)

// NewMsgBid creates a new MsgBid.
func NewMsgBid(args []interface{}) (common.Address, *auctionstypes.MsgBid) {
	if len(args) != 2 {
		return common.Address{}, nil
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil
	}

	msgBid := &auctionstypes.MsgBid{
		Amount: types.Coin{Amount: sdkmath.NewIntFromBigInt(amount), Denom: utils.BaseDenom},
		Sender: sdk.AccAddress(sender.Bytes()).String(),
	}

	return sender, msgBid
}

func NewMsgDepositCoin(args []interface{}) (common.Address, *auctionstypes.MsgDepositCoin) {
	if len(args) != 3 {
		return common.Address{}, nil
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil
	}

	denom, ok := args[1].(string)
	if !ok {
		return common.Address{}, nil
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, nil
	}

	msgDepositCoin := &auctionstypes.MsgDepositCoin{
		Amount: types.Coin{Amount: sdkmath.NewIntFromBigInt(amount), Denom: denom},
		Sender: sdk.AccAddress(sender.Bytes()).String(),
	}

	return sender, msgDepositCoin
}
