// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	bankprecompile "github.com/evmos/evmos/v16/precompiles/bank"
	"github.com/evmos/evmos/v16/precompiles/bech32"
	distprecompile "github.com/evmos/evmos/v16/precompiles/distribution"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	ics20precompile "github.com/evmos/evmos/v16/precompiles/ics20"
	osmosisoutpost "github.com/evmos/evmos/v16/precompiles/outposts/osmosis"
	strideoutpost "github.com/evmos/evmos/v16/precompiles/outposts/stride"
	"github.com/evmos/evmos/v16/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v16/precompiles/staking"
	vestingprecompile "github.com/evmos/evmos/v16/precompiles/vesting"
	"github.com/evmos/evmos/v16/utils"
	erc20Keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	"github.com/evmos/evmos/v16/x/evm/types"
	transferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
	vestingkeeper "github.com/evmos/evmos/v16/x/vesting/keeper"
	"golang.org/x/exp/maps"
)

// AvailableStaticPrecompiles returns the list of all available precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func AvailableStaticPrecompiles(
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

// WithStaticPrecompiles sets the available precompiled contracts.
func (k *Keeper) WithStaticPrecompiles(precompiles map[common.Address]vm.PrecompiledContract) *Keeper {
	if k.precompiles != nil {
		panic("available precompiles map already set")
	}

	if len(precompiles) == 0 {
		panic("empty precompiled contract map")
	}

	k.precompiles = precompiles
	return k
}

// GetStaticPrecompilesInstances returns the addresses and instances of the active static precompiles.
// Includes the Berlin precompiles.
func (k Keeper) GetStaticPrecompilesInstances(
	params *types.Params,
) ([]common.Address, map[common.Address]vm.PrecompiledContract) {
	activePrecompileMap := make(map[common.Address]vm.PrecompiledContract)
	activeLen := len(params.ActiveStaticPrecompiles)
	totalLen := activeLen + len(vm.PrecompiledAddressesBerlin)
	addresses := make([]common.Address, totalLen)
	// Append the Berlin precompiles to the active precompiles addresses
	// Add params precompiles
	for i, address := range params.ActiveStaticPrecompiles {
		hexAddress := common.HexToAddress(address)

		precompile, ok := k.precompiles[hexAddress]
		if !ok {
			panic(fmt.Sprintf("precompiled contract not initialized: %s", address))
		}

		activePrecompileMap[hexAddress] = precompile
		addresses[i] = hexAddress
	}

	// Add Berlin precompiles to map
	for _, precompile := range vm.PrecompiledAddressesBerlin {
		precompile, ok := k.precompiles[precompile]
		if !ok {
			panic(fmt.Sprintf("precompiled contract not initialized: %s", precompile))
		}
		activePrecompileMap[precompile.Address()] = precompile
	}
	// Copy Berlin precompiles to addresses
	copy(addresses[activeLen:], vm.PrecompiledAddressesBerlin)

	return addresses, activePrecompileMap
}

// IsAvailablePrecompile returns true if the given precompile address is contained in the
// EVM keeper's available precompiles map.
func (k Keeper) IsAvailablePrecompile(address common.Address) bool {
	_, ok := k.precompiles[address]
	return ok
}
