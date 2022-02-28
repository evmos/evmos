package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (suite *ScheduleTestSuite) TestScaleCoins() {
	testCases := []struct {
		name        string
		coins       sdk.Coins
		scale       sdk.Dec
		expectCoins sdk.Coins
	}{
		{
			"one coin",
			sdk.NewCoins(sdk.NewCoin("evmos", sdk.NewInt(10))),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewCoins(sdk.NewCoin("evmos", sdk.NewInt(5))),
		},
		{
			"zero coin",
			sdk.NewCoins(sdk.NewCoin("evmos", sdk.ZeroInt())),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewCoins(sdk.NewCoin("evmos", sdk.ZeroInt())),
		},
		{
			"two coins",
			sdk.NewCoins(
				sdk.NewCoin("evmos", sdk.NewInt(10)),
				sdk.NewCoin("photon", sdk.NewInt(20)),
			),
			sdk.NewDecWithPrec(50, 2),
			sdk.NewCoins(
				sdk.NewCoin("evmos", sdk.NewInt(5)),
				sdk.NewCoin("photon", sdk.NewInt(10)),
			),
		},
		{
			"zero scale",
			sdk.NewCoins(sdk.NewCoin("evmos", sdk.NewInt(10))),
			sdk.ZeroDec(),
			sdk.Coins(nil),
		},
	}
	for _, tc := range testCases {
		coins := ScaleCoins(tc.coins, tc.scale)
		suite.Require().Equal(tc.expectCoins, coins)
	}
}
