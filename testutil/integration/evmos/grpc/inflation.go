// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package grpc

import (
	"context"

	infltypes "github.com/evmos/evmos/v19/x/inflation/v1/types"
)

// GetPeriod returns the current period.
func (gqh *IntegrationHandler) GetPeriod() (*infltypes.QueryPeriodResponse, error) {
	inflationClient := gqh.network.GetInflationClient()
	return inflationClient.Period(
		context.Background(),
		&infltypes.QueryPeriodRequest{},
	)
}

// GetEpochMintProvision returns the minted provision for the current epoch.
func (gqh *IntegrationHandler) GetEpochMintProvision() (*infltypes.QueryEpochMintProvisionResponse, error) {
	inflationClient := gqh.network.GetInflationClient()
	return inflationClient.EpochMintProvision(
		context.Background(),
		&infltypes.QueryEpochMintProvisionRequest{},
	)
}

// GetSkippedEpochs returns the amount of epochs where inflation was skipped.
func (gqh *IntegrationHandler) GetSkippedEpochs() (*infltypes.QuerySkippedEpochsResponse, error) {
	inflationClient := gqh.network.GetInflationClient()
	return inflationClient.SkippedEpochs(
		context.Background(),
		&infltypes.QuerySkippedEpochsRequest{},
	)
}

// GetCirculatingSupply returns the circulating supply.
func (gqh *IntegrationHandler) GetCirculatingSupply() (*infltypes.QueryCirculatingSupplyResponse, error) {
	inflationClient := gqh.network.GetInflationClient()
	return inflationClient.CirculatingSupply(
		context.Background(),
		&infltypes.QueryCirculatingSupplyRequest{},
	)
}

// GetInflationRate returns the current inflation rate.
func (gqh *IntegrationHandler) GetInflationRate() (*infltypes.QueryInflationRateResponse, error) {
	inflationClient := gqh.network.GetInflationClient()
	return inflationClient.InflationRate(
		context.Background(),
		&infltypes.QueryInflationRateRequest{},
	)
}

// GetInflationParams returns the inflation module parameters.
func (gqh *IntegrationHandler) GetInflationParams() (*infltypes.QueryParamsResponse, error) {
	inflationClient := gqh.network.GetInflationClient()
	return inflationClient.Params(
		context.Background(),
		&infltypes.QueryParamsRequest{},
	)
}
