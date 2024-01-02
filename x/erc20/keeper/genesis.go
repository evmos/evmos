package keeper

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/x/erc20/types"
	"github.com/evmos/evmos/v16/x/evm/statedb"
)

// SetGenesisTokenPairs stores in state the provided token pairs in genesis
func (k Keeper) SetGenesisTokenPairs(ctx sdk.Context, pairs []types.TokenPair) error {
	for i, pair := range pairs {
		contractAddr := pair.GetERC20Contract()
		displayName := fmt.Sprintf("genToken%d", i)
		coinMeta := banktypes.Metadata{
			Description: fmt.Sprintf("Genesis token pair number %d", i),
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: pair.Denom, Exponent: 0, Aliases: []string{fmt.Sprintf("uGenToken%d", i)}},
				{Denom: displayName, Exponent: 6},
			},
			Base:    pair.Denom,
			Display: displayName,
			Name:    displayName,
			Symbol:  displayName,
		}
		if err := k.verifyMetadata(ctx, coinMeta); err != nil {
			return errorsmod.Wrapf(
				types.ErrInternalTokenPair, "coin metadata is invalid for genesis pair denom %s", pair.Denom,
			)
		}

		stateDB := statedb.New(ctx, k.evmKeeper, statedb.TxConfig{})
		cfg := k.getEVMConfig(ctx)

		// dummy message needed to instanciate EVM
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

		evm := k.evmKeeper.NewEVM(ctx, msg, cfg, nil, stateDB)
		code, err := generateContractCode(evm, coinMeta, contractAddr)
		if err != nil {
			return fmt.Errorf("error while getting contract code for genesis pair denom %s. %w", pair.Denom, err)
		}
		stateDB.SetCode(contractAddr, code)

		if err := stateDB.Commit(); err != nil {
			return fmt.Errorf("error while storing contract for genesis pair denom %s. %w", pair.Denom, err)
		}

		id := pair.GetID()
		k.SetTokenPair(ctx, pair)
		k.SetDenomMap(ctx, pair.Denom, id)
		k.SetERC20Map(ctx, pair.GetERC20Contract(), id)
	}

	return nil
}

// getEVMConfig is a helper function to get an EVM config
// needed to instanciate the EVM when token pairs are provided
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

// generateContractCode is a helper function to generate
// the ERC20 contract code to be stored at genesis when token pairs are provided
func generateContractCode(evm *vm.EVM, coinMeta banktypes.Metadata, contractAddr common.Address) ([]byte, error) {
	data, err := getContractDataBz(coinMeta)
	if err != nil {
		return nil, err
	}
	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	sender := vm.AccountRef(types.ModuleAddress)
	contract := vm.NewContract(sender, vm.AccountRef(contractAddr), big.NewInt(0), 1000000)
	contract.Code = data

	return evm.Interpreter().Run(contract, nil, false)
}
