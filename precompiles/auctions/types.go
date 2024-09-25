// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	"fmt"
	"math/big"

	erc20Keeper "github.com/evmos/evmos/v20/x/erc20/keeper"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/utils"
	auctionstypes "github.com/evmos/evmos/v20/x/auctions/types"
)

type AuctionInfoOutput struct {
	AuctionInfo AuctionInfo
}
type AuctionInfo struct {
	Tokens        []cmn.Coin
	HighestBid    cmn.Coin
	CurrentRound  uint64
	BidderAddress common.Address
}

// NewMsgBid creates a new MsgBid.
func NewMsgBid(args []interface{}) (common.Address, *auctionstypes.MsgBid, error) {
	if len(args) != 2 {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidHexAddress, args[0])
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidAmount, args[1])
	}

	msgBid := &auctionstypes.MsgBid{
		Amount: sdk.Coin{Amount: sdkmath.NewIntFromBigInt(amount), Denom: utils.BaseDenom},
		Sender: sdk.AccAddress(sender.Bytes()).String(),
	}

	return sender, msgBid, nil
}

// NewMsgDepositCoin creates a new MsgDepositCoin.
func NewMsgDepositCoin(args []interface{}, ctx sdk.Context, erc20Keeper erc20Keeper.Keeper) (common.Address, common.Address, *auctionstypes.MsgDepositCoin, error) {
	if len(args) != 3 {
		return common.Address{}, common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidHexAddress, args[0])
	}

	tokenAddress, ok := args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidHexAddress, args[1])
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidAmount, args[2])
	}

	denom, err := erc20Keeper.GetTokenDenom(ctx, tokenAddress)
	if err != nil {
		return common.Address{}, common.Address{}, nil, err
	}

	msgDepositCoin := &auctionstypes.MsgDepositCoin{
		Amount: sdk.Coin{Amount: sdkmath.NewIntFromBigInt(amount), Denom: denom},
		Sender: sdk.AccAddress(sender.Bytes()).String(),
	}

	return sender, tokenAddress, msgDepositCoin, nil
}

// FromResponse populates the AuctionInfoOutput from a QueryCurrentAuctionInfoResponse.
func (ai *AuctionInfoOutput) FromResponse(res *auctionstypes.QueryCurrentAuctionInfoResponse) *AuctionInfoOutput {
	senderBech := sdk.AccAddress(res.BidderAddress)
	senderHex := common.BytesToAddress(senderBech.Bytes())
	ai.AuctionInfo.BidderAddress = senderHex
	ai.AuctionInfo.HighestBid = cmn.Coin{
		Denom:  res.HighestBid.Denom,
		Amount: res.HighestBid.Amount.BigInt(),
	}
	ai.AuctionInfo.CurrentRound = res.CurrentRound
	ai.AuctionInfo.Tokens = make([]cmn.Coin, len(res.Tokens))
	for i, token := range res.Tokens {
		ai.AuctionInfo.Tokens[i] = cmn.Coin{
			Denom:  token.Denom,
			Amount: token.Amount.BigInt(),
		}
	}
	return ai
}

// Pack packs a given slice of abi arguments into a byte array.
func (ai *AuctionInfoOutput) Pack(args abi.Arguments) ([]byte, error) {
	return args.Pack(ai.AuctionInfo)
}
