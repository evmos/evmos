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
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

const (
	// EventTypeIBCTransferAuthorization defines the event type for the ICS20 TransferAuthorization transaction.
	EventTypeIBCTransferAuthorization = "IBCTransferAuthorization" //#nosec G101 -- no hardcoded credentials here
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

// EmitIBCTransferAuthorizationEvent creates a new IBC transfer authorization event emitted on a TransferAuthorization transaction.
func EmitIBCTransferAuthorizationEvent(
	event abi.Event,
	ctx sdk.Context,
	stateDB vm.StateDB,
	precompileAddr, granteeAddr, granterAddr common.Address,
	allocations []cmn.ICS20Allocation,
) error {
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(granteeAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(granterAddr)
	if err != nil {
		return err
	}

	// Prepare the event data: sourcePort, sourceChannel, denom, amount
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(allocations)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     precompileAddr,
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
