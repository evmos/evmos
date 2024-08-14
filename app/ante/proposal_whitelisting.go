package ante

import (
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// IsModuleWhiteList This is used for non-legacy gov transactions
// Returning true cause all txs are whitelisted
func IsModuleWhiteList(_ string) bool {
	return true
}

// IsProposalWhitelisted This is used for legacy gov transactions
// Returning true cause all txs are whitelisted
func IsProposalWhitelisted(_ govv1beta1.Content) bool {
	return true
}
