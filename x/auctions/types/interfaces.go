// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
	"github.com/evmos/evmos/v19/x/evm/types"
)

type EVMKeeper interface {
	GetStaticPrecompileInstance(params *types.Params, address common.Address) (vm.PrecompiledContract, bool, error)
	GetParams(ctx sdk.Context) (params types.Params)
}
