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

func (suite *InflationTestSuite) TestCalculateEpochMintProvision() {
	bondingParams := DefaultParams()
	bondingParams.ExponentialCalculation.B = sdk.NewDecWithPrec(5, 1)

	epochsPerPeriod := int64(365)

	testCases := []struct {
		name              string
		params            Params
		period            uint64
		expEpochProvision sdk.Dec
		expPass           bool
	}{
		{
			"pass - initial perid",
			DefaultParams(),
			uint64(0),
			sdk.NewDec(int64(847_602)),
			true,
		},
		{
			"pass - period 1",
			DefaultParams(),
			uint64(1),
			sdk.NewDec(int64(436_643)),
			true,
		},
		{
			"pass - period 2",
			DefaultParams(),
			uint64(2),
			sdk.NewDec(int64(231_164)),
			true,
		},
		{
			"pass - period 3",
			DefaultParams(),
			uint64(3),
			sdk.NewDec(int64(128_424)),
			true,
		},
		{
			"pass - period 20",
			DefaultParams(),
			uint64(20),
			sdk.NewDec(int64(25_685)),
			true,
		},
		{
			"pass - period 21",
			DefaultParams(),
			uint64(21),
			sdk.NewDec(int64(25_685)),
			true,
		},
		{
			"pass - with bonding - initial perid",
			bondingParams,
			uint64(0),
			sdk.NewDec(int64(635_702)),
			true,
		},
		{
			"pass - with bonding - period 1",
			bondingParams,
			uint64(1),
			sdk.NewDec(int64(327_482)),
			true,
		},
		{
			"pass - with bonding - period 2",
			bondingParams,
			uint64(2),
			sdk.NewDec(int64(173_373)),
			true,
		},
		{
			"pass - with bonding - period 3",
			bondingParams,
			uint64(3),
			sdk.NewDec(int64(96_318)),
			true,
		},
		{
			"pass - with bonding - period 20",
			bondingParams,
			uint64(20),
			sdk.NewDec(int64(19_264)),
			true,
		},
		{
			"pass - with bonding - period 21",
			bondingParams,
			uint64(21),
			sdk.NewDec(int64(19_263)),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			epochMintProvisions := CalculateEpochMintProvision(tc.params, tc.period, epochsPerPeriod)
			suite.Require().Equal(tc.expEpochProvision, epochMintProvisions)
		})
	}
}
