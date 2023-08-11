// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package authorization

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// ApprovalEvent is the event emitted on a successful Approve transaction.
type ApprovalEvent struct {
	Grantee  common.Address
	Granter  common.Address
	Coin     *sdk.Coin
	TypeUrls []string
}

// AllowanceChangeEvent is the event emitted on successful IncreaseAllowance or DecreaseAllowance transactions.
type AllowanceChangeEvent struct {
	Grantee  common.Address
	Granter  common.Address
	Values   []*big.Int
	TypeUrls []string
}

// RevocationEvent is the event emitted on a successful Revoke transaction.
type RevocationEvent struct {
	Grantee  common.Address
	Granter  common.Address
	TypeUrls []string
}

// EmitAllowanceChangeEvent creates a new allowance change event emitted on IncreaseAllowance
// and DecreaseAllowance transactions.
func EmitAllowanceChangeEvent(args cmn.EmitEventArgs) error {
	// check if the provided Event is correct type
	allowanceChangeEvent, ok := args.EventData.(AllowanceChangeEvent)
	if !ok {
		return fmt.Errorf("invalid Event type, expecting AllowanceChangeEvent but received %T", args.EventData)
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
	packed, err := arguments.Pack(allowanceChangeEvent.TypeUrls, allowanceChangeEvent.Values)
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
	revocationEvent, ok := args.EventData.(RevocationEvent)
	if !ok {
		return fmt.Errorf("invalid Event type, expecting RevocationEvent but received %T", args.EventData)
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
