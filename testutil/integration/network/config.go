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

// TODO: Create a function that creates an account an keeps a counter of the sequences

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

// ConfigModifier defines a function that can modify the NetworkConfig.
// The purpose of this is to force to be declarative when the default configuration
// requires to be changed.
//type ConfigModifier func(*NetworkConfig)

func (cfg *NetworkConfig) WithChainID(chainID string) {
	cfg.chainID = chainID
}

func (cfg *NetworkConfig) WithDenom(denom string) {
	cfg.denom = denom
}

func (cfg *NetworkConfig) WithAmountOfValidators(amount int) {
	cfg.amountOfValidators = amount
}

func (cfg *NetworkConfig) WithPreFundedAccounts(accounts []sdktypes.AccAddress) {
	cfg.preFundedAccounts = accounts
}

//// Option functions to set specific options
//func WithChainID(chainID string) ConfigModifier {
//	return func(cfg *NetworkConfig) {
//		cfg.chainID = chainID
//	}
//}

//// Option functions to set specific options
//func WithAmountOfValidators(amount int) ConfigModifier {
//	return func(cfg *NetworkConfig) {
//		cfg.AmountOfValidators = amount
//	}
//}
//
//// Option functions to set specific options
//func WithPreFundedAccounts(accounts ...sdktypes.AccAddress) ConfigModifier {
//	return func(cfg *NetworkConfig) {
//		cfg.PreFundedAccounts = accounts
//	}
//}
