// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"math/big"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testtx "github.com/evmos/evmos/v16/testutil/tx"
	evmostypes "github.com/evmos/evmos/v16/types"
	"github.com/evmos/evmos/v16/utils"
)

// Config defines the configuration for a chain.
// It allows for customization of the network to adjust to
// testing needs.
type Config struct {
	chainID            string
	eip155ChainID      *big.Int
	amountOfValidators int
	preFundedAccounts  []sdktypes.AccAddress
	balances           []banktypes.Balance
	denom              string
	customGenesisState CustomGenesisState
}

type CustomGenesisState map[string]interface{}

// DefaultConfig returns the default configuration for a chain.
func DefaultConfig() Config {
	account, _ := testtx.NewAccAddressAndKey()
	return Config{
		chainID:            utils.MainnetChainID + "-1",
		eip155ChainID:      big.NewInt(9001),
		amountOfValidators: 3,
		// No funded accounts besides the validators by default
		preFundedAccounts: []sdktypes.AccAddress{account},
		// NOTE: Per default, the balances are left empty, and the pre-funded accounts are used.
		balances:           nil,
		denom:              utils.BaseDenom,
		customGenesisState: nil,
	}
}

// getGenAccountsAndBalances takes the network configuration and returns the used
// genesis accounts and balances.
//
// NOTE: If the balances are set, the pre-funded accounts are ignored.
func getGenAccountsAndBalances(cfg Config) (genAccounts []authtypes.GenesisAccount, balances []banktypes.Balance) {
	if len(cfg.balances) > 0 {
		balances = cfg.balances
		accounts := getAccAddrsFromBalances(balances)
		genAccounts = createGenesisAccounts(accounts)
	} else {
		coin := sdktypes.NewCoin(cfg.denom, PrefundedAccountInitialBalance)
		genAccounts = createGenesisAccounts(cfg.preFundedAccounts)
		balances = createBalances(cfg.preFundedAccounts, coin)
	}

	return
}

// ConfigOption defines a function that can modify the NetworkConfig.
// The purpose of this is to force to be declarative when the default configuration
// requires to be changed.
type ConfigOption func(*Config)

// WithChainID sets a custom chainID for the network. It panics if the chainID is invalid.
func WithChainID(chainID string) ConfigOption {
	chainIDNum, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		panic(err)
	}
	return func(cfg *Config) {
		cfg.chainID = chainID
		cfg.eip155ChainID = chainIDNum
	}
}

// WithAmountOfValidators sets the amount of validators for the network.
func WithAmountOfValidators(amount int) ConfigOption {
	return func(cfg *Config) {
		cfg.amountOfValidators = amount
	}
}

// WithPreFundedAccounts sets the pre-funded accounts for the network.
func WithPreFundedAccounts(accounts ...sdktypes.AccAddress) ConfigOption {
	return func(cfg *Config) {
		cfg.preFundedAccounts = accounts
	}
}

// WithBalances sets the specific balances for the pre-funded accounts, that
// are being set up for the network.
func WithBalances(balances ...banktypes.Balance) ConfigOption {
	return func(cfg *Config) {
		cfg.balances = append(cfg.balances, balances...)
	}
}

// WithDenom sets the denom for the network.
func WithDenom(denom string) ConfigOption {
	return func(cfg *Config) {
		cfg.denom = denom
	}
}

// WithCustomGenesis sets the custom genesis of the network for specific modules.
func WithCustomGenesis(customGenesis CustomGenesisState) ConfigOption {
	return func(cfg *Config) {
		cfg.customGenesisState = customGenesis
	}
}
