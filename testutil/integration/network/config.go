package network

import (
	testtx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/utils"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// NetworkConfig defines the configuration for a chain.
// It allows for customization of the network to adjust to
// testing needs.
type NetworkConfig struct {
	chainID            string
	amountOfValidators int
	preFundedAccounts  []sdktypes.AccAddress
	denom              string
}

// DefaultChainConfig returns the default configuration for a chain.
func DefaultChainConfig() NetworkConfig {
	account, _ := testtx.NewAccAddressAndKey()
	return NetworkConfig{
		chainID:            utils.MainnetChainID + "-1",
		amountOfValidators: 3,
		// No funded accounts besides the validators by default
		preFundedAccounts: []sdktypes.AccAddress{account},
		denom:             utils.BaseDenom,
	}
}

// ConfigOption defines a function that can modify the NetworkConfig.
// The purpose of this is to force to be declarative when the default configuration
// requires to be changed.
type ConfigOption func(*NetworkConfig)

// WithChainID sets a custom chainID for the network.
func WithChainID(chainID string) ConfigOption {
	return func(cfg *NetworkConfig) {
		cfg.chainID = chainID
	}
}

// WithAmountOfValidators sets the amount of validators for the network.
func WithAmountOfValidators(amount int) ConfigOption {
	return func(cfg *NetworkConfig) {
		cfg.amountOfValidators = amount
	}
}

// WithPreFundedAccounts sets the pre-funded accounts for the network.
func WithPreFundedAccounts(accounts ...sdktypes.AccAddress) ConfigOption {
	return func(cfg *NetworkConfig) {
		cfg.preFundedAccounts = accounts
	}
}

// WithDenom sets the denom for the network.
func WithDenom(denom string) ConfigOption {
	return func(cfg *NetworkConfig) {
		cfg.denom = denom
	}
}
