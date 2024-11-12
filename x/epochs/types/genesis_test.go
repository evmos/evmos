package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	newGen := NewGenesisState([]EpochInfo{})

	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			"valid genesis constructor",
			newGen,
			true,
		},
		{
			"default",
			DefaultGenesisState(),
			true,
		},
		{
			"valid genesis",
			&GenesisState{
				Epochs: []EpochInfo{},
			},
			true,
		},
		{
			"valid genesis - with Epochs",
			&GenesisState{
				Epochs: []EpochInfo{
					{
						Identifier:              WeekEpochID,
						StartTime:               time.Time{},
						Duration:                time.Hour * 24 * 7,
						CurrentEpoch:            0,
						CurrentEpochStartHeight: 0,
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
					{
						Identifier:              DayEpochID,
						StartTime:               time.Time{},
						Duration:                time.Hour * 24,
						CurrentEpoch:            0,
						CurrentEpochStartHeight: 0,
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
				},
			},
			true,
		},
		{
			"invalid genesis - duplicated incentive",
			&GenesisState{
				Epochs: []EpochInfo{
					{
						Identifier:              WeekEpochID,
						StartTime:               time.Time{},
						Duration:                time.Hour * 24 * 7,
						CurrentEpoch:            0,
						CurrentEpochStartHeight: 0,
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
					{
						Identifier:              WeekEpochID,
						StartTime:               time.Time{},
						Duration:                time.Hour * 24 * 7,
						CurrentEpoch:            0,
						CurrentEpochStartHeight: 0,
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
				},
			},
			false,
		},
		{
			"invalid genesis - invalid Epoch",
			&GenesisState{
				Epochs: []EpochInfo{
					{
						Identifier:              WeekEpochID,
						StartTime:               time.Time{},
						Duration:                time.Hour * 24 * 7,
						CurrentEpoch:            -1,
						CurrentEpochStartHeight: 0,
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
