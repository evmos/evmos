// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package gov_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"

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
				txArgs.GasLimit = 200_000

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
						Limit:      10,
						CountTotal: true,
					},
				}
				txArgs.GasLimit = 200_000

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
	})
})
