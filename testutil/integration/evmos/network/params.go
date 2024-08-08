// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	feemarketypes "github.com/evmos/evmos/v19/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v19/x/inflation/v1/types"
)

func (n *IntegrationNetwork) UpdateEvmParams(params evmtypes.Params) error {
	return n.app.EvmKeeper.SetParams(n.ctx, params)
}

func (n *IntegrationNetwork) UpdateFeeMarketParams(params feemarketypes.Params) error {
	return n.app.FeeMarketKeeper.SetParams(n.ctx, params)
}

func (n *IntegrationNetwork) UpdateInflationParams(params infltypes.Params) error {
	return n.app.InflationKeeper.SetParams(n.ctx, params)
}
