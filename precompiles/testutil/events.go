// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	//nolint:stylecheck,revive // it's common practice to use the global imports for Ginkgo and Gomega
	. "github.com/onsi/gomega"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

// CheckAuthorizationEvents is a helper function used in the integration tests and checks if the approval event is emitted.
//
// If the amount is nil, it will not check for the amount in the event, which should be used for any generic approvals.
func CheckAuthorizationEvents(event abi.Event, precompileAddr, granter, grantee common.Address, res abci.ExecTxResult, height int64, msgTypes []string, amount *big.Int) {
	var log evmtypes.Log
	// Tx log is always the second last event
	txLogAttributes := res.Events[len(res.Events)-2].Attributes
	attr := txLogAttributes[0]

	err := json.Unmarshal([]byte(attr.Value), &log)
	Expect(err).To(BeNil(), "failed to unmarshal log")

	// Check the key of the log is the expected one
	Expect(attr.Key).To(Equal(evmtypes.AttributeKeyTxLog), "expected different key for log")

	// Check if the log has the expected indexed fields and data
	Expect(log.Address).To(Equal(precompileAddr.String()), "expected different address in event")
	Expect(log.BlockNumber).To(Equal(uint64(height)), "expected different block number in event") //nolint:gosec // G115
	Expect(log.Topics[0]).To(Equal(event.ID.String()), "expected different event ID")
	Expect(common.HexToAddress(log.Topics[1])).To(Equal(grantee), "expected different grantee in event")
	Expect(common.HexToAddress(log.Topics[2])).To(Equal(granter), "expected different granter in event")

	// Unpack the arguments from the Data field
	arguments := make(abi.Arguments, 0, len(event.Inputs))
	arguments = append(arguments, event.Inputs...)
	unpackedData, err := arguments.Unpack(log.Data)
	Expect(err).To(BeNil(), "failed to unpack log data")

	Expect(unpackedData[0]).To(Equal(msgTypes), "expected different message types in event")
	if amount != nil {
		Expect(len(unpackedData)).To(Equal(2), "expected different number of arguments in event")
		Expect(unpackedData[1]).To(Equal(amount), "expected different amount in event")
	}
}

// validateEvents checks if the provided event names are included as keys in the contract events.
func validateEvents(contractEvents map[string]abi.Event, events []string) ([]abi.Event, error) {
	expEvents := make([]abi.Event, 0, len(events))
	for _, eventStr := range events {
		event, found := contractEvents[eventStr]
		if !found {
			availableABIEvents := make([]string, 0, len(contractEvents))
			for event := range contractEvents {
				availableABIEvents = append(availableABIEvents, event)
			}
			availableABIEventsStr := strings.Join(availableABIEvents, ", ")
			return nil, fmt.Errorf("unknown event %q is not contained in given ABI events:\n%s", eventStr, availableABIEventsStr)
		}
		expEvents = append(expEvents, event)
	}
	return expEvents, nil
}
