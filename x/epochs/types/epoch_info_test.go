package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type EpochInfoTestSuite struct {
	suite.Suite
}

func TestEpochInfoTestSuite(t *testing.T) {
	suite.Run(t, new(EpochInfoTestSuite))
}

func (suite *EpochInfoTestSuite) TestStartEndEpoch() {
	startTime := time.Now()
	duration := time.Hour * 24
	ei := EpochInfo{StartTime: startTime, Duration: duration}

	ei.StartInitialEpoch()
	suite.Require().True(ei.EpochCountingStarted)
	suite.Require().Equal(int64(1), ei.CurrentEpoch)
	suite.Require().Equal(startTime, ei.CurrentEpochStartTime)

	ei.EndEpoch()
	suite.Require().Equal(int64(2), ei.CurrentEpoch)
	suite.Require().Equal(startTime.Add(duration), ei.CurrentEpochStartTime)
}

func (suite *EpochInfoTestSuite) TestValidateEpochInfo() {
	testCases := []struct {
		name       string
		ei         EpochInfo
		expectPass bool
	}{
		{
			"invalid - blank identifier",
			EpochInfo{
				"  ",
				time.Now(),
				time.Hour * 24,
				1,
				time.Now(),
				true,
				1,
			},
			false,
		},
		{
			"invalid - epoch duration zero",
			EpochInfo{
				"week",
				time.Now(),
				time.Hour * 0,
				1,
				time.Now(),
				true,
				1,
			},
			false,
		},
		{
			"invalid - negative current epoch",
			EpochInfo{
				"week",
				time.Now(),
				time.Hour * 24,
				-1,
				time.Now(),
				true,
				1,
			},
			false,
		},
		{
			"invalid - negative epoch start height",
			EpochInfo{
				"week",
				time.Now(),
				time.Hour * 24,
				1,
				time.Now(),
				true,
				-1,
			},
			false,
		},
		{
			"pass",
			EpochInfo{
				"week",
				time.Now(),
				time.Hour * 24,
				1,
				time.Now(),
				true,
				1,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.ei.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
