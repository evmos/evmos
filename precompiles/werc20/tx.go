// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmos "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// DepositMethod defines the ABI method name for the IWERC20 deposit
	// transaction.
	DepositMethod = "deposit"
	// WithdrawMethod defines the ABI method name for the IWERC20 withdraw
	// transaction.
	WithdrawMethod = "withdraw"
)

// Deposit handles the payable deposit function. It retrieves the deposited amount
// and sends it back to the sender using the bank keeper.
func (p Precompile) Deposit(ctx sdk.Context, contract *vm.Contract) ([]byte, error) {
	// Get the deposited amount from the contract value
	depositedAmount := contract.Value()

	// Get the sender's address
	sender := common.BytesToAddress(contract.Caller().Bytes())
	senderAccAddress := sdk.AccAddress(sender.Bytes())
	precompileAccAddr := sdk.AccAddress(p.Address().Bytes())

	// Send the coins back to the sender
	err := p.bankKeeper.SendCoins(ctx, precompileAccAddr, senderAccAddress, sdk.NewCoins(sdk.NewCoin(evmos.BaseDenom, math.NewIntFromBigInt(depositedAmount))))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Withdraw is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (e.g. EVMOS and WEVMOS).
func (p Precompile) Withdraw() ([]byte, error) {
	return nil, nil
}
