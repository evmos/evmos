package werc20

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

const (
	// EventTypeDeposit defines the event type for the Deposit transaction.
	EventTypeDeposit = "Deposit"
	// EventTypeWithdraw defines the event type for the Withdraw transaction.
	EventTypeWithdraw = "Withdraw"
)

// EmitDepositEvent creates a new Deposit event emitted on a Deposit transaction.
func (p Precompile) EmitDepositEvent(ctx sdk.Context, stateDB vm.StateDB, dst common.Address, amount *big.Int) error {
	event := p.ABI.Events[EventTypeDeposit]
	return p.createWERC20Event(ctx, stateDB, event, dst, amount)
}

// EmitDepositEvent creates a new Withdraw event emitted on Withdraw transaction.
func (p Precompile) EmitWithdrawEvent(ctx sdk.Context, stateDB vm.StateDB, src common.Address, amount *big.Int) error {
	event := p.ABI.Events[EventTypeWithdraw]
	return p.createWERC20Event(ctx, stateDB, event, src, amount)
}

func (p Precompile) createWERC20Event(
	ctx sdk.Context,
	stateDB vm.StateDB,
	event abi.Event,
	address common.Address,
	amount *big.Int,
) error {
	// Prepare the event topics
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(address)
	if err != nil {
		return err
	}

	bigIntType, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return err
	}

	// Create the ABI arguments
	wadArg := abi.Argument{
		Name:    "wad",
		Type:    bigIntType,
		Indexed: false,
	}

	// Check if the coin is set to infinite
	wad := abi.MaxUint256
	if amount != nil {
		wad = new(big.Int).Set(amount)
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{wadArg}
	packed, err := arguments.Pack(wad)
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
