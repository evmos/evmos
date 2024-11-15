// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package gov_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// General variables used for integration tests
var (
	// differentAddr is an address generated for testing purposes that e.g. raises the different origin error
	differentAddr = testutiltx.GenerateAddress()
	// defaultCallArgs  are the default arguments for calling the smart contract
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	callArgs factory.CallArgs
	// txArgs are the EVM transaction arguments to use in the transactions
	txArgs evmtypes.EvmTxArgs
	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
)

func TestKeeperIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

var _ = Describe("Calling governance precompile from EOA", func() {
	var s *PrecompileTestSuite
	const (
		proposalID uint64 = 1
		option     uint8  = 1
		metadata          = "metadata"
	)
	BeforeEach(func() {
		s = new(PrecompileTestSuite)
		s.SetupTest()

		// set the default call arguments
		callArgs = factory.CallArgs{
			ContractABI: s.precompile.ABI,
		}
		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.ABI.Events,
		}
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())

		// reset tx args each test to avoid keeping custom
		// values of previous tests (e.g. gasLimit)
		precompileAddr := s.precompile.Address()
		txArgs = evmtypes.EvmTxArgs{
			To: &precompileAddr,
		}
		txArgs.GasLimit = 200_000
	})

	// =====================================
	// 				TRANSACTIONS
	// =====================================
	Describe("Execute Vote transaction", func() {
		const method = gov.VoteMethod

		BeforeEach(func() {
			// set the default call arguments
			callArgs.MethodName = method
		})

		It("should return error if the provided gasLimit is too low", func() {
			txArgs.GasLimit = 30000
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0), proposalID, option, metadata,
			}

			_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, outOfGasCheck)
			Expect(err).To(BeNil())

			// tally result yes count should remain unchanged
			proposal, _ := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), proposalID)
			_, _, tallyResult, err := s.network.App.GovKeeper.Tally(s.network.GetContext(), proposal)
			Expect(err).To(BeNil())
			Expect(tallyResult.YesCount).To(Equal("0"), "expected tally result yes count to remain unchanged")
		})

		It("should return error if the origin is different than the voter", func() {
			callArgs.Args = []interface{}{
				differentAddr, proposalID, option, metadata,
			}

			voterSetCheck := defaultLogCheck.WithErrContains(gov.ErrDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, voterSetCheck)
			Expect(err).To(BeNil())
		})

		It("should vote success", func() {
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0), proposalID, option, metadata,
			}

			voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVote)

			_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, voterSetCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// tally result yes count should updated
			proposal, _ := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), proposalID)
			_, _, tallyResult, err := s.network.App.GovKeeper.Tally(s.network.GetContext(), proposal)
			Expect(err).To(BeNil())

			Expect(tallyResult.YesCount).To(Equal(math.NewInt(3e18).String()), "expected tally result yes count updated")
		})
	})

	Describe("Execute VoteWeighted transaction", func() {
		const method = gov.VoteWeightedMethod

		BeforeEach(func() {
			callArgs.MethodName = method
		})

		It("should return error if the provided gasLimit is too low", func() {
			txArgs.GasLimit = 30000
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				proposalID,
				[]gov.WeightedVoteOption{
					{Option: 1, Weight: "0.5"},
					{Option: 2, Weight: "0.5"},
				},
				metadata,
			}

			_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, outOfGasCheck)
			Expect(err).To(BeNil())

			// tally result should remain unchanged
			proposal, _ := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), proposalID)
			_, _, tallyResult, err := s.network.App.GovKeeper.Tally(s.network.GetContext(), proposal)
			Expect(err).To(BeNil())
			Expect(tallyResult.YesCount).To(Equal("0"), "expected tally result to remain unchanged")
		})

		It("should return error if the origin is different than the voter", func() {
			callArgs.Args = []interface{}{
				differentAddr,
				proposalID,
				[]gov.WeightedVoteOption{
					{Option: 1, Weight: "0.5"},
					{Option: 2, Weight: "0.5"},
				},
				metadata,
			}

			voterSetCheck := defaultLogCheck.WithErrContains(gov.ErrDifferentOrigin, s.keyring.GetAddr(0).String(), differentAddr.String())

			_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, voterSetCheck)
			Expect(err).To(BeNil())
		})

		It("should vote weighted success", func() {
			callArgs.Args = []interface{}{
				s.keyring.GetAddr(0),
				proposalID,
				[]gov.WeightedVoteOption{
					{Option: 1, Weight: "0.7"},
					{Option: 2, Weight: "0.3"},
				},
				metadata,
			}

			voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVoteWeighted)

			_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, callArgs, voterSetCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// tally result should be updated
			proposal, _ := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), proposalID)
			_, _, tallyResult, err := s.network.App.GovKeeper.Tally(s.network.GetContext(), proposal)
			Expect(err).To(BeNil())

			expectedYesCount := math.NewInt(21e17) // 70% of 3e18
			Expect(tallyResult.YesCount).To(Equal(expectedYesCount.String()), "expected tally result yes count updated")

			expectedAbstainCount := math.NewInt(9e17) // 30% of 3e18
			Expect(tallyResult.AbstainCount).To(Equal(expectedAbstainCount.String()), "expected tally result no count updated")
		})
	})

	// =====================================
	// 				QUERIES
	// =====================================
	Describe("Execute queries", func() {
		Context("vote query", func() {
			method := gov.GetVoteMethod
			BeforeEach(func() {
				// submit a vote
				voteArgs := factory.CallArgs{
					ContractABI: s.precompile.ABI,
					MethodName:  gov.VoteMethod,
					Args: []interface{}{
						s.keyring.GetAddr(0), proposalID, option, metadata,
					},
				}

				voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVote)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, voteArgs, voterSetCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")
				Expect(s.network.NextBlock()).To(BeNil())
			})
			It("should return a vote", func() {
				callArgs.MethodName = method
				callArgs.Args = []interface{}{proposalID, s.keyring.GetAddr(0)}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.VoteOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Vote.Voter).To(Equal(s.keyring.GetAddr(0)))
				Expect(out.Vote.ProposalId).To(Equal(proposalID))
				Expect(out.Vote.Metadata).To(Equal(metadata))
				Expect(out.Vote.Options).To(HaveLen(1))
				Expect(out.Vote.Options[0].Option).To(Equal(option))
				Expect(out.Vote.Options[0].Weight).To(Equal(math.LegacyOneDec().String()))
			})
		})

		Context("weighted vote query", func() {
			method := gov.GetVoteMethod
			BeforeEach(func() {
				// submit a weighted vote
				voteArgs := factory.CallArgs{
					ContractABI: s.precompile.ABI,
					MethodName:  gov.VoteWeightedMethod,
					Args: []interface{}{
						s.keyring.GetAddr(0),
						proposalID,
						[]gov.WeightedVoteOption{
							{Option: 1, Weight: "0.7"},
							{Option: 2, Weight: "0.3"},
						},
						metadata,
					},
				}

				voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVoteWeighted)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, voteArgs, voterSetCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")
				Expect(s.network.NextBlock()).To(BeNil())
			})

			It("should return a weighted vote", func() {
				callArgs.MethodName = method
				callArgs.Args = []interface{}{proposalID, s.keyring.GetAddr(0)}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.VoteOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Vote.Voter).To(Equal(s.keyring.GetAddr(0)))
				Expect(out.Vote.ProposalId).To(Equal(proposalID))
				Expect(out.Vote.Metadata).To(Equal(metadata))
				Expect(out.Vote.Options).To(HaveLen(2))
				Expect(out.Vote.Options[0].Option).To(Equal(uint8(1)))
				Expect(out.Vote.Options[0].Weight).To(Equal("0.7"))
				Expect(out.Vote.Options[1].Option).To(Equal(uint8(2)))
				Expect(out.Vote.Options[1].Weight).To(Equal("0.3"))
			})
		})

		Context("votes query", func() {
			method := gov.GetVotesMethod
			BeforeEach(func() {
				// submit votes
				for _, key := range s.keyring.GetKeys() {
					voteArgs := factory.CallArgs{
						ContractABI: s.precompile.ABI,
						MethodName:  gov.VoteMethod,
						Args: []interface{}{
							key.Addr, proposalID, option, metadata,
						},
					}

					voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVote)

					_, _, err := s.factory.CallContractAndCheckLogs(key.Priv, txArgs, voteArgs, voterSetCheck)
					Expect(err).To(BeNil(), "error while calling the precompile")
					Expect(s.network.NextBlock()).To(BeNil())
				}
			})
			It("should return all votes", func() {
				callArgs.MethodName = method
				callArgs.Args = []interface{}{
					proposalID,
					query.PageRequest{
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.VotesOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				votersCount := len(s.keyring.GetKeys())
				Expect(out.PageResponse.Total).To(Equal(uint64(votersCount)))
				Expect(out.PageResponse.NextKey).To(Equal([]byte{}))
				Expect(out.Votes).To(HaveLen(votersCount))
				for _, v := range out.Votes {
					Expect(v.ProposalId).To(Equal(proposalID))
					Expect(v.Metadata).To(Equal(metadata))
					Expect(v.Options).To(HaveLen(1))
					Expect(v.Options[0].Option).To(Equal(option))
					Expect(v.Options[0].Weight).To(Equal(math.LegacyOneDec().String()))
				}
			})
		})

		Context("deposit query", func() {
			method := gov.GetDepositMethod
			BeforeEach(func() {
				callArgs.MethodName = method
			})

			It("should return a deposit", func() {
				callArgs.Args = []interface{}{proposalID, s.keyring.GetAddr(0)}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.DepositOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Deposit.ProposalId).To(Equal(proposalID))
				Expect(out.Deposit.Depositor).To(Equal(s.keyring.GetAddr(0)))
				Expect(out.Deposit.Amount).To(HaveLen(1))
				Expect(out.Deposit.Amount[0].Denom).To(Equal(s.network.GetDenom()))
				Expect(out.Deposit.Amount[0].Amount.Cmp(big.NewInt(100))).To(Equal(0))
			})
		})

		Context("deposits query", func() {
			method := gov.GetDepositsMethod
			BeforeEach(func() {
				callArgs.MethodName = method
			})

			It("should return all deposits", func() {
				callArgs.Args = []interface{}{
					proposalID,
					query.PageRequest{
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.DepositsOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.PageResponse.Total).To(Equal(uint64(1)))
				Expect(out.PageResponse.NextKey).To(Equal([]byte{}))
				Expect(out.Deposits).To(HaveLen(1))
				for _, d := range out.Deposits {
					Expect(d.ProposalId).To(Equal(proposalID))
					Expect(d.Amount).To(HaveLen(1))
					Expect(d.Amount[0].Denom).To(Equal(s.network.GetDenom()))
					Expect(d.Amount[0].Amount.Cmp(big.NewInt(100))).To(Equal(0))
				}
			})
		})

		Context("tally result query", func() {
			method := gov.GetTallyResultMethod
			BeforeEach(func() {
				callArgs.MethodName = method
				voteArgs := factory.CallArgs{
					ContractABI: s.precompile.ABI,
					MethodName:  gov.VoteMethod,
					Args: []interface{}{
						s.keyring.GetAddr(0), proposalID, option, metadata,
					},
				}

				voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVote)

				_, _, err := s.factory.CallContractAndCheckLogs(s.keyring.GetPrivKey(0), txArgs, voteArgs, voterSetCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")
				Expect(s.network.NextBlock()).To(BeNil())
			})

			It("should return the tally result", func() {
				callArgs.Args = []interface{}{proposalID}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.TallyResultOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.TallyResult.Yes).To(Equal("3000000000000000000"))
				Expect(out.TallyResult.Abstain).To(Equal("0"))
				Expect(out.TallyResult.No).To(Equal("0"))
				Expect(out.TallyResult.NoWithVeto).To(Equal("0"))
			})
		})

		Context("proposal query", func() {
			method := gov.GetProposalMethod
			BeforeEach(func() {
				callArgs.MethodName = method
			})

			It("should return a proposal", func() {
				callArgs.Args = []interface{}{uint64(1)}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.ProposalOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				// Check proposal details
				Expect(out.Proposal.Id).To(Equal(uint64(1)))
				Expect(out.Proposal.Status).To(Equal(uint32(v1.StatusVotingPeriod)))
				Expect(out.Proposal.Proposer).To(Equal(s.keyring.GetAddr(0)))
				Expect(out.Proposal.Metadata).To(Equal("ipfs://CID"))
				Expect(out.Proposal.Title).To(Equal("test prop"))
				Expect(out.Proposal.Summary).To(Equal("test prop"))
				Expect(out.Proposal.Messages).To(HaveLen(1))
				Expect(out.Proposal.Messages[0]).To(Equal("/cosmos.bank.v1beta1.MsgSend"))

				// Check tally result
				Expect(out.Proposal.FinalTallyResult.Yes).To(Equal("0"))
				Expect(out.Proposal.FinalTallyResult.Abstain).To(Equal("0"))
				Expect(out.Proposal.FinalTallyResult.No).To(Equal("0"))
				Expect(out.Proposal.FinalTallyResult.NoWithVeto).To(Equal("0"))
			})

			It("should fail when proposal doesn't exist", func() {
				callArgs.Args = []interface{}{uint64(999)}

				_, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					defaultLogCheck.WithErrContains("proposal 999 doesn't exist"),
				)
				Expect(err).To(BeNil())
			})
		})

		Context("proposals query", func() {
			method := gov.GetProposalsMethod
			BeforeEach(func() {
				callArgs.MethodName = method
			})

			It("should return all proposals", func() {
				callArgs.Args = []interface{}{
					uint32(0), // StatusNil to get all proposals
					common.Address{},
					common.Address{},
					query.PageRequest{
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out gov.ProposalsOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Proposals).To(HaveLen(2))
				Expect(out.PageResponse.Total).To(Equal(uint64(2)))

				proposal := out.Proposals[0]
				Expect(proposal.Id).To(Equal(uint64(1)))
				Expect(proposal.Status).To(Equal(uint32(v1.StatusVotingPeriod)))
				Expect(proposal.Proposer).To(Equal(s.keyring.GetAddr(0)))
				Expect(proposal.Messages).To(HaveLen(1))
				Expect(proposal.Messages[0]).To(Equal("/cosmos.bank.v1beta1.MsgSend"))
			})

			It("should filter proposals by status", func() {
				callArgs.Args = []interface{}{
					uint32(v1.StatusVotingPeriod),
					common.Address{},
					common.Address{},
					query.PageRequest{
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil())

				var out gov.ProposalsOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Proposals).To(HaveLen(2))
				Expect(out.Proposals[0].Status).To(Equal(uint32(v1.StatusVotingPeriod)))
				Expect(out.Proposals[1].Status).To(Equal(uint32(v1.StatusVotingPeriod)))
			})

			It("should filter proposals by voter", func() {
				// First add a vote
				voteArgs := factory.CallArgs{
					ContractABI: s.precompile.ABI,
					MethodName:  gov.VoteMethod,
					Args: []interface{}{
						s.keyring.GetAddr(0), uint64(1), uint8(v1.OptionYes), "",
					},
				}
				_, _, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					voteArgs,
					passCheck.WithExpEvents(gov.EventTypeVote),
				)
				Expect(err).To(BeNil())

				// Wait for the vote to be included in the block
				Expect(s.network.NextBlock()).To(BeNil())

				// Query proposals filtered by voter
				callArgs.Args = []interface{}{
					uint32(0), // StatusNil
					s.keyring.GetAddr(0),
					common.Address{},
					query.PageRequest{
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil())

				var out gov.ProposalsOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Proposals).To(HaveLen(1))
			})

			It("should filter proposals by depositor", func() {
				callArgs.Args = []interface{}{
					uint32(0), // StatusNil
					common.Address{},
					s.keyring.GetAddr(0),
					query.PageRequest{
						CountTotal: true,
					},
				}

				_, ethRes, err := s.factory.CallContractAndCheckLogs(
					s.keyring.GetPrivKey(0),
					txArgs,
					callArgs,
					passCheck,
				)
				Expect(err).To(BeNil())

				var out gov.ProposalsOutput
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil())

				Expect(out.Proposals).To(HaveLen(1))
			})
		})
	})
})
