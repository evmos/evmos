package types_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/x/incentives/types"
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
			"Register types.GasMeter - pass",
			utiltx.GenerateAddress(),
			utiltx.GenerateAddress(),
			100,
			true,
		},
		{
			"Register types.GasMeter - zero Cumulative Gas",
			utiltx.GenerateAddress(),
			utiltx.GenerateAddress(),
			0,
			true,
		},
	}

	for _, tc := range testCases {
		gm := types.NewGasMeter(tc.contract, tc.participant, tc.cumulativeGas)
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
		gm         types.GasMeter
		expectPass bool
	}{
		{
			"Register gas meter - invalid contract address (no hex)",
			types.GasMeter{
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				utiltx.GenerateAddress().String(),
				10,
			},
			false,
		},
		{
			"Register gas meter - invalid participant address (no hex)",
			types.GasMeter{
				utiltx.GenerateAddress().String(),
				"0x5dCA2483280D9727c80b5518faC4556617fb19ZZ",
				10,
			},
			false,
		},
		{
			"Register gas meter - invalid address (invalid length 1)",
			types.GasMeter{
				"0x5dCA2483280D9727c80b5518faC4556617fb19",
				utiltx.GenerateAddress().String(),
				10,
			},
			false,
		},
		{
			"Register gas meter - invalid address (invalid length 2)",
			types.GasMeter{
				"0x5dCA2483280D9727c80b5518faC4556617fb194FFF",
				utiltx.GenerateAddress().String(),
				10,
			},
			false,
		},
		{
			"pass",
			types.GasMeter{
				utiltx.GenerateAddress().String(),
				utiltx.GenerateAddress().String(),
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
