// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
// Osmosis package contains the logic of the Osmosis outpost on the Evmos chain.
// This outpost uses the ics20 precompile to relay IBC packets to the Osmosis
// chain, targeting the XCSV
package osmosis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// SwapMethod is the name of the swap method
	SwapMethod = "swap"
	// SwapAction is the action name needed in the memo field
	SwapAction = "Swap"
)

// Swap is a transaction that swap tokens on the Osmosis chain using
// an ICS20 transfer with a custom memo field to trigger the XCS V2 contract.
func (p Precompile) Swap(
	_ sdk.Context,
	_ common.Address,
	_ vm.StateDB,
	_ *vm.Contract,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	return nil, nil
}
