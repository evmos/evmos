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
// ERC-20 (e.g. EVMOS and WEVMOS).
func (p Precompile) Deposit() ([]byte, error) {
	// TODO: Shouldn't this raise an error now? It's a no-op right now, but that invites users to waste gas without anything happening. Wouldn't an error be more clear that this is not useful anymore?
	return nil, nil
}

// Withdraw is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (e.g. EVMOS and WEVMOS).
func (p Precompile) Withdraw() ([]byte, error) {
	return nil, nil
}
