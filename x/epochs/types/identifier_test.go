package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EpochIdentifierTestSuite struct {
	suite.Suite
}

func TestEpochIdentifierTestSuite(t *testing.T) {
	suite.Run(t, new(EpochIdentifierTestSuite))
}

func (suite *EpochIdentifierTestSuite) TestValidateEpochIdentifierInterface() {
	testCases := []struct {
		name       string
		id         interface{}
		expectPass bool
	}{
		{
			"invalid - blank identifier",
			"",
			false,
		},
		{
			"invalid - blank identifier with spaces",
			"   ",
			false,
		},
		{
			"invalid - non-string",
			3,
			false,
		},
		{
			"pass",
			WeekEpochID,
			true,
		},
	}

	for _, tc := range testCases {
		err := ValidateEpochIdentifierInterface(tc.id)

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
