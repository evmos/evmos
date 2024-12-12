package gov_test

import (
	"fmt"

	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

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

func (s *PrecompileTestSuite) TestDeposit() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.DepositMethod]
	newDepositorAddr := utiltx.GenerateAddress()
	const proposalID uint64 = 1

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
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		{
			"fail - invalid depositor address",
			func() []interface{} {
				return []interface{}{
					"",
					proposalID,
					cmn.Coin{},
				}
			},
			func() {},
			200000,
			true,
			"invalid depositor",
		},
		{
			"fail - using a different depositor address",
			func() []interface{} {
				return []interface{}{
					newDepositorAddr,
					proposalID,
					[]cmn.Coin{
						{
							Denom:  evmtypes.GetEVMCoinDenom(),
							Amount: math.NewInt(1e18).BigInt(),
						},
					},
				}
			},
			func() {},
			200000,
			true,
			"does not match the depositor address",
		},
		{
			"fail - invalid coin",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					"invalid_coin",
				}
			},
			func() {},
			200000,
			true,
			"error while unpacking args to Coins struct",
		},
		{
			"success - deposit to proposal",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					[]cmn.Coin{
						{
							Denom:  evmtypes.GetEVMCoinDenom(),
							Amount: math.NewInt(1e18).BigInt(),
						},
					},
				}
			},
			func() {
				deposits, err := s.network.App.GovKeeper.GetDeposits(ctx, proposalID)
				s.Require().NoError(err)
				s.Require().Len(deposits, 1, "expected exactly one deposit")

				// 100 is the initial deposit
				s.Require().Equal(math.NewInt(1e18).AddRaw(100).BigInt(), deposits[0].Amount[0].Amount.BigInt())
				s.Require().Equal(evmtypes.GetEVMCoinDenom(), deposits[0].Amount[0].Denom)
				s.Require().Equal(s.keyring.GetAccAddr(0).String(), deposits[0].Depositor)
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

			_, err := s.precompile.Deposit(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCancelProposal() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.CancelProposalMethod]
	newProposerAddr := utiltx.GenerateAddress()
	const proposalID uint64 = 1

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - non-existent proposal",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					uint64(999),
				}
			},
			func() {},
			200000,
			true,
			"not found: key",
		},
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			"fail - invalid proposer address",
			func() []interface{} {
				return []interface{}{
					"",
					proposalID,
				}
			},
			func() {},
			200000,
			true,
			"invalid proposer",
		},
		{
			"fail - using a different proposer address",
			func() []interface{} {
				return []interface{}{
					newProposerAddr,
					proposalID,
				}
			},
			func() {},
			200000,
			true,
			"does not match the proposer address",
		},
		{
			"success - cancel proposal",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
				}
			},
			func() {
				found, err := s.network.App.GovKeeper.Proposals.Has(s.network.GetContext(), proposalID)
				s.Require().NoError(err)
				s.Require().False(found)
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

			_, err := s.precompile.CancelProposal(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
