// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
	revtypes "github.com/evmos/evmos/v16/x/revenue/v1/types"
)

type UpdateParamsInput struct {
	Tf      factory.TxFactory
	Network network.Network
	Pk      cryptotypes.PrivKey
	Params  interface{}
}

var authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()

// UpdateEvmParams helper function to update the EVM module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func UpdateEvmParams(input UpdateParamsInput) error {
	return updateModuleParams[evmtypes.Params](input, evmtypes.ModuleName)
}

// UpdateRevenueParams helper function to update the revenue module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func UpdateRevenueParams(input UpdateParamsInput) error {
	return updateModuleParams[revtypes.Params](input, revtypes.ModuleName)
}

// UpdateInflationParams helper function to update the inflation module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func UpdateInflationParams(input UpdateParamsInput) error {
	return updateModuleParams[infltypes.Params](input, infltypes.ModuleName)
}

// UpdateGovParams helper function to update the governance module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func UpdateGovParams(input UpdateParamsInput) error {
	return updateModuleParams[govv1types.Params](input, govtypes.ModuleName)
}

// UpdateFeeMarketParams helper function to update the feemarket module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func UpdateFeeMarketParams(input UpdateParamsInput) error {
	return updateModuleParams[feemarkettypes.Params](input, feemarkettypes.ModuleName)
}

// updateModuleParams helper function to update module parameters
// It submits an update params proposal, votes for it, and waits till it passes
func updateModuleParams[T interface{}](input UpdateParamsInput, moduleName string) error {
	newParams, ok := input.Params.(T)
	if !ok {
		return fmt.Errorf("invalid params type %T for module %s", input.Params, moduleName)
	}

	proposalMsg := createProposalMsg(newParams, moduleName)

	title := fmt.Sprintf("Update %s params", moduleName)
	proposalID, err := SubmitProposal(input.Tf, input.Network, input.Pk, title, proposalMsg)
	if err != nil {
		return err
	}

	return ApproveProposal(input.Tf, input.Network, input.Pk, proposalID)
}

// createProposalMsg creates the module-specific update params message
func createProposalMsg(params interface{}, name string) sdk.Msg {
	switch name {
	case evmtypes.ModuleName:
		return &evmtypes.MsgUpdateParams{Authority: authority, Params: params.(evmtypes.Params)}
	case revtypes.ModuleName:
		return &revtypes.MsgUpdateParams{Authority: authority, Params: params.(revtypes.Params)}
	case infltypes.ModuleName:
		return &infltypes.MsgUpdateParams{Authority: authority, Params: params.(infltypes.Params)}
	case govtypes.ModuleName:
		return &govv1types.MsgUpdateParams{Authority: authority, Params: params.(govv1types.Params)}
	case feemarkettypes.ModuleName:
		return &feemarkettypes.MsgUpdateParams{Authority: authority, Params: params.(feemarkettypes.Params)}
	default:
		return nil
	}
}
