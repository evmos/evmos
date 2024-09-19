// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"encoding/json"
	"github.com/evmos/evmos/v20/precompiles/common"
	"math/big"
	"reflect"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	gethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

const PrecompileAddress = "0x0000000000000000000000000000000000000900"

var EndAuctionEventABI = abi.Event{
	Name:      "AuctionEnd",
	RawName:   "AuctionEnd",
	Anonymous: false,
	Inputs: abi.Arguments{
		abi.Argument{
			Name:    "winner",
			Type:    abi.Type{Size: 20, T: abi.AddressTy},
			Indexed: true,
		},
		abi.Argument{
			Name:    "round",
			Type:    abi.Type{Size: 64, T: abi.UintTy},
			Indexed: true,
		},
		abi.Argument{
			Name: "coins",
			Type: abi.Type{
				Elem: &abi.Type{
					T:            abi.TupleTy,
					TupleRawName: "Coin",
					TupleElems: []*abi.Type{
						{T: abi.StringTy},
						{Size: 256, T: abi.UintTy},
					},
					TupleRawNames: []string{"denom", "amount"},
					TupleType:     reflect.TypeOf(cmn.Coin{}),
				},
				T: abi.SliceTy,
			},
		},
		abi.Argument{
			Name: "burned",
			Type: abi.Type{Size: 256, T: abi.UintTy},
		},
	},
	Sig: "AuctionEnd(address,uint64,(string,uint256)[],uint256)",
	ID:  crypto.Keccak256Hash([]byte("AuctionEnd(address,uint64,(string,uint256)[],uint256)")),
}

// EmitAuctionEndEvent emits an event as an ethereum tx log to be able to filter
// it via the JSON-RPC
func EmitAuctionEndEvent(ctx sdk.Context, winner sdk.AccAddress, round uint64, coins sdk.Coins, burnedAmt *big.Int) error {
	bidWinnerHexAddr := gethcommon.BytesToAddress(winner.Bytes())

	// event topics
	winnerTopic, err := common.MakeTopic(bidWinnerHexAddr)
	if err != nil {
		return errorsmod.Wrapf(err, "failed make log topic")
	}

	roundTopic, err := common.MakeTopic(round)
	if err != nil {
		return errorsmod.Wrapf(err, "failed make log topic")
	}

	// index the bidWinner address and round
	topics := []gethcommon.Hash{
		EndAuctionEventABI.ID,
		winnerTopic,
		roundTopic,
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{EndAuctionEventABI.Inputs[2], EndAuctionEventABI.Inputs[3]}

	// parse coins to use big int instead of sdkmath.Int
	eventCoins := common.NewCoinsResponse(coins)

	packed, err := arguments.Pack(eventCoins, burnedAmt)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to pack log data")
	}

	ethLog := &ethtypes.Log{
		Address:     gethcommon.HexToAddress(PrecompileAddress),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
		BlockHash:   gethcommon.BytesToHash(ctx.HeaderHash()),
	}
	// convert the log to the proto representation
	// to be consistent with the MsgEthTx response log type
	log := evmtypes.NewLogFromEth(ethLog)
	value, err := json.Marshal(log)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to encode log")
	}

	ctx.EventManager().EmitEvents(sdk.Events{sdk.NewEvent(
		evmtypes.EventTypeTxLog,
		sdk.NewAttribute(evmtypes.AttributeKeyTxLog, string(value)),
	)})

	return nil
}
