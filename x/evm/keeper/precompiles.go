// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/evmos/evmos/v16/precompiles/bech32"
	"github.com/evmos/evmos/v16/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"golang.org/x/exp/maps"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	bankprecompile "github.com/evmos/evmos/v16/precompiles/bank"
	distprecompile "github.com/evmos/evmos/v16/precompiles/distribution"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	ics20precompile "github.com/evmos/evmos/v16/precompiles/ics20"
	osmosisoutpost "github.com/evmos/evmos/v16/precompiles/outposts/osmosis"
	strideoutpost "github.com/evmos/evmos/v16/precompiles/outposts/stride"
	"github.com/evmos/evmos/v16/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v16/precompiles/staking"
	vestingprecompile "github.com/evmos/evmos/v16/precompiles/vesting"
	erc20Keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	"github.com/evmos/evmos/v16/x/evm/types"
	transferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
	vestingkeeper "github.com/evmos/evmos/v16/x/vesting/keeper"
)

// AvailablePrecompiles returns the list of all available precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func AvailablePrecompiles(
	chainID string,
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
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
		panic(fmt.Errorf("failed to instantiate bech32 precompile: %w", err))
	}

	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate staking precompile: %w", err))
	}

	distributionPrecompile, err := distprecompile.NewPrecompile(distributionKeeper, stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate distribution precompile: %w", err))
	}

	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(transferKeeper, channelKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ICS20 precompile: %w", err))
	}

	vestingPrecompile, err := vestingprecompile.NewPrecompile(vestingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate vesting precompile: %w", err))
	}

	bankPrecompile, err := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	var WEVMOSAddress common.Address
	if utils.IsMainnet(chainID) {
		WEVMOSAddress = common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
	} else {
		WEVMOSAddress = common.HexToAddress(erc20precompile.WEVMOSContractTestnet)
	}

	strideOutpost, err := strideoutpost.NewPrecompile(
		WEVMOSAddress,
		transferKeeper,
		erc20Keeper,
		authzKeeper,
		stakingKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate stride outpost: %w", err))
	}

	osmosisOutpost, err := osmosisoutpost.NewPrecompile(
		WEVMOSAddress,
		authzKeeper,
		bankKeeper,
		transferKeeper,
		stakingKeeper,
		erc20Keeper,
		channelKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate osmosis outpost: %w", err))
	}

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[vestingPrecompile.Address()] = vestingPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[bankPrecompile.Address()] = bankPrecompile

	// Outposts
	precompiles[strideOutpost.Address()] = strideOutpost
	precompiles[osmosisOutpost.Address()] = osmosisOutpost

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

// GetCachedPrecompiles returns the subset of the available precompiled contracts that
// are initalized in memory. If a precompile is not found in the cache, it will be
// instantiated and added to the cache.
func (k Keeper) GetCachedPrecompiles(
	ctx sdk.Context,
	activePrecompiles ...common.Address,
) map[common.Address]vm.PrecompiledContract {
	newActivePrecompileMap := make(map[common.Address]vm.PrecompiledContract)
	for _, address := range activePrecompiles {
		// Check cached precompiles
		cachedPrecompile, ok := k.precompiles[address]

		// If precompile is not found on cache, try to initiate it
		// It can only be an erc20 precompile
		if !ok {
			precompile, err := k.erc20Keeper.InstantiateERC20Precompile(ctx, address)
			if err != nil {
				panic(fmt.Sprintf("precompiled contract not initialized: %s", address))
			}
			fmt.Println(precompile.Address())
			newActivePrecompileMap[address] = precompile
			continue
		}
		newActivePrecompileMap[address] = cachedPrecompile
	}

	// Update cache
	k.precompiles = newActivePrecompileMap
	return k.precompiles
}

// AddEVMExtensions adds the given precompiles to the list of active precompiles in the EVM parameters
// and to the available precompiles map in the Keeper. This function returns an error if
// the precompiles are invalid or duplicated.
func (k *Keeper) AddEVMExtensions(ctx sdk.Context, precompiles ...vm.PrecompiledContract) error {
	addresses := make([]common.Address, len(precompiles))

	precompilesMap := make(map[common.Address]vm.PrecompiledContract, len(precompiles))
	precompilesMap = maps.Clone(k.precompiles)

	// Iterate over precompiles and:
	// - validate it doesn't already exist in the active precompiles
	// - get the string address and add it to the active precompiles
	params := k.GetParams(ctx)
	for i, precompile := range precompiles {
		// add to active precompiles
		address := precompile.Address()
		if ok := params.IsActivePrecompile(address.String()); ok {
			return errorsmod.Wrapf(types.ErrDuplicatePrecompile, "precompile already registered: %s", address)
		}

		addresses[i] = address
		precompilesMap[address] = precompile
	}

	err := k.EnablePrecompiles(ctx, addresses...)
	if err != nil {
		return fmt.Errorf("failed to enable precompiles: %w", err)
	}

	// set the new precompiles map in memory for faster
	// access during the block processing
	if !ctx.IsCheckTx() {
		k.precompiles = precompilesMap
	}
	return nil
}

// IsAvailablePrecompile returns true if the given precompile address is contained in the
// EVM keeper's available precompiles map.
func (k Keeper) IsAvailablePrecompile(ctx sdk.Context, address common.Address) bool {
	params := k.GetParams(ctx)
	return params.IsActivePrecompile(address.String())
}
