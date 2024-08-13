// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v20

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v19/contracts"
	evmostypes "github.com/evmos/evmos/v19/types"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmoscore "github.com/evmos/evmos/v19/x/evm/core/core"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	"github.com/evmos/evmos/v19/x/evm/statedb"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	erc20k erc20keeper.Keeper,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running migrations...")
		if err := AddCodeToERC20Extensions(ctx, erc20k, ek, bk); err == nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// AddCodeToERC20Extensions adds code and code hash to the ERC20 precompiles with the EVM.
func AddCodeToERC20Extensions(ctx sdk.Context,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) error {
	// need the interpreter to get the code after running the constructor
	interpreter, err := getInterpreter(ctx, evmKeeper)
	if err != nil {
		return err
	}

	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
		meta, found := bankKeeper.GetDenomMetaData(ctx, tokenPair.Denom)
		if !found {
			err = fmt.Errorf("no bank metadata found for %s", tokenPair.Denom)
			return true
		}

		decimals := uint8(0)
		if len(meta.DenomUnits) > 0 {
			decimalsIdx := len(meta.DenomUnits) - 1
			decimals = uint8(meta.DenomUnits[decimalsIdx].Exponent)
		}

		var ctorArgs []byte
		ctorArgs, err = contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack(
			"",
			meta.Name,
			meta.Symbol,
			decimals,
		)
		if err != nil {
			return true
		}

		data := make([]byte, len(contracts.ERC20MinterBurnerDecimalsContract.Bin)+len(ctorArgs))
		copy(data[:len(contracts.ERC20MinterBurnerDecimalsContract.Bin)], contracts.ERC20MinterBurnerDecimalsContract.Bin)
		copy(data[len(contracts.ERC20MinterBurnerDecimalsContract.Bin):], ctorArgs)

		contractAddr := common.HexToAddress(tokenPair.Erc20Address)
		contract := vm.NewContract(vm.AccountRef(erc20types.ModuleAddress), vm.AccountRef(contractAddr), common.Big0, 200000)
		contract.Code = data
		contract.CodeHash = crypto.Keccak256Hash(data)

		ret, err := interpreter.Run(contract, nil, false)
		if err != nil {
			return true
		}

		codeHash := crypto.Keccak256(ret)
		evmKeeper.SetCode(ctx, codeHash, ret)
		err = evmKeeper.SetAccount(ctx, contractAddr, statedb.Account{
			CodeHash: codeHash,
			Nonce:    0,
			Balance:  common.Big0,
		})

		return err != nil
	})

	return err
}

func getInterpreter(
	ctx sdk.Context,
	evmKeeper *evmkeeper.Keeper,
) (vm.Interpreter, error) {
	txConfig := evmKeeper.TxConfig(ctx, common.Hash{})
	stateDB := statedb.New(ctx, evmKeeper, txConfig)

	eip155ChainID, err := evmostypes.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	cfg, err := evmKeeper.EVMConfig(ctx, sdk.ConsAddress(ctx.BlockHeader().ProposerAddress), eip155ChainID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to load evm config")
	}

	blockCtx := vm.BlockContext{
		CanTransfer: evmoscore.CanTransfer,
		Transfer:    evmoscore.Transfer,
		GetHash:     evmKeeper.GetHashFn(ctx),
		Coinbase:    cfg.CoinBase,
		GasLimit:    evmostypes.BlockGasLimit(ctx),
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(ctx.BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0),
		BaseFee:     cfg.BaseFee,
		Random:      nil,
	}

	evm := vm.NewEVM(blockCtx, vm.TxContext{}, stateDB, cfg.ChainConfig, vm.Config{})
	// need the interpreter to get the code after running the constructor
	return evm.Interpreter(), nil
}
