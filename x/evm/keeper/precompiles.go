// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/evmos/evmos/v15/precompiles/bech32"

	"golang.org/x/exp/maps"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	distprecompile "github.com/evmos/evmos/v15/precompiles/distribution"
	ics20precompile "github.com/evmos/evmos/v15/precompiles/ics20"
	strideoutpost "github.com/evmos/evmos/v15/precompiles/outposts/stride"
	"github.com/evmos/evmos/v15/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v15/precompiles/staking"
	vestingprecompile "github.com/evmos/evmos/v15/precompiles/vesting"
	erc20Keeper "github.com/evmos/evmos/v15/x/erc20/keeper"
	transferkeeper "github.com/evmos/evmos/v15/x/ibc/transfer/keeper"
	vestingkeeper "github.com/evmos/evmos/v15/x/vesting/keeper"
)

// AvailablePrecompiles returns the list of all available precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func AvailablePrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	vestingKeeper vestingkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
) map[common.Address]vm.PrecompiledContract {
	// Clone the mapping from the latest EVM fork.
	precompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// secp256r1 precompile as per EIP-7212
	p256Precompile := &p256.Precompile{}

	bech32Precompile, err := bech32.NewPrecompile(6000)
	if err != nil {
		panic(fmt.Errorf("failed to load bech32 precompile: %w", err))
	}

	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load staking precompile: %w", err))
	}

	distributionPrecompile, err := distprecompile.NewPrecompile(distributionKeeper, stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load distribution precompile: %w", err))
	}

	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(transferKeeper, channelKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load ICS20 precompile: %w", err))
	}

	vestingPrecompile, err := vestingprecompile.NewPrecompile(vestingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load vesting precompile: %w", err))
	}

	strideOutpost, err := strideoutpost.NewPrecompile(transfertypes.PortID, "channel-25", transferKeeper, erc20Keeper, authzKeeper, stakingKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to load stride outpost: %w", err))
	}

	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[vestingPrecompile.Address()] = vestingPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[strideOutpost.Address()] = strideOutpost
	return precompiles
}

// WithPrecompiles sets the available precompiled contracts.
func (k *Keeper) WithPrecompiles(precompiles map[common.Address]vm.PrecompiledContract) *Keeper {
	if k.precompiles != nil {
		panic("available precompiles map already set")
	}

	if len(precompiles) == 0 {
		panic("empty precompiled contract map")
	}

	k.precompiles = precompiles
	return k
}

// Precompiles returns the subset of the available precompiled contracts that
// are active given the current parameters.
func (k Keeper) Precompiles(
	activePrecompiles ...common.Address,
) map[common.Address]vm.PrecompiledContract {
	activePrecompileMap := make(map[common.Address]vm.PrecompiledContract)

	for _, address := range activePrecompiles {
		precompile, ok := k.precompiles[address]
		if !ok {
			panic(fmt.Sprintf("precompiled contract not initialized: %s", address))
		}

		activePrecompileMap[address] = precompile
	}

	return activePrecompileMap
}

// AddEVMExtensions adds the given precompiles to the list of active precompiles in the EVM parameters
// and to the available precompiles map in the Keeper. This function returns an error if
// the precompiles are invalid or duplicated.
func (k *Keeper) AddEVMExtensions(ctx sdk.Context, precompiles ...vm.PrecompiledContract) error {
	params := k.GetParams(ctx)

	addresses := make([]string, len(precompiles))
	precompilesMap := maps.Clone(k.precompiles)

	for i, precompile := range precompiles {
		// add to active precompiles
		address := precompile.Address()
		addresses[i] = address.String()

		// add to available precompiles, but check for duplicates
		if _, ok := precompilesMap[address]; ok {
			return fmt.Errorf("precompile already registered: %s", address)
		}
		precompilesMap[address] = precompile
	}

	params.ActivePrecompiles = append(params.ActivePrecompiles, addresses...)

	// NOTE: the active precompiles are sorted and validated before setting them
	// in the params
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}

	// update the pointer to the map with the newly added EVM Extensions
	k.precompiles = precompilesMap
	return nil
}

// IsAvailablePrecompile returns true if the given precompile address is contained in the
// EVM keeper's available precompiles map.
func (k Keeper) IsAvailablePrecompile(address common.Address) bool {
	_, ok := k.precompiles[address]
	return ok
}

// GetAvailablePrecompileAddrs returns the list of available precompile addresses.
//
// NOTE: uses index based approach instead of append because it's supposed to be faster.
// Check https://stackoverflow.com/questions/21362950/getting-a-slice-of-keys-from-a-map.
func (k Keeper) GetAvailablePrecompileAddrs() []common.Address {
	addresses := make([]common.Address, len(k.precompiles))
	i := 0

	//#nosec G705 -- two operations in for loop here are fine
	for address := range k.precompiles {
		addresses[i] = address
		i++
	}

	sort.Slice(addresses, func(i, j int) bool {
		return bytes.Compare(addresses[i].Bytes(), addresses[j].Bytes()) == -1
	})

	return addresses
}
