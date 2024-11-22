// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	testtx "github.com/evmos/evmos/v20/testutil/tx"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/utils"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
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
	chainCoins         ChainCoins
	customGenesisState CustomGenesisState
	otherCoinDenom     []string
	operatorsAddrs     []sdktypes.AccAddress
	customBaseAppOpts  []func(*baseapp.BaseApp)
}

type CustomGenesisState map[string]interface{}

// DefaultConfig returns the default configuration for a chain.
func DefaultConfig() Config {
	account, _ := testtx.NewAccAddressAndKey()

	// Default chainID is mainnet.
	chainID := utils.MainnetChainID + "-1"
	eip155ChainID, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		panic("chain ID with invalid eip155 value")
	}
	return Config{
		chainID:            chainID,
		eip155ChainID:      eip155ChainID,
		chainCoins:         DefaultChainCoins(),
		amountOfValidators: 3,
		// Only one account besides the validators
		preFundedAccounts: []sdktypes.AccAddress{account},
		// NOTE: Per default, the balances are left empty, and the pre-funded accounts are used.
		balances:           nil,
		customGenesisState: nil,
	}
}

// getGenAccountsAndBalances takes the network configuration and returns the used
// genesis accounts and balances.
//
// NOTE: If the balances are set, the pre-funded accounts are ignored.
func getGenAccountsAndBalances(cfg Config, validators []stakingtypes.Validator) (genAccounts []authtypes.GenesisAccount, balances []banktypes.Balance) {
	if len(cfg.balances) > 0 {
		balances = cfg.balances
		accounts := getAccAddrsFromBalances(balances)
		genAccounts = createGenesisAccounts(accounts)
	} else {
		genAccounts = createGenesisAccounts(cfg.preFundedAccounts)

		denomDecimals := cfg.chainCoins.DenomDecimalsMap()
		denoms := maps.Keys(denomDecimals)
		balances = createBalances(cfg.preFundedAccounts, append(cfg.otherCoinDenom, denoms...), denomDecimals)
	}

	// append validators to genesis accounts and balances
	valAccs := make([]sdktypes.AccAddress, len(validators))
	for i, v := range validators {
		valAddr, err := sdktypes.ValAddressFromBech32(v.OperatorAddress)
		if err != nil {
			panic(fmt.Sprintf("failed to derive validator address from %q: %s", v.OperatorAddress, err.Error()))
		}
		valAccs[i] = sdktypes.AccAddress(valAddr.Bytes())
	}
	genAccounts = append(genAccounts, createGenesisAccounts(valAccs)...)

	return
}

// ConfigOption defines a function that can modify the NetworkConfig.
// The purpose of this is to force to be declarative when the default configuration
// requires to be changed.
type ConfigOption func(*Config)

// WithChainID sets a custom chainID for the network. Changing the chainID
// change automatically also the EVM coin used. It panics if the chainID is invalid.
func WithChainID(chainID string) ConfigOption {
	eip155ChainID, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		panic(err)
	}

	components := strings.Split(chainID, "-")
	evmCoinInfo := evmtypes.ChainsCoinInfo[components[0]]

	return func(cfg *Config) {
		cfg.chainID = chainID
		cfg.eip155ChainID = eip155ChainID

		if cfg.chainCoins.IsBaseEqualToEVM() {
			cfg.chainCoins.baseCoin.Denom = evmCoinInfo.Denom
			cfg.chainCoins.baseCoin.Decimals = evmCoinInfo.Decimals
		}
		cfg.chainCoins.evmCoin.Denom = evmCoinInfo.Denom
		cfg.chainCoins.evmCoin.Decimals = evmCoinInfo.Decimals
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

// WithBaseCoin sets the denom and decimals for the base coin in the network.
func WithBaseCoin(denom string, decimals uint8) ConfigOption {
	return func(cfg *Config) {
		cfg.chainCoins.baseCoin.Denom = denom
		cfg.chainCoins.baseCoin.Decimals = evmtypes.Decimals(decimals)
	}
}

// WithEVMCoin sets the denom and decimals for the evm coin in the network.
func WithEVMCoin(_ string, _ uint8) ConfigOption {
	// The evm config can be changed only via chain ID because it should be
	// handled properly from the configurator.
	panic("EVM coin can be changed only via ChainID: se WithChainID method")
}

// WithCustomGenesis sets the custom genesis of the network for specific modules.
func WithCustomGenesis(customGenesis CustomGenesisState) ConfigOption {
	return func(cfg *Config) {
		cfg.customGenesisState = customGenesis
	}
}

// WithOtherDenoms sets other possible coin denominations for the network.
func WithOtherDenoms(otherDenoms []string) ConfigOption {
	return func(cfg *Config) {
		cfg.otherCoinDenom = otherDenoms
	}
}

// WithValidatorOperators overwrites the used operator address for the network instantiation.
func WithValidatorOperators(keys []sdktypes.AccAddress) ConfigOption {
	return func(cfg *Config) {
		cfg.operatorsAddrs = keys
	}
}

// WithCustomBaseAppOpts sets custom base app options for the network.
func WithCustomBaseAppOpts(opts ...func(*baseapp.BaseApp)) ConfigOption {
	return func(cfg *Config) {
		cfg.customBaseAppOpts = opts
	}
}
