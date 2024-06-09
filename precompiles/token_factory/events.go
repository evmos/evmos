package tokenfactory

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"math/big"
)

const (
	EventCreateERC20 = "ERC20Created"
	EventMint        = "Mint"
)

// EmitCreateERC20Event emits an event when a new ERC20 token is created
func (p Precompile) EmitCreateERC20Event(ctx sdk.Context, stateDB vm.StateDB, creator, tokenAddress common.Address, name, symbol string, decimals uint8, initialSupply *big.Int) error {
	events := p.ABI.Events[EventCreateERC20]
	topics := make([]common.Hash, 7)

	var err error
	// The first topic is always the signature of the event.
	topics[0] = events.ID

	topics[1], err = cmn.MakeTopic(creator)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{events.Inputs[2], events.Inputs[3], events.Inputs[4], events.Inputs[5], events.Inputs[6]}
	packed, err := arguments.Pack(name, symbol, decimals, initialSupply, tokenAddress)
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

// EmitEventMint creates a new Transfer event emitted on transfer and transferFrom transactions.
func (p Precompile) EmitEventMint(ctx sdk.Context, stateDB vm.StateDB, to common.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventMint]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(to)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[1]}
	packed, err := arguments.Pack(amount)
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
