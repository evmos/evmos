package grpc

import (
	"context"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
)

// GetBaseFee returns the base fee from the feemarket module.
func (gqh *IntegrationGrpcHandler) GetBaseFee() (*feemarkettypes.QueryBaseFeeResponse, error) {
	feeMarketClient := gqh.network.GetFeeMarketClient()
	return feeMarketClient.BaseFee(context.Background(), &feemarkettypes.QueryBaseFeeRequest{})
}
