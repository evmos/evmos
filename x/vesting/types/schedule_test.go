package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"
)

type ScheduleTestSuite struct {
	suite.Suite
}

func TestScheduleSuite(t *testing.T) {
	suite.Run(t, new(ScheduleTestSuite))
}

func period(length int64, amount int64) sdkvesting.Period {
	return sdkvesting.Period{
		Length: length,
		Amount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amount)),
	}
}

func (suite *ScheduleTestSuite) TestReadSchedule() {
	testCases := []struct {
		name       string
		startTime  int64
		endTime    int64
		readTime   int64
		totalCoins sdk.Coins
		periods    sdkvesting.Periods
		expAmount  sdk.Coins
	}{
		{
			name:       "empty",
			startTime:  0,
			endTime:    0,
			readTime:   0,
			totalCoins: sdk.Coins{},
			periods:    sdkvesting.Periods{},
			expAmount:  sdk.Coins{},
		},
		{
			name:       "before start time",
			startTime:  100,
			endTime:    200,
			readTime:   0,
			totalCoins: sdk.Coins{},
			periods:    sdkvesting.Periods{},
			expAmount:  sdk.NewCoins(),
		},
		{
			name:       "at start time",
			startTime:  100,
			endTime:    200,
			readTime:   100,
			totalCoins: sdk.Coins{},
			periods:    sdkvesting.Periods{},
			expAmount:  sdk.NewCoins(),
		},
		{
			name:       "at end time",
			startTime:  100,
			endTime:    200,
			readTime:   200,
			totalCoins: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
			periods:    sdkvesting.Periods{},
			expAmount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
		},
		{
			name:       "after end time",
			startTime:  100,
			endTime:    200,
			readTime:   250,
			totalCoins: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
			periods:    sdkvesting.Periods{},
			expAmount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
		},
		{
			name:       "between start and end of first period",
			startTime:  100,
			endTime:    200,
			readTime:   120,
			totalCoins: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 70)),
			periods:    sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expAmount:  sdk.Coins{},
		},
		{
			name:       "at first period end",
			startTime:  100,
			endTime:    200,
			readTime:   125,
			totalCoins: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 70)),
			periods:    sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expAmount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)),
		},
		{
			name:       "between first and second period",
			startTime:  100,
			endTime:    200,
			readTime:   150,
			totalCoins: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 70)),
			periods:    sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expAmount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)),
		},
		{
			name:       "last period, before end time",
			startTime:  100,
			endTime:    200,
			readTime:   199,
			totalCoins: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 70)),
			periods:    sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expAmount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 30)),
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			coins := ReadSchedule(tc.startTime, tc.endTime, tc.periods, tc.totalCoins, tc.readTime)
			suite.Require().Equal(tc.expAmount, coins)
		})
	}
}

func (suite *ScheduleTestSuite) TestReadPastPeriodCount() {
	testCases := []struct {
		name      string
		startTime int64
		endTime   int64
		readTime  int64
		periods   sdkvesting.Periods
		expCount  int
	}{
		{
			name:      "empty",
			startTime: 0,
			endTime:   0,
			readTime:  0,
			periods:   sdkvesting.Periods{},
			expCount:  0,
		},
		{
			name:      "single period, at end time",
			startTime: 100,
			endTime:   150,
			readTime:  150,
			periods:   sdkvesting.Periods{period(50, 50)},
			expCount:  1,
		},
		{
			name:      "before start time",
			startTime: 100,
			endTime:   170,
			readTime:  0,
			periods:   sdkvesting.Periods{period(10, 10), period(20, 20), period(40, 40)},
			expCount:  0,
		},
		{
			name:      "at start time",
			startTime: 100,
			endTime:   200,
			readTime:  100,
			periods:   sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expCount:  0,
		},
		{
			name:      "at end of first period",
			startTime: 100,
			endTime:   200,
			readTime:  125,
			periods:   sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expCount:  1,
		},
		{
			name:      "after first period end, before second period end",
			startTime: 100,
			endTime:   200,
			readTime:  135,
			periods:   sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expCount:  1,
		},
		{
			name:      "at end time, all periods passed",
			startTime: 100,
			endTime:   200,
			readTime:  200,
			periods:   sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expCount:  3,
		},
		{
			name:      "after end time, all periods passed",
			startTime: 100,
			endTime:   200,
			readTime:  250,
			periods:   sdkvesting.Periods{period(25, 10), period(50, 20), period(25, 40)},
			expCount:  3,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			count := ReadPastPeriodCount(tc.startTime, tc.endTime, tc.periods, tc.readTime)
			suite.Require().Equal(tc.expCount, count)
		})
	}
}

