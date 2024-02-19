package v17

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v16/utils"
	"github.com/evmos/evmos/v16/x/evm/types"

	v16types "github.com/evmos/evmos/v16/x/evm/migrations/v17/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 5 to
// version 6. Specifically, it adds the new EVMChannels param.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		paramsV16 v16types.V16Params
		params    types.Params
	)

	store := ctx.KVStore(storeKey)

	paramsV16Bz := store.Get(types.KeyPrefixParams)
	cdc.MustUnmarshal(paramsV16Bz, &paramsV16)

	params.EvmDenom = paramsV16.EvmDenom
	params.EnableCreate = paramsV16.EnableCreate
	params.EnableCall = paramsV16.EnableCall
	params.ExtraEIPs = paramsV16.ExtraEIPs
	params.ChainConfig = types.ChainConfig{
		HomesteadBlock:      paramsV16.ChainConfig.HomesteadBlock,
		DAOForkBlock:        paramsV16.ChainConfig.DAOForkBlock,
		DAOForkSupport:      paramsV16.ChainConfig.DAOForkSupport,
		EIP150Block:         paramsV16.ChainConfig.EIP150Block,
		EIP150Hash:          paramsV16.ChainConfig.EIP150Hash,
		EIP155Block:         paramsV16.ChainConfig.EIP155Block,
		EIP158Block:         paramsV16.ChainConfig.EIP158Block,
		ByzantiumBlock:      paramsV16.ChainConfig.ByzantiumBlock,
		ConstantinopleBlock: paramsV16.ChainConfig.ConstantinopleBlock,
		PetersburgBlock:     paramsV16.ChainConfig.PetersburgBlock,
		IstanbulBlock:       paramsV16.ChainConfig.IstanbulBlock,
		MuirGlacierBlock:    paramsV16.ChainConfig.MuirGlacierBlock,
		BerlinBlock:         paramsV16.ChainConfig.BerlinBlock,
		LondonBlock:         paramsV16.ChainConfig.LondonBlock,
		ArrowGlacierBlock:   paramsV16.ChainConfig.ArrowGlacierBlock,
		GrayGlacierBlock:    paramsV16.ChainConfig.GrayGlacierBlock,
		MergeNetsplitBlock:  paramsV16.ChainConfig.MergeNetsplitBlock,
		ShanghaiBlock:       paramsV16.ChainConfig.ShanghaiBlock,
		CancunBlock:         paramsV16.ChainConfig.CancunBlock,
	}
	params.AllowUnprotectedTxs = paramsV16.AllowUnprotectedTxs
	params.ActiveStaticPrecompiles = paramsV16.ActivePrecompiles
	params.EVMChannels = types.DefaultEVMChannels

	// DefaultEVMChannels are for Evmos mainnet
	// leave empty for testnet
	if ctx.ChainID() == utils.TestnetChainID+"-4" {
		params.EVMChannels = []string{}
	}

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)

	store.Set(types.KeyPrefixParams, bz)
	return nil
}
