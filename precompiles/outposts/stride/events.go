package stride

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

const (
	// LiquidStakeEvmos is the event type emitted on a Transfer transaction to Autopilot on Stride.
	LiquidStakeEvmos = "LiquidStakeEvmos"
)

// EmitIBCTransferEvent creates a new IBC transfer event emitted on a Transfer transaction.
func (p Precompile) EmitIBCTransferEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	senderAddr common.Address,
	receiver string,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	memo string,
) error {
	// Prepare the event topics
	event := p.ABI.Events[LiquidStakeEvmos]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and receiver are indexed
	topics[1], err = cmn.MakeTopic(senderAddr)
	if err != nil {
		return err
	}

	// TODO: Change to custom Coin type or alternatively use uint256 for amount and string for denom
	//topics[2], err = cmn.MakeTopic()
	//if err != nil {
	//	return err
	//}

	// Prepare the event data: denom, amount, memo
	arguments := abi.Arguments{}
	packed, err := arguments.Pack()
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