func (suite *ScheduleTestSuite) TestDisjunctPeriods() {
	testCases := []struct {
		name         string
		startPeriodA int64
		startPeriodB int64
		periodsA     sdkvesting.Periods
		periodsB     sdkvesting.Periods
		expStartTime int64
		expEndTime   int64
		expPeriods   sdkvesting.Periods
	}{
		{
			name:         "empty_empty",
			startPeriodA: 0,
			periodsA:     sdkvesting.Periods{},
			startPeriodB: 0,
			periodsB:     sdkvesting.Periods{},
			expStartTime: 0,
			expPeriods:   sdkvesting.Periods{},
		},
		{
			name:         "some_empty",
			startPeriodA: -123,
			periodsA:     sdkvesting.Periods{period(45, 8), period(67, 13)},
			startPeriodB: -124,
			periodsB:     sdkvesting.Periods{},
			expStartTime: -124,
			expEndTime:   -11,
			expPeriods:   sdkvesting.Periods{period(46, 8), period(67, 13)},
		},
		{
			name:         "one_one",
			startPeriodA: 0,
			periodsA:     sdkvesting.Periods{period(12, 34)},
			startPeriodB: 0,
			periodsB:     sdkvesting.Periods{period(25, 68)},
			expStartTime: 0,
			expEndTime:   25,
			expPeriods:   sdkvesting.Periods{period(12, 34), period(13, 68)},
		},
		{
			name:         "tied",
			startPeriodA: 12,
			periodsA:     sdkvesting.Periods{period(24, 3)},
			startPeriodB: 0,
			periodsB:     sdkvesting.Periods{period(36, 7)},
			expStartTime: 0,
			expEndTime:   36,
			expPeriods:   sdkvesting.Periods{period(36, 10)},
		},
		{
			name:         "residual",
			startPeriodA: 105,
			periodsA:     sdkvesting.Periods{period(45, 309), period(80, 243), period(30, 401)},
			startPeriodB: 120,
			periodsB:     sdkvesting.Periods{period(40, 823)},
			expStartTime: 105,
			expEndTime:   260,
			expPeriods:   sdkvesting.Periods{period(45, 309), period(10, 823), period(70, 243), period(30, 401)},
		},
		{
			name:         "typical",
			startPeriodA: 1000,
			periodsA:     sdkvesting.Periods{period(100, 25), period(100, 25), period(100, 25), period(100, 25)},
			startPeriodB: 1200,
			periodsB:     sdkvesting.Periods{period(100, 10), period(100, 10), period(100, 10), period(100, 10)},
			expStartTime: 1000,
			expEndTime:   1600,
			expPeriods:   sdkvesting.Periods{period(100, 25), period(100, 25), period(100, 35), period(100, 35), period(100, 10), period(100, 10)},
		},
	}
	for _, tc := range testCases { //nolint:dupl
		suite.Run(tc.name, func() {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got sdkvesting.Periods
				if i == 0 {
					gotStart, gotEnd, got = DisjunctPeriods(tc.startPeriodA, tc.startPeriodB, tc.periodsA, tc.periodsB)
				} else {
					gotStart, gotEnd, got = DisjunctPeriods(tc.startPeriodB, tc.startPeriodA, tc.periodsB, tc.periodsA)
				}
				suite.Require().Equal(tc.expStartTime, gotStart)
				suite.Require().Equal(tc.expEndTime, gotEnd)
				suite.Require().Equal(len(tc.expPeriods), len(got))

				for i, gotPeriod := range got {
					wantPeriod := tc.expPeriods[i]
					suite.Require().Equal(wantPeriod.Length, gotPeriod.Length)
					suite.Require().True(gotPeriod.Amount.IsEqual(wantPeriod.Amount),
						"period %d amount: got %v, expPeriods %v", i, gotPeriod.Amount, wantPeriod.Amount,
					)
				}
			}
		})
	}
}

