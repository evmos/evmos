// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ante_test

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	evmante "github.com/evmos/evmos/v18/x/evm/ante"
)

func (suite *EvmAnteTestSuite) TestBuildEvmExecutionCtx() {
	network := network.New()

	ctx := evmante.BuildEvmExecutionCtx(network.GetContext())

	suite.Equal(storetypes.GasConfig{}, ctx.KVGasConfig())
	suite.Equal(storetypes.GasConfig{}, ctx.TransientKVGasConfig())
}
