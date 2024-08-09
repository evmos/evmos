// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"encoding/json"
	"reflect"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	cmn "github.com/evmos/evmos/v19/precompiles/common"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

//// EmitRoundFinished emits an event for the RoundFinished event.
//func EmitRoundFinished(ctx sdk.Context, evmKeeper evmkeeper.Keeper, stateDB vm.StateDB, roundID *big.Int) error {
//	// Get the precompile instance
//	evmParams := evmKeeper.GetParams(ctx)
//	precompile, found, err := evmKeeper.GetStaticPrecompileInstance(&evmParams, common.HexToAddress(auctionsprecompile.PrecompileAddress))
//	if err != nil || !found {
//		return err
//	}
//
//	// Prepare the events
//	p := precompile.(auctionsprecompile.Precompile)
//	event := p.Events[auctionsprecompile.EventTypeRoundFinished]
//	topics := make([]common.Hash, 1)
//
//	// The first topic is always the signature of the event.
//	topics[0] = event.ID
//
//	// Pack the arguments to be used as the Data field
//	arguments := abi.Arguments{event.Inputs[0]}
//	packed, err := arguments.Pack(roundID)
//	if err != nil {
//		return err
//	}
//
//	stateDB.AddLog(&ethtypes.Log{
//		Address:     p.Address(),
//		Topics:      topics,
//		Data:        packed,
//		BlockNumber: uint64(ctx.BlockHeight()),
//	})
//
//	return nil
//}

var endAuctionAbiEvent = abi.Event{
	Name:      "AuctionEnd",
	RawName:   "AuctionEnd",
	Anonymous: false,
	Inputs: abi.Arguments{
		abi.Argument{
			Name:    "winner",
			Type:    abi.Type{Size: 20, T: 7},
			Indexed: true,
		},
		abi.Argument{
			Name: "coins",
			Type: abi.Type{
				Elem: &abi.Type{
					T:            6,
					TupleRawName: "Coin",
					TupleElems: []*abi.Type{
						{T: 3},
						{Size: 256, T: 1},
					},
					TupleRawNames: []string{"denom", "amount"},
					TupleType:     reflect.TypeOf(cmn.Coin{}),
				},
				T: 4,
			},
		},
		abi.Argument{
			Name: "burned",
			Type: abi.Type{Size: 256, T: 1},
		},
	},
	Sig: "AuctionEnd(address,(string,uint256)[],uint256)",
	ID:  crypto.Keccak256Hash([]byte("AuctionEnd(address,(string,uint256)[],uint256)")),
}

// EmitAuctionEndEvent emits an event as an ethereum tx log to be able to filter
// it via the JSON-RPC
func EmitAuctionEndEvent(ctx sdk.Context, winner sdk.AccAddress, coins sdk.Coins, burnedAmt math.Int) error {
	bidWinnerHexAddr := common.BytesToAddress(winner.Bytes())

	// event topics
	winnerTopic, err := cmn.MakeTopic(bidWinnerHexAddr)
	if err != nil {
		return errorsmod.Wrapf(err, "failed make log topic")
	}

	// index the bidWinner address
	topics := []common.Hash{
		endAuctionAbiEvent.ID,
		winnerTopic,
	}
	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{endAuctionAbiEvent.Inputs[1], endAuctionAbiEvent.Inputs[2]}

	// parse coins to use big int instead of sdkmath.Int
	eventCoins := make([]cmn.Coin, coins.Len())
	for i, c := range coins {
		eventCoins[i].Amount = c.Amount.BigInt()
		eventCoins[i].Denom = c.Denom
	}

	packed, err := arguments.Pack(eventCoins, burnedAmt.BigInt())
	if err != nil {
		return errorsmod.Wrapf(err, "failed to pack log data")
	}

	ethLog := &ethtypes.Log{
		Address:     common.HexToAddress("0x0000000000000000000000000000000000000805"), // ?? set the auctions precompile address in the log. Or should we use the auctions mod address?  TODO: get it from a constaant instead of hardcoded
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	}
	value, err := json.Marshal(ethLog)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to encode log")
	}

	ctx.EventManager().EmitEvents(sdk.Events{sdk.NewEvent(
		evmtypes.EventTypeTxLog,
		sdk.NewAttribute(evmtypes.AttributeKeyTxLog, string(value)),
	)})

	return nil
}
