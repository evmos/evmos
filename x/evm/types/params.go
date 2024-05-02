// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v18/precompiles/p256"
	"github.com/evmos/evmos/v18/types"
	"github.com/evmos/evmos/v18/utils"
)

var (
	// DefaultEVMDenom defines the default EVM denomination on Evmos
	DefaultEVMDenom = utils.BaseDenom
	// DefaultAllowUnprotectedTxs rejects all unprotected txs (i.e false)
	DefaultAllowUnprotectedTxs = false
	// DefaultEnableCreate enables contract creation (i.e true)
	DefaultEnableCreate = true
	// DefaultEnableCall enables contract calls (i.e true)
	DefaultEnableCall = true
	// AvailableEVMExtensions defines the default active precompiles
	AvailableEVMExtensions = []string{
		p256.PrecompileAddress,                       // P256 precompile
		"0x0000000000000000000000000000000000000400", // Bech32 precompile
		"0x0000000000000000000000000000000000000800", // Staking precompile
		"0x0000000000000000000000000000000000000801", // Distribution precompile
		"0x0000000000000000000000000000000000000802", // ICS20 transfer precompile
		"0x0000000000000000000000000000000000000803", // Vesting precompile
		"0x0000000000000000000000000000000000000804", // Bank precompile
	}
	// DefaultExtraEIPs defines the default extra EIPs to be included
	// On v15, EIP 3855 was enabled
	DefaultExtraEIPs   = []int64{3855}
	DefaultEVMChannels = []string{
		"channel-10", // Injective
		"channel-31", // Cronos
		"channel-83", // Kava
	}
)

// NewParams creates a new Params instance
func NewParams(
	evmDenom string,
	allowUnprotectedTxs,
	enableCreate,
	enableCall bool,
	config ChainConfig,
	extraEIPs []int64,
	activePrecompiles,
	evmChannels []string,
) Params {
	return Params{
		EvmDenom:            evmDenom,
		AllowUnprotectedTxs: allowUnprotectedTxs,
		EnableCreate:        enableCreate,
		EnableCall:          enableCall,
		ExtraEIPs:           extraEIPs,
		ChainConfig:         config,
		ActivePrecompiles:   activePrecompiles,
		EVMChannels:         evmChannels,
	}
}

// DefaultParams returns default evm parameters
// ExtraEIPs is empty to prevent overriding the latest hard fork instruction set
// ActivePrecompiles is empty to prevent overriding the default precompiles
// from the EVM configuration.
func DefaultParams() Params {
	return Params{
		EvmDenom:            DefaultEVMDenom,
		EnableCreate:        DefaultEnableCreate,
		EnableCall:          DefaultEnableCall,
		ChainConfig:         DefaultChainConfig(),
		ExtraEIPs:           DefaultExtraEIPs,
		AllowUnprotectedTxs: DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   AvailableEVMExtensions,
		EVMChannels:         DefaultEVMChannels,
	}
}

// validateChannels checks if channels ids are valid
func validateChannels(i interface{}) error {
	channels, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, channel := range channels {
		if err := host.ChannelIdentifierValidator(channel); err != nil {
			return errorsmod.Wrap(
				channeltypes.ErrInvalidChannelIdentifier, err.Error(),
			)
		}
	}

	return nil
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	if err := validateEVMDenom(p.EvmDenom); err != nil {
		return err
	}

	if err := validateEIPs(p.ExtraEIPs); err != nil {
		return err
	}

	if err := validateBool(p.EnableCall); err != nil {
		return err
	}

	if err := validateBool(p.EnableCreate); err != nil {
		return err
	}

	if err := validateBool(p.AllowUnprotectedTxs); err != nil {
		return err
	}

	if err := validateChainConfig(p.ChainConfig); err != nil {
		return err
	}

	if err := ValidatePrecompiles(p.ActivePrecompiles); err != nil {
		return err
	}

	return validateChannels(p.EVMChannels)
}

// EIPs returns the ExtraEIPS as a int slice
func (p Params) EIPs() []int {
	eips := make([]int, len(p.ExtraEIPs))
	for i, eip := range p.ExtraEIPs {
		eips[i] = int(eip)
	}
	return eips
}

// HasCustomPrecompiles returns true if the ActivePrecompiles slice is not empty.
func (p Params) HasCustomPrecompiles() bool {
	return len(p.ActivePrecompiles) > 0
}

// GetActivePrecompilesAddrs is a util function that the Active Precompiles
// as a slice of addresses.
func (p Params) GetActivePrecompilesAddrs() []common.Address {
	precompiles := make([]common.Address, len(p.ActivePrecompiles))
	for i, precompile := range p.ActivePrecompiles {
		precompiles[i] = common.HexToAddress(precompile)
	}
	return precompiles
}

// IsEVMChannel returns true if the channel provided is in the list of
// EVM channels
func (p Params) IsEVMChannel(channel string) bool {
	return slices.Contains(p.EVMChannels, channel)
}

// IsActivePrecompile returns true if the given precompile address is
// registered as an active precompile.
func (p Params) IsActivePrecompile(address string) bool {
	_, found := sort.Find(len(p.ActivePrecompiles), func(i int) int {
		return strings.Compare(address, p.ActivePrecompiles[i])
	})

	return found
}

func validateEVMDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter EVM denom type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEIPs(i interface{}) error {
	eips, ok := i.([]int64)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	uniqueEIPs := make(map[int64]struct{})

	for _, eip := range eips {
		if !vm.ValidEip(int(eip)) {
			return fmt.Errorf("EIP %d is not activateable, valid EIPs are: %s", eip, vm.ActivateableEips())
		}

		if _, ok := uniqueEIPs[eip]; ok {
			return fmt.Errorf("found duplicate EIP: %d", eip)
		}
		uniqueEIPs[eip] = struct{}{}
	}

	return nil
}

func validateChainConfig(i interface{}) error {
	cfg, ok := i.(ChainConfig)
	if !ok {
		return fmt.Errorf("invalid chain config type: %T", i)
	}

	return cfg.Validate()
}

// ValidatePrecompiles checks if the precompile addresses are valid and unique.
func ValidatePrecompiles(i interface{}) error {
	precompiles, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid precompile slice type: %T", i)
	}

	seenPrecompiles := make(map[string]struct{})
	for _, precompile := range precompiles {
		if _, ok := seenPrecompiles[precompile]; ok {
			return fmt.Errorf("duplicate precompile %s", precompile)
		}

		if err := types.ValidateAddress(precompile); err != nil {
			return fmt.Errorf("invalid precompile %s", precompile)
		}

		seenPrecompiles[precompile] = struct{}{}
	}

	// NOTE: Check that the precompiles are sorted. This is required for the
	// precompiles to be found correctly when using the IsActivePrecompile method,
	// because of the use of sort.Find.
	if !slices.IsSorted(precompiles) {
		return fmt.Errorf("precompiles need to be sorted: %s", precompiles)
	}

	return nil
}

// IsLondon returns if london hardfork is enabled.
func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	return ethConfig.IsLondon(big.NewInt(height))
}
