// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"fmt"
	"math/big"
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v19/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

var (
	// DefaultEVMDenom defines the default EVM denomination on Evmos
	DefaultEVMDenom = utils.BaseDenom
	// DefaultAllowUnprotectedTxs rejects all unprotected txs (i.e false)
	DefaultAllowUnprotectedTxs = false
	// DefaultStaticPrecompiles defines the default active precompiles
	DefaultStaticPrecompiles = []string{
		P256PrecompileAddress,         // P256 precompile
		Bech32PrecompileAddress,       // Bech32 precompile
		StakingPrecompileAddress,      // Staking precompile
		DistributionPrecompileAddress, // Distribution precompile
		ICS20PrecompileAddress,        // ICS20 transfer precompile
		VestingPrecompileAddress,      // Vesting precompile
		BankPrecompileAddress,         // Bank precompile
	}
	// DefaultExtraEIPs defines the default extra EIPs to be included
	// On v15, EIP 3855 was enabled
	DefaultExtraEIPs   = []string{"ethereum_3855"}
	DefaultEVMChannels = []string{
		"channel-10", // Injective
		"channel-31", // Cronos
		"channel-83", // Kava
	}
	DefaultCreateAllowlistAddresses []string
	DefaultCallAllowlistAddresses   []string
	DefaultAccessControl            = AccessControl{
		Create: AccessControlType{
			AccessType:        AccessTypePermissionless,
			AccessControlList: DefaultCreateAllowlistAddresses,
		},
		Call: AccessControlType{
			AccessType:        AccessTypePermissionless,
			AccessControlList: DefaultCreateAllowlistAddresses,
		},
	}
)

// NewParams creates a new Params instance
func NewParams(
	evmDenom string,
	allowUnprotectedTxs bool,
	config ChainConfig,
	extraEIPs []string,
	activeStaticPrecompiles,
	evmChannels []string,
	accessControl AccessControl,
) Params {
	return Params{
		EvmDenom:                evmDenom,
		AllowUnprotectedTxs:     allowUnprotectedTxs,
		ExtraEIPs:               extraEIPs,
		ChainConfig:             config,
		ActiveStaticPrecompiles: activeStaticPrecompiles,
		EVMChannels:             evmChannels,
		AccessControl:           accessControl,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{
		EvmDenom:                DefaultEVMDenom,
		ChainConfig:             DefaultChainConfig(),
		ExtraEIPs:               DefaultExtraEIPs,
		AllowUnprotectedTxs:     DefaultAllowUnprotectedTxs,
		ActiveStaticPrecompiles: DefaultStaticPrecompiles,
		EVMChannels:             DefaultEVMChannels,
		AccessControl:           DefaultAccessControl,
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

	if err := validateBool(p.AllowUnprotectedTxs); err != nil {
		return err
	}

	if err := validateChainConfig(p.ChainConfig); err != nil {
		return err
	}

	if err := ValidatePrecompiles(p.ActiveStaticPrecompiles); err != nil {
		return err
	}

	if err := p.AccessControl.Validate(); err != nil {
		return err
	}

	return validateChannels(p.EVMChannels)
}

// EIPs returns the ExtraEIPS as a slice.
func (p Params) EIPs() []string {
	eips := make([]string, len(p.ExtraEIPs))
	copy(eips, p.ExtraEIPs)
	return eips
}

// GetActiveStaticPrecompilesAddrs is a util function that the Active Precompiles
// as a slice of addresses.
func (p Params) GetActiveStaticPrecompilesAddrs() []common.Address {
	precompiles := make([]common.Address, len(p.ActiveStaticPrecompiles))
	for i, precompile := range p.ActiveStaticPrecompiles {
		precompiles[i] = common.HexToAddress(precompile)
	}
	return precompiles
}

// IsEVMChannel returns true if the channel provided is in the list of
// EVM channels
func (p Params) IsEVMChannel(channel string) bool {
	return slices.Contains(p.EVMChannels, channel)
}

func (ac AccessControl) Validate() error {
	if err := ac.Create.Validate(); err != nil {
		return err
	}

	if err := ac.Call.Validate(); err != nil {
		return err
	}

	return nil
}

func (act AccessControlType) Validate() error {
	if err := validateAccessType(act.AccessType); err != nil {
		return err
	}

	if err := validateAllowlistAddresses(act.AccessControlList); err != nil {
		return err
	}
	return nil
}

func validateAccessType(i interface{}) error {
	accessType, ok := i.(AccessType)
	if !ok {
		return fmt.Errorf("invalid access type type: %T", i)
	}

	switch accessType {
	case AccessTypePermissionless, AccessTypeRestricted, AccessTypePermissioned:
		return nil
	default:
		return fmt.Errorf("invalid access type: %s", accessType)
	}
}

func validateAllowlistAddresses(i interface{}) error {
	addresses, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid whitelist addresses type: %T", i)
	}

	for _, address := range addresses {
		if err := types.ValidateAddress(address); err != nil {
			return fmt.Errorf("invalid whitelist address: %s", address)
		}
	}
	return nil
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
	eips, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	uniqueEIPs := make(map[string]struct{})

	for _, eip := range eips {
		if !vm.ExistsEipActivator(eip) {
			return fmt.Errorf("EIP %s is not activateable, valid EIPs are: %s", eip, vm.ActivateableEips())
		}

		if err := vm.ValidateEIPName(eip); err != nil {
			return fmt.Errorf("EIP %s name is not valid", eip)
		}

		if _, ok := uniqueEIPs[eip]; ok {
			return fmt.Errorf("found duplicate EIP: %s", eip)
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

	// NOTE: Check that the precompiles are sorted. This is required
	// to ensure determinism
	if !slices.IsSorted(precompiles) {
		return fmt.Errorf("precompiles need to be sorted: %s", precompiles)
	}

	return nil
}

// IsLondon returns if london hardfork is enabled.
func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	return ethConfig.IsLondon(big.NewInt(height))
}
