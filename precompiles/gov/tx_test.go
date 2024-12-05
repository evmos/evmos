package gov_test

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

func (s *PrecompileTestSuite) TestVote() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.VoteMethod]
	newVoterAddr := utiltx.GenerateAddress()
	const proposalID uint64 = 1
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
					proposalID,
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
					proposalID,
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
					proposalID,
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
			"fail - invalid vote option",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					option + 10,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid vote option",
		},
		{
			"success - vote proposal success",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					option,
					metadata,
				}
			},
			func() {
				proposal, _ := s.network.App.GovKeeper.Proposals.Get(ctx, proposalID)
				_, _, tallyResult, err := s.network.App.GovKeeper.Tally(ctx, proposal)
				s.Require().NoError(err)
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
			ctx = s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			_, err := s.precompile.Vote(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestVoteWeighted() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.VoteWeightedMethod]
	newVoterAddr := utiltx.GenerateAddress()
	const proposalID uint64 = 1
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
					proposalID,
					[]gov.WeightedVoteOption{},
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
					proposalID,
					[]gov.WeightedVoteOption{},
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"does not match the voter address",
		},
		{
			"fail - invalid vote option",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					[]gov.WeightedVoteOption{{Option: 10, Weight: "1.0"}},
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid vote option",
		},
		{
			"fail - invalid weight sum",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					[]gov.WeightedVoteOption{
						{Option: 1, Weight: "0.5"},
						{Option: 2, Weight: "0.6"},
					},
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"total weight overflow 1.00",
		},
		{
			"success - vote weighted proposal",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					[]gov.WeightedVoteOption{
						{Option: 1, Weight: "0.7"},
						{Option: 2, Weight: "0.3"},
					},
					metadata,
				}
			},
			func() {
				proposal, _ := s.network.App.GovKeeper.Proposals.Get(ctx, proposalID)
				_, _, tallyResult, err := s.network.App.GovKeeper.Tally(ctx, proposal)
				s.Require().NoError(err)
				s.Require().Equal("2100000000000000000", tallyResult.YesCount)
				s.Require().Equal("900000000000000000", tallyResult.AbstainCount)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			_, err := s.precompile.VoteWeighted(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestUpdateParams() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.UpdateParamsMethod]

	// Create some test parameters
	testParams := govv1.DefaultParams()

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		authorizer  func() common.Address
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - unauthorized address",
			func() []interface{} {
				return []interface{}{testParams}
			},
			func() common.Address {
				address := utiltx.GenerateAddress()
				return address
			},
			func() {},
			200000,
			true,
			"invalid authority",
		},
		{
			"success - valid params update",
			func() []interface{} {
				// Set the caller to be the authority
				return []interface{}{testParams}
			},
			func() common.Address {
				authority := s.network.App.GovKeeper.GetAuthority()

				bech32Prefix := strings.SplitN(authority, "1", 2)[0]

				addressBz, err := sdk.GetFromBech32(authority, bech32Prefix)
				s.Require().NoError(err)

				authorityAddress := common.BytesToAddress(addressBz)

				return authorityAddress
			},
			func() {
				params, err := s.network.App.GovKeeper.Params.Get(ctx)
				s.Require().NoError(err)
				s.Require().Equal(testParams.MinDeposit, params.MinDeposit)
				s.Require().Equal(testParams.VotingPeriod, params.VotingPeriod)
				s.Require().Equal(testParams.Quorum, params.Quorum)
				s.Require().Equal(testParams.Threshold, params.Threshold)
				s.Require().Equal(testParams.VetoThreshold, params.VetoThreshold)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, tc.authorizer(), s.precompile, tc.gas)

			_, err := s.precompile.UpdateParams(ctx, tc.authorizer(), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
