// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package gov_test

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/gov"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	"github.com/evmos/evmos/v18/precompiles/testutil/contracts"
	testutiltx "github.com/evmos/evmos/v18/testutil/tx"
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
	defaultCallArgs contracts.CallArgs

	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs
)

var _ = Describe("Calling distribution precompile from EOA", func() {
	BeforeEach(func() {
		s.SetupTest()

		// set the default call arguments
		defaultCallArgs = contracts.CallArgs{
			ContractAddr: s.precompile.Address(),
			ContractABI:  s.precompile.ABI,
			PrivKey:      s.privKey,
		}

		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.ABI.Events,
		}
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())
	})

	// =====================================
	// 				TRANSACTIONS
	// =====================================
	Describe("Execute Vote transaction", func() {
		const method = gov.VoteMethod
		const proposalId uint64 = 1
		const option uint8 = 1
		const metadata = "metadata"
		// defaultVoteArgs are the default arguments to set the withdraw address
		//
		// NOTE: this has to be populated in the BeforeEach block because the private key otherwise is not yet initialized.
		var defaultVoteArgs contracts.CallArgs

		BeforeEach(func() {
			// set the default call arguments
			defaultVoteArgs = defaultCallArgs.WithMethodName(method)
		})

		It("should return error if the provided gasLimit is too low", func() {
			voteArgs := defaultVoteArgs.
				WithGasLimit(30000).
				WithArgs(s.address, proposalId, option, metadata)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, voteArgs, outOfGasCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring("out of gas"), "expected out of gas error")

			// tally result yes count should remain unchanged
			proposal, _ := s.app.GovKeeper.GetProposal(s.ctx, proposalId)
			_, _, tallyResult := s.app.GovKeeper.Tally(s.ctx, proposal)
			Expect(tallyResult.YesCount).To(Equal("0"), "expected tally result yes count to remain unchanged")
		})

		It("should return error if the origin is different than the voter", func() {
			voteArgs := defaultVoteArgs.WithArgs(differentAddr, proposalId, option, metadata)

			voterSetCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.address.String(), differentAddr.String())

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, voteArgs, voterSetCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(gov.ErrDifferentOrigin, s.address, differentAddr)), "expected different origin error")
		})

		It("should vote success", func() {
			voteArgs := defaultVoteArgs.WithArgs(s.address, proposalId, option, metadata)

			voterSetCheck := passCheck.WithExpEvents(gov.EventTypeVote)

			_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, voteArgs, voterSetCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			// tally result yes count should updated
			proposal, _ := s.app.GovKeeper.GetProposal(s.ctx, proposalId)
			_, _, tallyResult := s.app.GovKeeper.Tally(s.ctx, proposal)
			Expect(tallyResult.YesCount).To(Equal(math.NewInt(3e18).String()), "expected tally result yes count updated")
		})
	})

	// =====================================
	// 				QUERIES
	// =====================================
	Describe("Execute queries", func() {
		It("should get proposal info - proposal query", func() {
		})
	})
})
