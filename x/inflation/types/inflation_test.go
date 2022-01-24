package types

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateEpochMintProvisions() {
	testCases := []struct {
		name              string
		period            uint64
		expEpochProvision sdk.Dec
		expPass           bool
	}{
		{
			"pass - initial perid",
			uint64(0),
			sdk.NewDec(int64(847_602)),
			true,
		},
		{
			"pass - period 1",
			uint64(1),
			sdk.NewDec(int64(436_643)),
			true,
		},
		{
			"pass - period 2",
			uint64(2),
			sdk.NewDec(int64(231_164)),
			true,
		},
		{
			"pass - period 3",
			uint64(3),
			sdk.NewDec(int64(128_424)),
			true,
		},
		{
			"pass - period 20",
			uint64(20),
			sdk.NewDec(int64(25_685)),
			true,
		},
		{
			"pass - period 21",
			uint64(21),
			sdk.NewDec(int64(25_685)),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			params := DefaultParams()
			epochMintProvisions := CalculateEpochMintProvisions(params, tc.period)
			suite.Require().Equal(tc.expEpochProvision, epochMintProvisions)
		})
	}
}
