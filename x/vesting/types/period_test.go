package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"
)

type PeriodTestSuite struct {
	suite.Suite
}

func TestPeriodSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func period(length int64, amount int64) sdkvesting.Period {
	return sdkvesting.Period{
		Length: length,
		Amount: sdk.NewCoins(sdk.NewInt64Coin("test", amount)),
	}
}

func (suite *PeriodTestSuite) TestReadSchedule() {
	periods := []sdkvesting.Period{period(10, 10), period(20, 20), period(40, 40)}
	total := sdk.NewCoins(sdk.NewInt64Coin("test", 70))
	testCases := []struct {
		time int64
		want int64
	}{
		{0, 0}, {100, 0}, {105, 0}, {110, 10}, {120, 10}, {130, 30},
		{150, 30}, {170, 70}, {180, 70},
	}
	for _, tc := range testCases {
		gotCoins := ReadSchedule(100, 170, periods, total, tc.time)
		got := gotCoins.AmountOf("test").Int64()
		suite.Require().Equal(tc.want, got, "ReadSchedule at %d want %d, got %d", tc.time, tc.want, got)
	}
}

func (suite *PeriodTestSuite) TestDisjunctPeriods() {
	testCases := []struct {
		name      string
		startP    int64
		p         []sdkvesting.Period
		startQ    int64
		q         []sdkvesting.Period
		wantStart int64
		wantEnd   int64
		want      []sdkvesting.Period
	}{
		{
			name:      "empty_empty",
			startP:    0,
			p:         []sdkvesting.Period{},
			startQ:    0,
			q:         []sdkvesting.Period{},
			wantStart: 0,
			want:      []sdkvesting.Period{},
		},
		{
			name:      "some_empty",
			startP:    -123,
			p:         []sdkvesting.Period{period(45, 8), period(67, 13)},
			startQ:    -124,
			q:         []sdkvesting.Period{},
			wantStart: -124,
			wantEnd:   -11,
			want:      []sdkvesting.Period{period(46, 8), period(67, 13)},
		},
		{
			name:      "one_one",
			startP:    0,
			p:         []sdkvesting.Period{period(12, 34)},
			startQ:    0,
			q:         []sdkvesting.Period{period(25, 68)},
			wantStart: 0,
			wantEnd:   25,
			want:      []sdkvesting.Period{period(12, 34), period(13, 68)},
		},
		{
			name:      "tied",
			startP:    12,
			p:         []sdkvesting.Period{period(24, 3)},
			startQ:    0,
			q:         []sdkvesting.Period{period(36, 7)},
			wantStart: 0,
			wantEnd:   36,
			want:      []sdkvesting.Period{period(36, 10)},
		},
		{
			name:      "residual",
			startP:    105,
			p:         []sdkvesting.Period{period(45, 309), period(80, 243), period(30, 401)},
			startQ:    120,
			q:         []sdkvesting.Period{period(40, 823)},
			wantStart: 105,
			wantEnd:   260,
			want:      []sdkvesting.Period{period(45, 309), period(10, 823), period(70, 243), period(30, 401)},
		},
		{
			name:      "typical",
			startP:    1000,
			p:         []sdkvesting.Period{period(100, 25), period(100, 25), period(100, 25), period(100, 25)},
			startQ:    1200,
			q:         []sdkvesting.Period{period(100, 10), period(100, 10), period(100, 10), period(100, 10)},
			wantStart: 1000,
			wantEnd:   1600,
			want:      []sdkvesting.Period{period(100, 25), period(100, 25), period(100, 35), period(100, 35), period(100, 10), period(100, 10)},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got []sdkvesting.Period
				if i == 0 {
					gotStart, gotEnd, got = DisjunctPeriods(tc.startP, tc.startQ, tc.p, tc.q)
				} else {
					gotStart, gotEnd, got = DisjunctPeriods(tc.startQ, tc.startP, tc.q, tc.p)
				}
				suite.Require().Equal(tc.wantStart, gotStart, "wrong start time: got %d, want %d", gotStart, tc.wantStart)
				suite.Require().Equal(tc.wantEnd, gotEnd, "wrong end time: got %d, want %d", gotEnd, tc.wantEnd)
				suite.Require().Equal(len(tc.want), len(got), "wrong number of periods: got %v, want %v", got, tc.want)

				for i, gotPeriod := range got {
					wantPeriod := tc.want[i]
					suite.Require().Equal(wantPeriod.Length, gotPeriod.Length, "period %d length: got %d, want %d", i, gotPeriod.Length, wantPeriod.Length)
					suite.Require().False(gotPeriod.Amount.IsEqual(wantPeriod.Amount), "period %d amount: got %v, want %v", i, gotPeriod.Amount, wantPeriod.Amount)
				}
			}
		})
	}
}

