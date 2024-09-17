package gov_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

func (s *PrecompileTestSuite) TestGetVotes() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.GetVotesMethod]
	testCases := []struct {
		name     string
		args     []interface{}
		expPass  bool
		expTotal uint64
		gas      uint64
		malleate func() []gov.WeightedVote
	}{
		{
			name:     "valid query",
			args:     []interface{}{uint64(1), query.PageRequest{Limit: 10, CountTotal: true}},
			expPass:  true,
			expTotal: 1,
			gas:      200000,
			malleate: func() []gov.WeightedVote {
				err := s.network.App.GovKeeper.AddVote(s.network.GetContext(), 1, s.keyring.GetAddr(0).Bytes(), []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}, "")
				s.Require().NoError(err)
				return []gov.WeightedVote{
					{ProposalId: 1, Voter: s.keyring.GetAddr(0), Options: []gov.WeightedVoteOption{{Option: uint8(v1.OptionYes), Weight: "1.0"}}},
				}
			},
		},
		{
			name:    "invalid proposal ID",
			args:    []interface{}{uint64(0), query.PageRequest{Limit: 10, CountTotal: true}},
			expPass: false,
			gas:     200000,
			malleate: func() []gov.WeightedVote {
				return []gov.WeightedVote{}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			votes := tc.malleate()
			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.GetVotes(ctx, &method, contract, tc.args)

			if tc.expPass {
				var out gov.VotesOutput
				err = s.precompile.UnpackIntoInterface(&out, gov.GetVotesMethod, bz)
				s.Require().NoError(err)
				s.Require().Equal(votes, out.Votes)
				s.Require().Equal(tc.expTotal, out.PageResponse.Total)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGetVote() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.GetVoteMethod]
	testCases := []struct {
		name     string
		args     []interface{}
		expPass  bool
		expVote  gov.WeightedVote
		gas      uint64
		malleate func()
	}{
		{
			name:    "valid query",
			args:    []interface{}{uint64(1), s.keyring.GetAddr(0)},
			expPass: true,
			gas:     200000,
			malleate: func() {
				err := s.network.App.GovKeeper.AddVote(s.network.GetContext(), 1, sdk.AccAddress(s.keyring.GetAddr(0).Bytes()), []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}, "")
				s.Require().NoError(err)
			},
			expVote: gov.WeightedVote{
				ProposalId: 1,
				Voter:      s.keyring.GetAddr(0),
				Options:    []gov.WeightedVoteOption{{Option: uint8(v1.OptionYes), Weight: "1.0"}},
				Metadata:   "",
			},
		},
		{
			name:     "invalid proposal ID",
			args:     []interface{}{uint64(0), s.keyring.GetAddr(0)},
			expPass:  false,
			gas:      200000,
			malleate: func() {},
		},
		{
			name:    "non-existent vote",
			args:    []interface{}{uint64(1), s.keyring.GetAddr(1)},
			expPass: false,
			gas:     200000,
			malleate: func() {
				err := s.network.App.GovKeeper.AddVote(s.network.GetContext(), 1, s.keyring.GetAddr(0).Bytes(), []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}, "")
				s.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			tc.malleate()
			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.GetVote(ctx, &method, contract, tc.args)

			if tc.expPass {
				s.Require().NoError(err)
				var out gov.WeightedVote
				err = s.precompile.UnpackIntoInterface(&out, gov.GetVoteMethod, bz)
				s.Require().NoError(err)
				s.Require().Equal(tc.expVote.ProposalId, out.ProposalId)
				s.Require().Equal(tc.expVote.Voter, out.Voter)
				s.Require().Equal(tc.expVote.Options, out.Options)
				s.Require().Equal(tc.expVote.Metadata, out.Metadata)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
