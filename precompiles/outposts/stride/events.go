package stride

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	"math/big"
)

const (
	// LiquidStakeEvmos is the event type emitted on a Transfer transaction to Autopilot on Stride.
	LiquidStakeEvmos = "LiquidStake"
)

// EmitLiquidStakeEvent creates a new liquid stake event on the EVM stateDB.
func (p Precompile) EmitLiquidStakeEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	senderAddr, erc20Addr common.Address,
	amount *big.Int,
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

	topics[2], err = cmn.MakeTopic(erc20Addr)
	if err != nil {
		return err
	}

	// Prepare the event data: amount
	arguments := abi.Arguments{event.Inputs[2]}
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