func (suite *ScheduleTestSuite) TestConjunctPeriods() {
	testCases := []struct {
		name         string
		startPeriodA int64
		startPeriodB int64
		periodsA     sdkvesting.Periods
		periodsB     sdkvesting.Periods
		expStartTime int64
		expEndTime   int64
		expPeriods   sdkvesting.Periods
	}{
		{
			name:         "empty_empty",
			startPeriodA: 0,
			periodsA:     sdkvesting.Periods{},
			startPeriodB: 0,
			periodsB:     sdkvesting.Periods{},
			expStartTime: 0,
			expEndTime:   0,
			expPeriods:   sdkvesting.Periods{},
		},
		{
			name:         "some_empty",
			startPeriodA: -123,
			periodsA:     sdkvesting.Periods{period(45, 8), period(67, 13)},
			startPeriodB: -124,
			periodsB:     sdkvesting.Periods{},
			expStartTime: -124,
			expEndTime:   -124,
			expPeriods:   sdkvesting.Periods{},
		},
		{
			name:         "one_one",
			startPeriodA: 0,
			periodsA:     sdkvesting.Periods{period(12, 34)},
			startPeriodB: 0,
			periodsB:     sdkvesting.Periods{period(25, 68)},
			expStartTime: 0,
			expEndTime:   25,
			expPeriods:   sdkvesting.Periods{period(25, 34)},
		},
		{
			name:         "tied",
			startPeriodA: 12,
			periodsA:     sdkvesting.Periods{period(24, 3)},
			startPeriodB: 0,
			periodsB:     sdkvesting.Periods{period(36, 7)},
			expStartTime: 0,
			expEndTime:   36,
			expPeriods:   sdkvesting.Periods{period(36, 3)},
		},
		{
			name:         "residual",
			startPeriodA: 105,
			periodsA:     sdkvesting.Periods{period(45, 309), period(80, 243), period(30, 401)},
			startPeriodB: 120,
			periodsB:     sdkvesting.Periods{period(40, 823)},
			expStartTime: 105,
			expEndTime:   260,
			expPeriods:   sdkvesting.Periods{period(55, 309), period(70, 243), period(30, 271)},
		},
		{
			name:         "overlap",
			startPeriodA: 1000,
			periodsA:     sdkvesting.Periods{period(100, 25), period(100, 25), period(100, 25), period(100, 25)},
			startPeriodB: 1200,
			periodsB:     sdkvesting.Periods{period(100, 10), period(100, 10), period(100, 10), period(100, 10)},
			expStartTime: 1000,
			expEndTime:   1600,
			expPeriods:   sdkvesting.Periods{period(300, 10), period(100, 10), period(100, 10), period(100, 10)},
		},
		{
			name:         "clip",
			startPeriodA: 100,
			periodsA:     sdkvesting.Periods{period(10, 1), period(10, 1), period(10, 1), period(10, 1), period(10, 1)},
			startPeriodB: 100,
			periodsB:     sdkvesting.Periods{period(1, 3)},
			expStartTime: 100,
			expEndTime:   130,
			expPeriods:   sdkvesting.Periods{period(10, 1), period(10, 1), period(10, 1)},
		},
	}
	for _, tc := range testCases { //nolint:dupl
		suite.Run(tc.name, func() {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got sdkvesting.Periods
				if i == 0 {
					gotStart, gotEnd, got = ConjunctPeriods(tc.startPeriodA, tc.startPeriodB, tc.periodsA, tc.periodsB)
				} else {
					gotStart, gotEnd, got = ConjunctPeriods(tc.startPeriodB, tc.startPeriodA, tc.periodsB, tc.periodsA)
				}
				suite.Require().Equal(tc.expStartTime, gotStart)
				suite.Require().Equal(tc.expEndTime, gotEnd)
				suite.Require().Equal(len(tc.expPeriods), len(got))

				for i, gotPeriod := range got {
					wantPeriod := tc.expPeriods[i]
					suite.Require().Equal(wantPeriod.Length, gotPeriod.Length)
					suite.Require().True(gotPeriod.Amount.IsEqual(wantPeriod.Amount),
						"period %d amount: got %v, expPeriods %v", i, gotPeriod.Amount, wantPeriod.Amount,
					)
				}
			}
		})
	}
}

