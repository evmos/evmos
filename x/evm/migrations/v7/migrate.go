// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v7

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v6types "github.com/evmos/evmos/v19/x/evm/migrations/v7/types"
	"github.com/evmos/evmos/v19/x/evm/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 6 to
// version 7. Specifically, it changes the type of the Params ExtraEIPs from
// []int64 to []string and introduces the access control.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		paramsV6 v6types.V6Params
		params   types.Params
	)

	store := ctx.KVStore(storeKey)

	paramsV6Bz := store.Get(types.KeyPrefixParams)
	cdc.MustUnmarshal(paramsV6Bz, &paramsV6)

	params.EvmDenom = paramsV6.EvmDenom
	params.ChainConfig = types.ChainConfig{
		HomesteadBlock:      paramsV6.ChainConfig.HomesteadBlock,
		DAOForkBlock:        paramsV6.ChainConfig.DAOForkBlock,
		DAOForkSupport:      paramsV6.ChainConfig.DAOForkSupport,
		EIP150Block:         paramsV6.ChainConfig.EIP150Block,
		EIP150Hash:          paramsV6.ChainConfig.EIP150Hash,
		EIP155Block:         paramsV6.ChainConfig.EIP155Block,
		EIP158Block:         paramsV6.ChainConfig.EIP158Block,
		ByzantiumBlock:      paramsV6.ChainConfig.ByzantiumBlock,
		ConstantinopleBlock: paramsV6.ChainConfig.ConstantinopleBlock,
		PetersburgBlock:     paramsV6.ChainConfig.PetersburgBlock,
		IstanbulBlock:       paramsV6.ChainConfig.IstanbulBlock,
		MuirGlacierBlock:    paramsV6.ChainConfig.MuirGlacierBlock,
		BerlinBlock:         paramsV6.ChainConfig.BerlinBlock,
		LondonBlock:         paramsV6.ChainConfig.LondonBlock,
		ArrowGlacierBlock:   paramsV6.ChainConfig.ArrowGlacierBlock,
		GrayGlacierBlock:    paramsV6.ChainConfig.GrayGlacierBlock,
		MergeNetsplitBlock:  paramsV6.ChainConfig.MergeNetsplitBlock,
		ShanghaiBlock:       paramsV6.ChainConfig.ShanghaiBlock,
		CancunBlock:         paramsV6.ChainConfig.CancunBlock,
	}
	params.AllowUnprotectedTxs = paramsV6.AllowUnprotectedTxs
	params.ActiveStaticPrecompiles = paramsV6.ActivePrecompiles
	params.EVMChannels = paramsV6.EVMChannels

	// set the default access control configuration
	params.AccessControl = types.DefaultAccessControl

	// Migrate old ExtraEIPs from int64 to string. Since no Evmos EIPs have been
	// created before and activators contains only `ethereum_XXXX` activations,
	// all values will be prefixed with `ethereum_`.
	params.ExtraEIPs = make([]string, 0, len(paramsV6.ExtraEIPs))
	for _, eip := range paramsV6.ExtraEIPs {
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
