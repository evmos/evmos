package gov_test

import (
	"math/big"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	evmostypes "github.com/evmos/evmos/v20/types"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

var (
	_, _, addr = testdata.KeyTestPubAddr()
	// gov account authority address
	govAcct = authtypes.NewModuleAddress(govtypes.ModuleName)
	// TestProposalMsgs are msgs used on a proposal.
	TestProposalMsgs = []sdk.Msg{
		banktypes.NewMsgSend(govAcct, addr, sdk.NewCoins(sdk.NewCoin(evmostypes.BaseDenom, math.NewInt(1000)))),
	}
)

func (s *PrecompileTestSuite) TestGetVotes() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.GetVotesMethod]
	testCases := []struct {
		name     string
		malleate func() []gov.WeightedVote
		args     []interface{}
		expPass  bool
		expTotal uint64
		gas      uint64
	}{
		{
			name: "valid query",
			malleate: func() []gov.WeightedVote {
				err := s.network.App.GovKeeper.AddVote(s.network.GetContext(), 1, s.keyring.GetAccAddr(0), []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}, "")
				s.Require().NoError(err)
				return []gov.WeightedVote{
					{ProposalId: 1, Voter: s.keyring.GetAddr(0), Options: []gov.WeightedVoteOption{{Option: uint8(v1.OptionYes), Weight: "1.0"}}},
				}
			},
			args:     []interface{}{uint64(1), query.PageRequest{Limit: 10, CountTotal: true}},
			expPass:  true,
			expTotal: 1,
			gas:      200_000,
		},
		{
			name:    "invalid proposal ID",
			args:    []interface{}{uint64(0), query.PageRequest{Limit: 10, CountTotal: true}},
			expPass: false,
			gas:     200_000,
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
	var voter sdk.AccAddress
	method := s.precompile.Methods[gov.GetVoteMethod]
	testCases := []struct {
		name          string
		malleate      func()
		propNumber    uint64
		expPass       bool
		expPropNumber uint64
		expVoter      common.Address
		gas           uint64
		errContains   string
	}{
		{
			name: "valid query",
			malleate: func() {
				err := s.network.App.GovKeeper.AddVote(s.network.GetContext(), 1, voter, []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}, "")
				s.Require().NoError(err)
			},
			propNumber:    uint64(1),
			expPropNumber: uint64(1),
			expVoter:      common.BytesToAddress(voter.Bytes()),
			expPass:       true,
			gas:           200_000,
		},
		{
			name:       "invalid proposal ID",
			propNumber: uint64(10),
			expPass:    false,
			gas:        200_000,
			malleate: func() {
				err := s.network.App.GovKeeper.AddVote(s.network.GetContext(), 1, voter, []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}, "")
				s.Require().NoError(err)
			},
			errContains: "not found for proposal",
		},
		{
			name:        "non-existent vote",
			propNumber:  uint64(1),
			expPass:     false,
			gas:         200_000,
			malleate:    func() {},
			errContains: "not found for proposal",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			voter = s.keyring.GetAccAddr(0)

			tc.malleate()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			args := []interface{}{tc.propNumber, common.BytesToAddress(voter.Bytes())}
			bz, err := s.precompile.GetVote(ctx, &method, contract, args)

			expVote := gov.WeightedVote{
				ProposalId: tc.expPropNumber,
				Voter:      common.BytesToAddress(voter.Bytes()),
				Options:    []gov.WeightedVoteOption{{Option: uint8(v1.OptionYes), Weight: "1.0"}},
				Metadata:   "",
			}

			if tc.expPass {
				s.Require().NoError(err)
				var out gov.VoteOutput
				err = s.precompile.UnpackIntoInterface(&out, gov.GetVoteMethod, bz)

				s.Require().NoError(err)
				s.Require().Equal(expVote.ProposalId, out.Vote.ProposalId)
				s.Require().Equal(expVote.Voter, out.Vote.Voter)
				s.Require().Equal(expVote.Options, out.Vote.Options)
				s.Require().Equal(expVote.Metadata, out.Vote.Metadata)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGetDeposit() {
	var depositor sdk.AccAddress
	method := s.precompile.Methods[gov.GetDepositMethod]
	testCases := []struct {
		name          string
		malleate      func()
		propNumber    uint64
		expPass       bool
		expPropNumber uint64
		gas           uint64
		errContains   string
	}{
		{
			name:          "valid query",
			malleate:      func() {},
			propNumber:    uint64(1),
			expPropNumber: uint64(1),
			expPass:       true,
			gas:           200_000,
		},
		{
			name:        "invalid proposal ID",
			propNumber:  uint64(10),
			expPass:     false,
			gas:         200_000,
			malleate:    func() {},
			errContains: "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			depositor = s.keyring.GetAccAddr(0)

			tc.malleate()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			args := []interface{}{tc.propNumber, common.BytesToAddress(depositor.Bytes())}
			bz, err := s.precompile.GetDeposit(ctx, &method, contract, args)

			if tc.expPass {
				s.Require().NoError(err)
				var out gov.DepositOutput
				err = s.precompile.UnpackIntoInterface(&out, gov.GetDepositMethod, bz)

				s.Require().NoError(err)
				s.Require().Equal(tc.expPropNumber, out.Deposit.ProposalId)
				s.Require().Equal(common.BytesToAddress(depositor.Bytes()), out.Deposit.Depositor)
				s.Require().Equal([]cmn.Coin{{Denom: "aevmos", Amount: big.NewInt(100)}}, out.Deposit.Amount)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGetDeposits() {
	method := s.precompile.Methods[gov.GetDepositsMethod]
	testCases := []struct {
		name     string
		malleate func() []gov.DepositData
		args     []interface{}
		expPass  bool
		expTotal uint64
		gas      uint64
	}{
		{
			name: "valid query",
			malleate: func() []gov.DepositData {
				return []gov.DepositData{
					{ProposalId: 1, Depositor: s.keyring.GetAddr(0), Amount: []cmn.Coin{{Denom: s.network.GetDenom(), Amount: big.NewInt(100)}}},
				}
			},
			args:     []interface{}{uint64(1), query.PageRequest{Limit: 10, CountTotal: true}},
			expPass:  true,
			expTotal: 1,
			gas:      200_000,
		},
		{
			name:    "invalid proposal ID",
			args:    []interface{}{uint64(0), query.PageRequest{Limit: 10, CountTotal: true}},
			expPass: false,
			gas:     200_000,
			malleate: func() []gov.DepositData {
				return []gov.DepositData{}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx := s.network.GetContext()

			deposits := tc.malleate()
			contract, ctx := testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.GetDeposits(ctx, &method, contract, tc.args)
			if tc.expPass {
				var out gov.DepositsOutput
				err = s.precompile.UnpackIntoInterface(&out, gov.GetDepositsMethod, bz)
				s.Require().NoError(err)
				s.Require().Equal(deposits, out.Deposits)
				s.Require().Equal(tc.expTotal, out.PageResponse.Total)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGetTallyResult() {
	method := s.precompile.Methods[gov.GetTallyResultMethod]
	testCases := []struct {
		name        string
		malleate    func() gov.TallyResultData
		propNumber  uint64
		expPass     bool
		gas         uint64
		errContains string
	}{
		{
			name: "valid query",
			malleate: func() gov.TallyResultData {
				proposal, err := s.network.App.GovKeeper.SubmitProposal(s.network.GetContext(), TestProposalMsgs, "", "Proposal", "testing proposal", s.keyring.GetAccAddr(0), false)
				s.Require().NoError(err)
				votingStarted, err := s.network.App.GovKeeper.AddDeposit(s.network.GetContext(), proposal.Id, s.keyring.GetAccAddr(0), sdk.NewCoins(sdk.NewCoin(s.network.GetDenom(), math.NewInt(100))))
				s.Require().NoError(err)
				s.Require().True(votingStarted)
				err = s.network.App.GovKeeper.AddVote(s.network.GetContext(), proposal.Id, s.keyring.GetAccAddr(0), v1.NewNonSplitVoteOption(v1.OptionYes), "")
				s.Require().NoError(err)
				return gov.TallyResultData{
					Yes:        "3000000000000000000",
					Abstain:    "0",
					No:         "0",
					NoWithVeto: "0",
				}
			},
			propNumber: uint64(1),
			expPass:    true,
			gas:        200_000,
		},
		{
			name:        "invalid proposal ID",
			propNumber:  uint64(10),
			expPass:     false,
			gas:         200_000,
			malleate:    func() gov.TallyResultData { return gov.TallyResultData{} },
			errContains: "proposal 10 doesn't exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			expTally := tc.malleate()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			args := []interface{}{tc.propNumber}
			bz, err := s.precompile.GetTallyResult(ctx, &method, contract, args)

			if tc.expPass {
				s.Require().NoError(err)
				var out gov.TallyResultOutput
				err = s.precompile.UnpackIntoInterface(&out, gov.GetTallyResultMethod, bz)

				s.Require().NoError(err)
				s.Require().Equal(expTally, out.TallyResult)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}
