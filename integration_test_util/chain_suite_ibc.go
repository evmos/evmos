package integration_test_util

import (
	"fmt"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"strings"
)

// QueryDenomHash returns the denom hash of given denom trace information.
func (suite *ChainIntegrationTestSuite) QueryDenomHash(port, channel, denom string) string {
	denomHashRes, err := suite.QueryClients.IbcTransfer.DenomHash(suite.CurrentContext, &ibctransfertypes.QueryDenomHashRequest{
		Trace: fmt.Sprintf("%s/%s/%s", port, channel, denom),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(denomHashRes)
	suite.Require().NotEmpty(denomHashRes.Hash)
	suite.Require().Falsef(strings.HasPrefix(denomHashRes.Hash, "ibc/"), "denom hash %s can not has prefix ibc/")
	suite.Require().Equalf(strings.ToUpper(denomHashRes.Hash), denomHashRes.Hash, "denom hash %s must be all uppercase")
	return fmt.Sprintf("ibc/%s", denomHashRes.Hash)
}
