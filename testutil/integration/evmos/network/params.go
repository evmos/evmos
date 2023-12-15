// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	infltypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
	revtypes "github.com/evmos/evmos/v16/x/revenue/v1/types"
)

func (n *IntegrationNetwork) UpdateEvmParams(params evmtypes.Params) error {
	// FIXME this will not work with sdk v0.50
	// To make this work, will need to do it as in the real chain
	// (submit gov proposal & vote for it)
	// An alternative is to create a network with a custom genesis
	// where the params are the ones desired

	// return n.app.EvmKeeper.SetParams(n.ctx, params)
	return errors.New("Not implemented")
}

func (n *IntegrationNetwork) UpdateRevenueParams(params revtypes.Params) error {
	// FIXME this will not work with sdk v0.50
	// To make this work, will need to do it as in the real chain
	// (submit gov proposal & vote for it)
	// An alternative is to create a network with a custom genesis
	// where the params are the ones desired

	// return n.app.RevenueKeeper.SetParams(n.ctx, params)
	return errors.New("Not implemented")
}

func (n *IntegrationNetwork) UpdateInflationParams(params infltypes.Params) error {
	// FIXME this will not work with sdk v0.50
	// To make this work, will need to do it as in the real chain
	// (submit gov proposal & vote for it)
	// An alternative is to create a network with a custom genesis
	// where the params are the ones desired

	// return n.app.InflationKeeper.SetParams(n.ctx, params)
	return errors.New("Not implemented")
}

func (n *IntegrationNetwork) UpdateGovParams(params govtypes.Params) error {
	// FIXME this will not work with sdk v0.50
	// To make this work, will need to do it as in the real chain
	// (submit gov proposal & vote for it)
	// An alternative is to create a network with a custom genesis
	// where the params are the ones desired

	// return n.app.GovKeeper.Params.Set(n.ctx, params)
	return errors.New("Not implemented")
}
