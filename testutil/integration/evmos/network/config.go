// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"fmt"
	"math/big"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testtx "github.com/evmos/evmos/v18/testutil/tx"
	evmostypes "github.com/evmos/evmos/v18/types"
	"github.com/evmos/evmos/v18/utils"
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

// NewConfig returns a new configuration based on the given
// settings.
func NewConfig(
	chainID string,
	denom string,
	nValidators int,
) Config {
	EIP155ChainID, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		panic(fmt.Sprintf("invalid chainID while setting up integration test config: %s", err))
	}

	// TODO: check if this is actually necessary?
	account, _ := testtx.NewAccAddressAndKey()

	return Config{
		chainID:            chainID,
		eip155ChainID:      EIP155ChainID,
		amountOfValidators: nValidators,
		preFundedAccounts:  []sdktypes.AccAddress{account},
		balances:           nil,
		denom:              denom,
		customGenesisState: nil,
	}
}

// DefaultConfig returns the default configuration for a chain.
func DefaultConfig() Config {
	nVals := 3
	return NewConfig(
		utils.MainnetChainID+"-1",
		utils.BaseDenom,
		nVals,
	)
}

// GetGenAccountsAndBalances takes the network configuration and returns the used
// genesis accounts and balances.
//
// NOTE: If the balances are set, the pre-funded accounts are ignored.
func GetGenAccountsAndBalances(cfg Config) (genAccounts []authtypes.GenesisAccount, balances []banktypes.Balance) {
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

// GetChainID returns the chainID field of the config.
func (cfg Config) GetChainID() string {
	return cfg.chainID
}

// GetEIP155ChainID returns the EIP-155 chainID number.
func (cfg Config) GetEIP155ChainID() *big.Int {
	return cfg.eip155ChainID
}

// GetAmountOfValidators returns the amount of validators field of the config.
func (cfg Config) GetAmountOfValidators() int {
	return cfg.amountOfValidators
}

// GetPreFundedAccounts returns the pre-funded accounts field of the config.
func (cfg Config) GetPreFundedAccounts() []sdktypes.AccAddress {
	return cfg.preFundedAccounts
}

// GetBalances returns the balances field of the config.
func (cfg Config) GetBalances() []banktypes.Balance {
	return cfg.balances
}

// GetDenom returns the denom field of the config.
func (cfg Config) GetDenom() string {
	return cfg.denom
}

// GetCustomGenesisState returns the custom genesis state field of the config.
func (cfg Config) GetCustomGenesisState() CustomGenesisState {
	return cfg.customGenesisState
}
