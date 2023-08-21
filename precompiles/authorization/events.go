// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package authorization

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// EventApproval is the event emitted on a successful Approve transaction.
type EventApproval struct {
	Grantee common.Address
	Granter common.Address
	Value   *big.Int
	Methods []string
}

// EventAllowanceChange is the event emitted on successful IncreaseAllowance or DecreaseAllowance transactions.
type EventAllowanceChange struct {
	Grantee common.Address
	Granter common.Address
	Values  []*big.Int
	Methods []string
}

// EventRevocation is the event emitted on a successful Revoke transaction.
type EventRevocation struct {
	Grantee  common.Address
	Granter  common.Address
	TypeUrls []string
}

// EmitAllowanceChangeEvent creates a new allowance change event emitted on IncreaseAllowance
// and DecreaseAllowance transactions.
func EmitAllowanceChangeEvent(args cmn.EmitEventArgs) error {
	// check if the provided Event is correct type
	allowanceChangeEvent, ok := args.EventData.(EventAllowanceChange)
	if !ok {
		return fmt.Errorf("invalid Event type, expecting EventAllowanceChange but received %T", args.EventData)
	}

	// Prepare the event topics
	event := args.ContractEvents[EventTypeAllowanceChange]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(allowanceChangeEvent.Grantee)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(allowanceChangeEvent.Granter)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3]}
	packed, err := arguments.Pack(allowanceChangeEvent.Methods, allowanceChangeEvent.Values)
	if err != nil {
		return err
	}

	args.StateDB.AddLog(&ethtypes.Log{
		Address:     args.ContractAddr,
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(args.Ctx.BlockHeight()),
	})

	return nil
}

// EmitRevocationEvent creates a new approval event emitted on a Revoke transaction.
func EmitRevocationEvent(args cmn.EmitEventArgs) error {
	// Prepare the event topics
	revocationEvent, ok := args.EventData.(EventRevocation)
	if !ok {
		return fmt.Errorf("invalid Event type, expecting EventRevocation but received %T", args.EventData)
	}
	// Prepare the event topics
	event := args.ContractEvents[EventTypeRevocation]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(revocationEvent.Grantee)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(revocationEvent.Granter)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(revocationEvent.TypeUrls)
	if err != nil {
		return err
	}

	args.StateDB.AddLog(&ethtypes.Log{
		Address:     args.ContractAddr,
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(args.Ctx.BlockHeight()),
	})

	return nil
}
