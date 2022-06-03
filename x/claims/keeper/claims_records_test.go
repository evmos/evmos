package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v5/x/claims/types"
)

func (suite *KeeperTestSuite) TestsClaimsRecords() {
	addr1, err := sdk.AccAddressFromBech32("evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625")
	suite.Require().NoError(err)
	addr2, err := sdk.AccAddressFromBech32("evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms")
	suite.Require().NoError(err)

	cr1 := types.NewClaimsRecord(sdk.NewInt(1000))
	cr2 := types.NewClaimsRecord(sdk.NewInt(200))
	cr2.MarkClaimed(types.ActionDelegate)

	expRecords := []types.ClaimsRecordAddress{
		{
			Address:                addr2.String(),
			InitialClaimableAmount: cr2.InitialClaimableAmount,
			ActionsCompleted:       cr2.ActionsCompleted,
		},
		{
			Address:                addr1.String(),
			InitialClaimableAmount: cr1.InitialClaimableAmount,
			ActionsCompleted:       cr1.ActionsCompleted,
		},
	}

	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, cr1)
	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr2, cr2)

	records := suite.app.ClaimsKeeper.GetClaimsRecords(suite.ctx)
	suite.Require().Equal(expRecords, records)
}
