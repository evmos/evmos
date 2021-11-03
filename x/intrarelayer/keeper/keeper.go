package keeper

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	"github.com/tharsis/evmos/x/intrarelayer/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	vm "github.com/ethereum/go-ethereum/core/vm"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Keeper of this module maintains collections of intrarelayer.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	bankKeeper types.BankKeeper
	govKeeper  types.GovKeeper
	evmKeeper  *evmkeeper.Keeper // TODO: use interface
}

// NewKeeper creates new instances of the intrarelayer Keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	bk types.BankKeeper,
	govKeeper types.GovKeeper,
	evmKeeper *evmkeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramstore: ps,
		bankKeeper: bk,
		govKeeper:  govKeeper,
		evmKeeper:  evmKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ExecuteEVMusingModuleAddress(ctx sdk.Context, contractAddr, from common.Address, transferData []byte) error {
	params := k.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.evmKeeper.ChainID())
	// NOTE: pass in an empty coinbase address and nil tracer as we don't need them for the check below
	cfg := &evmtypes.EVMConfig{
		ChainConfig: ethCfg,
		Params:      params,
		CoinBase:    common.Address{},
		BaseFee:     big.NewInt(0),
	}
	tx := k.createModuleTx(contractAddr, from, transferData)
	msg, err := tx.AsMessage(ethtypes.MakeSigner(ethCfg, new(big.Int)), nil)
	if err != nil {
		return nil
	}

	vmConfig := k.evmKeeper.VMConfig(msg, cfg.Params, evmtypes.NewNoOpTracer())
	evm := k.evmKeeper.NewEVM(msg, cfg, evmtypes.NewNoOpTracer())
	interpreter := vm.NewEVMInterpreter(evm, vmConfig)

	// this calls should be enough for getting values

	addr := vm.AccountRef(contractAddr)

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	code := evm.StateDB.GetCode(from)
	if len(code) == 0 {
		// ret, err = nil, nil // gas is unchanged
		return nil
	}

	addrCopy := from
	contract := vm.NewContract(addr, vm.AccountRef(from), new(big.Int), 0)
	contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), code)
	ret, err := interpreter.Run(contract, transferData, false)
	if err != nil {
		return nil
	}
	fmt.Println(ret)
	return nil
}

func (k Keeper) createModuleTx(contractAddr, from common.Address, transferData []byte) *evmtypes.MsgEthereumTx {
	chainID := k.evmKeeper.ChainID()
	// args, err := json.Marshal(&evm.TransactionArgs{To: &contractAddr, From: &from, Data: (*hexutil.Bytes)(&transferData)})
	// if err != nil {
	// 	return nil
	// }
	// res, err := suite.queryClientEvm.EstimateGas(k.ctx, &evm.EthCallRequest{
	// 	Args:   args,
	// 	GasCap: uint64(config.DefaultGasCap),
	// })
	// if err != nil {
	// 	return nil
	// }

	nonce := k.evmKeeper.GetNonce(from)

	ercTransferTx := evmtypes.NewTx(
		chainID,
		nonce,
		&contractAddr,
		nil,
		uint64(0),
		nil,
		big.NewInt(0),
		big.NewInt(1),
		transferData,
		&ethtypes.AccessList{}, // accesses
	)

	ercTransferTx.From = from.Hex()
	return ercTransferTx
}
