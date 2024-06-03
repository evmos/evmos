package gov_test

import (
	"fmt"

	"cosmossdk.io/math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/precompiles/testutil"

	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/gov"
	utiltx "github.com/evmos/evmos/v18/testutil/tx"
)

func (s *PrecompileTestSuite) TestVote() {
	method := s.precompile.Methods[gov.VoteMethod]
	newVoterAddr := utiltx.GenerateAddress()
	const proposalId uint64 = 1
	const option uint8 = 1
	const metadata = "metadata"

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
		},
		{
			"fail - invalid voter address",
			func() []interface{} {
				return []interface{}{
					"",
					proposalId,
					option,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid voter address",
		},
		{
			"fail - invalid voter address",
			func() []interface{} {
				return []interface{}{
					common.Address{},
					proposalId,
					option,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid voter address",
		},
		{
			"fail - using a different voter address",
			func() []interface{} {
				return []interface{}{
					newVoterAddr,
					proposalId,
					option,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"does not match the voter address",
		},
		{
			"success - vote proposal success",
			func() []interface{} {
				return []interface{}{
					s.address,
					proposalId,
					option,
					metadata,
				}
			},
			func() {
				proposal, _ := s.app.GovKeeper.GetProposal(s.ctx, proposalId)
				_, _, tallyResult := s.app.GovKeeper.Tally(s.ctx, proposal)
				s.Require().Equal(math.NewInt(3e18).String(), tallyResult.YesCount)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			_, err := s.precompile.Vote(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
