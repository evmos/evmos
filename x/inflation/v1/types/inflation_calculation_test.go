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
			// (300_000_000 * (1 - 0.5) ** 0 + 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("282534246575342465753425.000000000000000000"),
			true,
		},
		{
			"pass - period 1",
			DefaultParams(),
			uint64(1),
			math.LegacyOneDec(),
			// (300_000_000 * (1 - 0.5) ** 1 + 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("145547945205479452054795.000000000000000000"),
			true,
		},
		{
			"pass - period 2",
			DefaultParams(),
			uint64(2),
			math.LegacyOneDec(),
			// (300_000_000 * (1 - 0.5) ** 2 + 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("77054794520547945205479.000000000000000000"),
			true,
		},
		{
			"pass - period 3",
			DefaultParams(),
			uint64(3),
			math.LegacyOneDec(),
			// (300_000_000 * (1 - 0.5) ** 3 + 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("42808219178082191780822.000000000000000000"),
			true,
		},
		{
			"pass - period 20",
			DefaultParams(),
			uint64(20),
			math.LegacyOneDec(),
			// (300_000_000 * (1 - 0.5) ** 20 + 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("8561905116251070205479.000000000000000000"),
			true,
		},
		{
			"pass - period 21",
			DefaultParams(),
			uint64(21),
			math.LegacyOneDec(),
			// (300_000_000 * (1 - 0.5) ** 21 + 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("8561774475933754280822.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - initial period",
			bondingParams,
			uint64(0),
			math.LegacyZeroDec(),
			// (300_000_000 * (1 - 0.5) ** 0 * (1 + 0.4)+ 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("392123287671232882795743.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 1",
			bondingParams,
			uint64(1),
			math.LegacyZeroDec(),
			// (300_000_000 * (1 - 0.5) ** 1 * (1 + 0.4)+ 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("200342465753424660575954.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 2",
			bondingParams,
			uint64(2),
			math.LegacyZeroDec(),
			// (300_000_000 * (1 - 0.5) ** 2 * (1 + 0.4)+ 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("104452054794520549466059.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 3",
			bondingParams,
			uint64(3),
			math.LegacyZeroDec(),
			// (300_000_000 * (1 - 0.5) ** 3 * (1 + 0.4)+ 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("56506849315068493911112.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 20",
			bondingParams,
			uint64(20),
			math.LegacyZeroDec(),
			// (300_000_000 * (1 - 0.5) ** 20 * (1 + 0.4)+ 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("8562009628504922945212.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 21",
			bondingParams,
			uint64(21),
			math.LegacyZeroDec(),
			// (300_000_000 * (1 - 0.5) ** 21 * (1 + 0.4)+ 9_375_000) / 3 / 365 * 10 ** 18
			math.LegacyMustNewDecFromStr("8561826732060680650688.000000000000000000"),
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

			// Here we use a relative error because the expected values are computed with another
			// software and can be slightly differences. Accepted error is less than 0.001%.
			tol := math.LegacyNewDecWithPrec(1, 5)
			relativeError := tc.expEpochProvision.Sub(epochMintProvisions).Abs().Quo(tc.expEpochProvision)
			valid := relativeError.LTE(tol)

			suite.Require().True(valid)
		})
	}
}
