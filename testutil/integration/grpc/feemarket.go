package grpc

import (
	"context"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
)

// GetBaseFee returns the base fee from the feemarket module.
func (gqh *GrpcQueryHelper) GetBaseFee() (*feemarkettypes.QueryBaseFeeResponse, error) {
	feeMarketClient := gqh.getFeeMarketClient()
	return feeMarketClient.BaseFee(context.Background(), &feemarkettypes.QueryBaseFeeRequest{})
}
