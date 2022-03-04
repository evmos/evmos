package types

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateEpochMintProvision() {
	bondingParams := DefaultParams()
	bondingParams.ExponentialCalculation.MaxVariance = sdk.NewDecWithPrec(40, 2)
	epochsPerPeriod := int64(365)

	testCases := []struct {
		name              string
		params            Params
		period            uint64
		bondedRatio       sdk.Dec
		expEpochProvision sdk.Dec
		expPass           bool
	}{
		{
			"pass - initial perid",
			DefaultParams(),
			uint64(0),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("847602739726027397260274.000000000000000000"),
			true,
		},
		{
			"pass - period 1",
			DefaultParams(),
			uint64(1),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("436643835616438356164384.000000000000000000"),
			true,
		},
		{
			"pass - period 2",
			DefaultParams(),
			uint64(2),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("231164383561643835616438.000000000000000000"),
			true,
		},
		{
			"pass - period 3",
			DefaultParams(),
			uint64(3),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("128424657534246575342466.000000000000000000"),
			true,
		},
		{
			"pass - period 20",
			DefaultParams(),
			uint64(20),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("25685715348753210410959.000000000000000000"),
			true,
		},
		{
			"pass - period 21",
			DefaultParams(),
			uint64(21),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("25685323427801262739726.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - initial period",
			bondingParams,
			uint64(0),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("1186643835616438356164384.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 1",
			bondingParams,
			uint64(1),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("611301369863013698630137.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 2",
			bondingParams,
			uint64(2),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("323630136986301369863014.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 3",
			bondingParams,
			uint64(3),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("179794520547945205479452.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 20",
			bondingParams,
			uint64(20),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("35960001488254494575342.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 21",
			bondingParams,
			uint64(21),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("35959452798921767835616.000000000000000000"),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			epochMintProvisions := CalculateEpochMintProvision(
				tc.params,
				tc.period,
				epochsPerPeriod,
				tc.bondedRatio,
			)

			suite.Require().Equal(tc.expEpochProvision, epochMintProvisions)
		})
	}
}
