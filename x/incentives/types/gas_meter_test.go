package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/evmos/evmos/v11/tests"
)

type GasMeterTestSuite struct {
	suite.Suite
}

func TestGasMeterSuite(t *testing.T) {
	suite.Run(t, new(GasMeterTestSuite))
}

func (suite *GasMeterTestSuite) TestGasMeterNew() {
	testCases := []struct {
		name          string
		contract      common.Address
		participant   common.Address
		cumulativeGas uint64
		expectPass    bool
	}{
		{
			"Register GasMeter - pass",
			tests.GenerateAddress(),
			tests.GenerateAddress(),
			100,
			true,
		},
		{
			"Register GasMeter - zero Cumulative Gas",
			tests.GenerateAddress(),
			tests.GenerateAddress(),
			0,
			true,
		},
	}

	for _, tc := range testCases {
		gm := NewGasMeter(tc.contract, tc.participant, tc.cumulativeGas)
		err := gm.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *GasMeterTestSuite) TestGasMeter() {
	testCases := []struct {
		msg        string
		gm         GasMeter
		expectPass bool
	}{
		{
			"Register gas meter - invalid contract address (no hex)",
			GasMeter{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				tests.GenerateAddress().String(),
				10,
			},
			false,
		},
		{
			"Register gas meter - invalid participant address (no hex)",
			GasMeter{
				tests.GenerateAddress().String(),
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				10,
			},
			false,
		},
		{
			"Register gas meter - invalid address (invalid length 1)",
			GasMeter{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				tests.GenerateAddress().String(),
				10,
			},
			false,
		},
		{
			"Register gas meter - invalid address (invalid length 2)",
			GasMeter{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				tests.GenerateAddress().String(),
				10,
			},
			false,
		},
		{
			"pass",
			GasMeter{
				tests.GenerateAddress().String(),
				tests.GenerateAddress().String(),
				10,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.gm.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}
