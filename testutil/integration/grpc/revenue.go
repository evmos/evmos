package grpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	revtypes "github.com/evmos/evmos/v14/x/revenue/v1/types"
)

// GetRevenue returns the revenue for the given address.
func (gqh *GrpcQueryHelper) GetRevenue(address common.Address) (*revtypes.QueryRevenueResponse, error) {
	revenueClient := gqh.getRevenueClient()
	return revenueClient.Revenue(context.Background(), &revtypes.QueryRevenueRequest{
		ContractAddress: address.String(),
	})
}

func (gqh *GrpcQueryHelper) GetRevenueParams() (*revtypes.QueryParamsResponse, error) {
	revenueClient := gqh.getRevenueClient()
	return revenueClient.Params(context.Background(), &revtypes.QueryParamsRequest{})
}