func (suite *ScheduleTestSuite) TestAlignSchedules() {
	testCases := []struct {
		name             string
		startTimePeriodA int64
		startTimePeriodB int64
		periodsA         sdkvesting.Periods
		periodsB         sdkvesting.Periods
		expStartTime     int64
		expEndTime       int64
		expPeriodsALen   int64
		expPeriodsBLen   int64
	}{
		{
			"empty values",
			0,
			0,
			sdkvesting.Periods{},
			sdkvesting.Periods{},
			0,
			0,
			0,
			0,
		},
		{
			"same periods and start time",
			100,
			100,
			sdkvesting.Periods{period(10, 50), period(30, 7)},
			sdkvesting.Periods{period(10, 50), period(30, 7)},
			100,
			140,
			10,
			10,
		},
		{
			"same periods, different start time",
			100,
			200,
			sdkvesting.Periods{period(3600, 50), period(3600, 50)}, // 1 hr ea
			sdkvesting.Periods{period(3600, 50), period(3600, 50)}, // 1 hr ea
			100,
			7400, // 3600 + 3600 + 200
			3600,
			3700, // 3600 + 100 diff
		},
		{
			"different periods, same start time, same end time",
			100,
			100,
			sdkvesting.Periods{period(3600, 50), period(3600, 50)},
			sdkvesting.Periods{period(1800, 25), period(1800, 25), period(1800, 25), period(1800, 25)},
			100,
			7300, // duration + start time = 7200 + 100
			3600,
			1800,
		},
		{
			"different periods, same start time (0), same end time",
			0,
			0,
			sdkvesting.Periods{period(3600, 50), period(3600, 50)},
			sdkvesting.Periods{period(1800, 25), period(1800, 25), period(1800, 25), period(1800, 25)},
			0,
			7200,
			3600,
			1800,
		},
		{
			"different periods, same start time, different end time",
			0,
			0,
			sdkvesting.Periods{period(3600, 50)},
			sdkvesting.Periods{period(1800, 25), period(1800, 25), period(1800, 25), period(1800, 25)},
			0,
			7200,
			3600,
			1800,
		},
		{
			"one empty period, same start time (0), different end time",
			0,
			0,
			sdkvesting.Periods{},
			sdkvesting.Periods{period(3600, 50)},
			0,
			3600,
			0,
			3600,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			start, end := AlignSchedules(tc.startTimePeriodA, tc.startTimePeriodB, tc.periodsA, tc.periodsB)
			suite.Require().Equal(tc.expStartTime, start)
			suite.Require().Equal(tc.expEndTime, end)

			if len(tc.periodsA) > 0 {
				suite.Require().Equal(tc.expPeriodsALen, tc.periodsA[0].Length)
			}
			if len(tc.periodsB) > 0 {
				suite.Require().Equal(tc.expPeriodsBLen, tc.periodsB[0].Length)
			}
		})
	}
}
