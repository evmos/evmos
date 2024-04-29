// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

const (
	// DepositMethod defines the ABI method name for the IWERC20 deposit
	// transaction.
	DepositMethod = "deposit"
	// WithdrawMethod defines the ABI method name for the IWERC20 withdraw
	// transaction.
	WithdrawMethod = "withdraw"
)

// Deposit is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (eg. EVMOS and WEVMOS). It only emits the Deposit event.
func (p Precompile) Deposit() ([]byte, error) {
	return nil, nil
}

// Withdraw is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (eg. EVMOS and WEVMOS). It only emits the Withdraw event.
func (p Precompile) Withdraw() ([]byte, error) {
	return nil, nil
}
