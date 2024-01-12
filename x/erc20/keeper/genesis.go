// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"fmt"
	"math/big"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/ibc"
	"github.com/evmos/evmos/v16/x/erc20/types"
	"github.com/evmos/evmos/v16/x/evm/statedb"
)

// SetGenesisTokenPairs stores in state the provided token pairs in genesis
// A genesis token pair is a token pair that will be registered from genesis.
// A current use case of this is e2e tests,
// where we may need to register a token pair to make it work.
// Currently, there's no other way to do this because the RegisterCoinProposal was deprecated.
func (k Keeper) SetGenesisTokenPairs(ctx sdk.Context, pairs []types.TokenPair) error {
	if len(pairs) == 0 {
		return nil
	}
	// We need to store the ERC20 contracts in the provided addresses at genesis.
	// To do so, we'll need to seed the stateDB with those contracts.
	// We cannot use EVM keeper functions for this because that will create a contract address
	// dependent on the sender nonce.
	stateDB := statedb.New(ctx, k.evmKeeper, statedb.TxConfig{})
	for _, pair := range pairs {
		contractAddr := pair.GetERC20Contract()
		coinMeta, err := k.getTokenPairMeta(ctx, pair)
		if err != nil {
			return fmt.Errorf("error while generating metadata for genesis pair denom %s. %w", pair.Denom, err)
		}
		code, err := k.generateContractCode(ctx, stateDB, coinMeta, contractAddr)
		if err != nil {
			return fmt.Errorf("error while getting contract code for genesis pair denom %s. %w", pair.Denom, err)
		}

		// Store the ERC20 contract code in the address provided in genesis
		stateDB.SetCode(contractAddr, code)

		id := pair.GetID()
		k.SetTokenPair(ctx, pair)
		k.SetDenomMap(ctx, pair.Denom, id)
		k.SetERC20Map(ctx, contractAddr, id)
	}

	// Commit the changes to effectively seed the state DB with
	// the genesis token pairs ERC20 contracts
	return stateDB.Commit()
}

// getTokenPairMeta is a helper function to generate token pair metadata for the genesis token pairs
func (k Keeper) getTokenPairMeta(ctx sdk.Context, pair types.TokenPair) (banktypes.Metadata, error) {
	// The corresponding IBC denom trace should be included in genesis
	denomTrace, err := ibc.GetDenomTrace(*k.transferKeeper, ctx, pair.Denom)
	if err != nil {
		return banktypes.Metadata{}, err
	}

	// validate base denom length
	if len(denomTrace.BaseDenom) < 2 {
		return banktypes.Metadata{}, fmt.Errorf("denom trace base denom is too short. Should be at least 2 characters long, got %q", denomTrace.BaseDenom)
	}

	// check the denom prefix to define the corresponding exponent
	exponent, err := ibc.DeriveDecimalsFromDenom(denomTrace.BaseDenom)
	if err != nil {
		return banktypes.Metadata{}, err
	}

	meta := banktypes.Metadata{
		Description: fmt.Sprintf("%s IBC coin", denomTrace.BaseDenom),
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: pair.Denom, Exponent: 0, Aliases: []string{denomTrace.BaseDenom}},
			{Denom: denomTrace.BaseDenom[1:], Exponent: uint32(exponent)},
		},
		Base:    pair.Denom,
		Display: denomTrace.BaseDenom[1:],
		Name:    strings.ToUpper(string(denomTrace.BaseDenom[1])) + denomTrace.BaseDenom[2:],
		Symbol:  strings.ToUpper(denomTrace.BaseDenom[1:]),
	}

	if err := k.verifyMetadata(ctx, meta); err != nil {
		return banktypes.Metadata{}, errorsmod.Wrapf(
			types.ErrInternalTokenPair, "coin metadata is invalid for genesis pair denom %s", pair.Denom,
		)
	}
	return meta, nil
}

// generateContractCode is a helper function to generate
// the ERC20 contract code to be stored at genesis when token pairs are provided
func (k Keeper) generateContractCode(ctx sdk.Context, stateDB *statedb.StateDB, coinMeta banktypes.Metadata, contractAddr common.Address) ([]byte, error) {
	evm := k.newEVM(ctx, stateDB)
	data, err := getContractDataBz(coinMeta)
	if err != nil {
		return nil, err
	}
	// Initialize a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	sender := vm.AccountRef(types.ModuleAddress)
	contract := vm.NewContract(sender, vm.AccountRef(contractAddr), big.NewInt(0), 1000000)
	contract.Code = data

	// Run the contract's constructor function to get the contract code to
	// be stored on chain
	return evm.Interpreter().Run(contract, nil, false)
}

// getEVMConfig is a helper function to get an EVM config
// needed to instantiate the EVM when token pairs are provided
// at genesis
func (k Keeper) getEVMConfig(ctx sdk.Context) *statedb.EVMConfig {
	params := k.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.evmKeeper.ChainID())

	baseFee := k.evmKeeper.GetBaseFee(ctx, ethCfg)
	return &statedb.EVMConfig{
		Params:      params,
		ChainConfig: ethCfg,
		CoinBase:    common.Address{},
		BaseFee:     baseFee,
	}
}

// newEVM is a helper function used during genesis
// to instantiate the EVM to set genesis state
func (k Keeper) newEVM(ctx sdk.Context, db *statedb.StateDB) *vm.EVM {
	cfg := k.getEVMConfig(ctx)

	// dummy message needed to instantiate EVM
	msg := ethtypes.NewMessage(
		types.ModuleAddress,
		nil,
		0,
		big.NewInt(0), // amount
		0,             // gasLimit
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		big.NewInt(0), // gasPrice
		nil,
		ethtypes.AccessList{}, // AccessList
		false,                 // isFake
	)

	return k.evmKeeper.NewEVM(ctx, msg, cfg, nil, db)
}
