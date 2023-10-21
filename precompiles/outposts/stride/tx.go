// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// LiquidStakeMethod is the name of the liquidStake method
	LiquidStakeMethod = "liquidStake"
	// RedeemMethod is the name of the redeem method
	RedeemMethod = "redeem"
	// LiquidStakeAction is the action name needed in the memo field
	LiquidStakeAction = "LiquidStake"
	// RedeemAction is the action name needed in the memo field
	RedeemAction = "Redeem"
)

// LiquidStake is a transaction that liquid stakes tokens using
// a ICS20 transfer with a custom memo field that will trigger Stride's Autopilot middleware
func (p Precompile) LiquidStake(
	_ sdk.Context,
	_ common.Address,
	_ vm.StateDB,
	_ *vm.Contract,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	return nil, nil
}

// Redeem is a transaction that redeems the native tokens using the liquid stake
// tokens. It executes a ICS20 transfer with a custom memo field that will
// trigger Stride's Autopilot middleware
func (p Precompile) Redeem(
	_ sdk.Context,
	_ common.Address,
	_ vm.StateDB,
	_ *vm.Contract,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	return nil, nil
}
