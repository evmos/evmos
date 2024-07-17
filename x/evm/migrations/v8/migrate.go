// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v8

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v7types "github.com/evmos/evmos/v18/x/evm/migrations/v8/types"
	"github.com/evmos/evmos/v18/x/evm/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 7 to
// version 8. Specifically, it changes the type of the Params ExtraEIPs from
// []int64 to []string.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		paramsV7 v7types.V7Params
		params   types.Params
	)

	store := ctx.KVStore(storeKey)

	paramsV7Bz := store.Get(types.KeyPrefixParams)
	cdc.MustUnmarshal(paramsV7Bz, &paramsV7)

	params.EvmDenom = paramsV7.EvmDenom
	params.ChainConfig = types.ChainConfig{
		HomesteadBlock:      paramsV7.ChainConfig.HomesteadBlock,
		DAOForkBlock:        paramsV7.ChainConfig.DAOForkBlock,
		DAOForkSupport:      paramsV7.ChainConfig.DAOForkSupport,
		EIP150Block:         paramsV7.ChainConfig.EIP150Block,
		EIP150Hash:          paramsV7.ChainConfig.EIP150Hash,
		EIP155Block:         paramsV7.ChainConfig.EIP155Block,
		EIP158Block:         paramsV7.ChainConfig.EIP158Block,
		ByzantiumBlock:      paramsV7.ChainConfig.ByzantiumBlock,
		ConstantinopleBlock: paramsV7.ChainConfig.ConstantinopleBlock,
		PetersburgBlock:     paramsV7.ChainConfig.PetersburgBlock,
		IstanbulBlock:       paramsV7.ChainConfig.IstanbulBlock,
		MuirGlacierBlock:    paramsV7.ChainConfig.MuirGlacierBlock,
		BerlinBlock:         paramsV7.ChainConfig.BerlinBlock,
		LondonBlock:         paramsV7.ChainConfig.LondonBlock,
		ArrowGlacierBlock:   paramsV7.ChainConfig.ArrowGlacierBlock,
		GrayGlacierBlock:    paramsV7.ChainConfig.GrayGlacierBlock,
		MergeNetsplitBlock:  paramsV7.ChainConfig.MergeNetsplitBlock,
		ShanghaiBlock:       paramsV7.ChainConfig.ShanghaiBlock,
		CancunBlock:         paramsV7.ChainConfig.CancunBlock,
	}
	params.AllowUnprotectedTxs = paramsV7.AllowUnprotectedTxs
	params.ActivePrecompiles = paramsV7.ActivePrecompiles
	params.EVMChannels = paramsV7.EVMChannels

	create := types.AccessControlType{
		AccessType:        types.AccessType(paramsV7.AccessControl.Call.AccessType),
		AccessControlList: paramsV7.AccessControl.Call.AccessControlList,
	}

	call := types.AccessControlType{
		AccessType:        types.AccessType(paramsV7.AccessControl.Create.AccessType),
		AccessControlList: paramsV7.AccessControl.Create.AccessControlList,
	}

	params.AccessControl = types.AccessControl{
		Create: create,
		Call:   call,
	}

	// Migrate old ExtraEIPs from int64 to string. Since no Evmos EIPs have been
	// created before and activators contains only `ethereum_XXXX` activations,
	// all values will be prefixed with `ethereum_`.
	params.ExtraEIPs = make([]string, 0, len(paramsV7.ExtraEIPs))
	for _, eip := range paramsV7.ExtraEIPs {
		eipName := fmt.Sprintf("ethereum_%d", eip)
		params.ExtraEIPs = append(params.ExtraEIPs, eipName)
	}

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)

	store.Set(types.KeyPrefixParams, bz)

	return nil
}