func (suite *PeriodTestSuite) TestConjunctPeriods() {
	testCases := []struct {
		name      string
		startP    int64
		p         []sdkvesting.Period
		startQ    int64
		q         []sdkvesting.Period
		wantStart int64
		wantEnd   int64
		want      []sdkvesting.Period
	}{
		{
			name:      "empty_empty",
			startP:    0,
			p:         []sdkvesting.Period{},
			startQ:    0,
			q:         []sdkvesting.Period{},
			wantStart: 0,
			wantEnd:   0,
			want:      []sdkvesting.Period{},
		},
		{
			name:      "some_empty",
			startP:    -123,
			p:         []sdkvesting.Period{period(45, 8), period(67, 13)},
			startQ:    -124,
			q:         []sdkvesting.Period{},
			wantStart: -124,
			wantEnd:   -124,
			want:      []sdkvesting.Period{},
		},
		{
			name:      "one_one",
			startP:    0,
			p:         []sdkvesting.Period{period(12, 34)},
			startQ:    0,
			q:         []sdkvesting.Period{period(25, 68)},
			wantStart: 0,
			wantEnd:   25,
			want:      []sdkvesting.Period{period(25, 34)},
		},
		{
			name:      "tied",
			startP:    12,
			p:         []sdkvesting.Period{period(24, 3)},
			startQ:    0,
			q:         []sdkvesting.Period{period(36, 7)},
			wantStart: 0,
			wantEnd:   36,
			want:      []sdkvesting.Period{period(36, 3)},
		},
		{
			name:      "residual",
			startP:    105,
			p:         []sdkvesting.Period{period(45, 309), period(80, 243), period(30, 401)},
			startQ:    120,
			q:         []sdkvesting.Period{period(40, 823)},
			wantStart: 105,
			wantEnd:   260,
			want:      []sdkvesting.Period{period(55, 309), period(70, 243), period(30, 271)},
		},
		{
			name:      "overlap",
			startP:    1000,
			p:         []sdkvesting.Period{period(100, 25), period(100, 25), period(100, 25), period(100, 25)},
			startQ:    1200,
			q:         []sdkvesting.Period{period(100, 10), period(100, 10), period(100, 10), period(100, 10)},
			wantStart: 1000,
			wantEnd:   1600,
			want:      []sdkvesting.Period{period(300, 10), period(100, 10), period(100, 10), period(100, 10)},
		},
		{
			name:      "clip",
			startP:    100,
			p:         []sdkvesting.Period{period(10, 1), period(10, 1), period(10, 1), period(10, 1), period(10, 1)},
			startQ:    100,
			q:         []sdkvesting.Period{period(1, 3)},
			wantStart: 100,
			wantEnd:   130,
			want:      []sdkvesting.Period{period(10, 1), period(10, 1), period(10, 1)},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got []sdkvesting.Period
				if i == 0 {
					gotStart, gotEnd, got = ConjunctPeriods(tc.startP, tc.startQ, tc.p, tc.q)
				} else {
					gotStart, gotEnd, got = ConjunctPeriods(tc.startQ, tc.startP, tc.q, tc.p)
				}
				suite.Require().Equal(tc.wantStart, gotStart, "wrong start time: got %d, want %d", gotStart, tc.wantStart)
				suite.Require().Equal(tc.wantEnd, gotEnd, "wrong end time: got %d, want %d", gotEnd, tc.wantEnd)
				suite.Require().Equal(len(tc.want), len(got), "wrong number of periods: got %v, want %v", got, tc.want)

				for i, gotPeriod := range got {
					wantPeriod := tc.want[i]
					suite.Require().Equal(wantPeriod.Length, gotPeriod.Length, "period %d length: got %d, want %d", i, gotPeriod.Length, wantPeriod.Length)
					suite.Require().False(gotPeriod.Amount.IsEqual(wantPeriod.Amount), "period %d amount: got %v, want %v", i, gotPeriod.Amount, wantPeriod.Amount)
				}
			}
		})
	}
}

func (suite *PeriodTestSuite) TestAlignSchedules() {
	p := []sdkvesting.Period{period(10, 50), period(30, 7)}
	q := []sdkvesting.Period{period(40, 6), period(10, 8), period(5, 3)}
	start, end := AlignSchedules(100, 200, p, q)

	suite.Require().Equal(100, start, "want start 100, got %d", start)
	suite.Require().Equal(255, end, "want end 255, got %d", end)
	suite.Require().Equal(10, p[0].Length, "want p first length unchanged, got %d", p[0].Length)
	suite.Require().Equal(140, q[0].Length, "want q first length 140, got %d", q[0].Length)
}
