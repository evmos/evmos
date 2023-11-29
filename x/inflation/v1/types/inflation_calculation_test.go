package types

import (
	fmt "fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateEpochMintProvision() {
	bondingParams := DefaultParams()
	bondingParams.ExponentialCalculation.MaxVariance = math.LegacyNewDecWithPrec(40, 2)
	epochsPerPeriod := int64(365)

	testCases := []struct {
		name              string
		params            Params
		period            uint64
		bondedRatio       math.LegacyDec
		expEpochProvision math.LegacyDec
		expPass           bool
	}{
		{
			"pass - initial perid",
			DefaultParams(),
			uint64(0),
			math.LegacyOneDec(),
			math.LegacyMustNewDecFromStr("847602739726027397260274.000000000000000000"),
			true,
		},
		{
			"pass - period 1",
			DefaultParams(),
			uint64(1),
			math.LegacyOneDec(),
			math.LegacyMustNewDecFromStr("436643835616438356164384.000000000000000000"),
			true,
		},
		{
			"pass - period 2",
			DefaultParams(),
			uint64(2),
			math.LegacyOneDec(),
			math.LegacyMustNewDecFromStr("231164383561643835616438.000000000000000000"),
			true,
		},
		{
			"pass - period 3",
			DefaultParams(),
			uint64(3),
			math.LegacyOneDec(),
			math.LegacyMustNewDecFromStr("128424657534246575342466.000000000000000000"),
			true,
		},
		{
			"pass - period 20",
			DefaultParams(),
			uint64(20),
			math.LegacyOneDec(),
			math.LegacyMustNewDecFromStr("25685715348753210410959.000000000000000000"),
			true,
		},
		{
			"pass - period 21",
			DefaultParams(),
			uint64(21),
			math.LegacyOneDec(),
			math.LegacyMustNewDecFromStr("25685323427801262739726.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - initial period",
			bondingParams,
			uint64(0),
			math.LegacyZeroDec(),
			math.LegacyMustNewDecFromStr("1186643835616438356164384.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 1",
			bondingParams,
			uint64(1),
			math.LegacyZeroDec(),
			math.LegacyMustNewDecFromStr("611301369863013698630137.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 2",
			bondingParams,
			uint64(2),
			math.LegacyZeroDec(),
			math.LegacyMustNewDecFromStr("323630136986301369863014.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 3",
			bondingParams,
			uint64(3),
			math.LegacyZeroDec(),
			math.LegacyMustNewDecFromStr("179794520547945205479452.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 20",
			bondingParams,
			uint64(20),
			math.LegacyZeroDec(),
			math.LegacyMustNewDecFromStr("35960001488254494575342.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 21",
			bondingParams,
			uint64(21),
			math.LegacyZeroDec(),
			math.LegacyMustNewDecFromStr("35959452798921767835616.000000000000000000"),
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
