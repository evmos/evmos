// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ics20_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	evmoscontracts "github.com/evmos/evmos/v18/contracts"
	evmostesting "github.com/evmos/evmos/v18/ibc/testing"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/erc20"
	"github.com/evmos/evmos/v18/precompiles/ics20"
	"github.com/evmos/evmos/v18/precompiles/testutil"
	"github.com/evmos/evmos/v18/precompiles/testutil/contracts"
	evmosutil "github.com/evmos/evmos/v18/testutil"
	teststypes "github.com/evmos/evmos/v18/types/tests"
	"github.com/evmos/evmos/v18/utils"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v18/x/inflation/v1/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// General variables used for integration tests
var (
	// defaultCallArgs and defaultApproveArgs are the default arguments for calling the smart contract and to
	// call the approve method specifically.
	//
	// NOTE: this has to be populated in a BeforeEach block because the contractAddr would otherwise be a nil address.
	defaultCallArgs, defaultApproveArgs contracts.CallArgs

	// defaultLogCheck instantiates a log check arguments struct with the precompile ABI events populated.
	defaultLogCheck testutil.LogCheckArgs
	// passCheck defines the arguments to check if the precompile returns no error
	passCheck testutil.LogCheckArgs
	// outOfGasCheck defines the arguments to check if the precompile returns out of gas error
	outOfGasCheck testutil.LogCheckArgs

	// gasPrice defines a default gas price to be used in the testing suite
	gasPrice = big.NewInt(200_000)

	// array of allocations with only one allocation for 'aevmos' coin
	defaultSingleAlloc []cmn.ICS20Allocation

	// interchainSenderContract is the compiled contract calling the interchain functionality
	interchainSenderContract evmtypes.CompiledContract
)

var _ = Describe("IBCTransfer Precompile", func() {
	BeforeEach(func() {
		s.suiteIBCTesting = true
		s.SetupTest()
		s.setupAllocationsForTesting()

		var err error
		Expect(err).To(BeNil(), "error while loading the interchain sender contract: %v", err)

		// set the default call arguments
		defaultCallArgs = contracts.CallArgs{
			ContractAddr: s.precompile.Address(),
			ContractABI:  s.precompile.ABI,
			PrivKey:      s.privKey,
			GasPrice:     gasPrice,
		}
		defaultApproveArgs = defaultCallArgs.WithMethodName(authorization.ApproveMethod)

		defaultLogCheck = testutil.LogCheckArgs{
			ABIEvents: s.precompile.ABI.Events,
		}
		passCheck = defaultLogCheck.WithExpPass(true)
		outOfGasCheck = defaultLogCheck.WithErrContains(vm.ErrOutOfGas.Error())
	})

	Describe("Execute approve transaction", func() {
		BeforeEach(func() {
			// check no previous authorization exist
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(0), "expected no authorizations before tests")
			defaultSingleAlloc = []cmn.ICS20Allocation{
				{
					SourcePort:        ibctesting.TransferPort,
					SourceChannel:     s.transferPath.EndpointA.ChannelID,
					SpendLimit:        defaultCmnCoins,
					AllowedPacketData: []string{"memo"},
				},
			}
		})

		// TODO uncomment when enforcing grantee != origin
		// It("should return error if the origin is same as the spender", func() {
		// 	approveArgs := defaultApproveArgs.WithArgs(
		// 		s.address,
		// 		defaultSingleAlloc,
		// 	)

		// 	_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, approveArgs, differentOriginCheck)
		// 	Expect(err).To(BeNil(), "error while calling the precompile")

		// 	s.chainA.NextBlock()

		// 	// check no authorization exist
		// 	auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
		// 	Expect(err).To(BeNil(), "error while getting authorizations")
		// 	Expect(auths).To(HaveLen(0), "expected no authorization")
		// })

		It("should return error if the provided gasLimit is too low", func() {
			approveArgs := defaultApproveArgs.
				WithGasLimit(30000).
				WithArgs(
					s.differentAddr,
					defaultSingleAlloc,
				)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, approveArgs, outOfGasCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring(vm.ErrOutOfGas.Error()))

			s.chainA.NextBlock()

			// check no authorization exist
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil())
			Expect(auths).To(HaveLen(0))
		})

		It("should approve the corresponding allocation", func() {
			approveArgs := defaultApproveArgs.WithArgs(
				s.differentAddr,
				defaultSingleAlloc,
			)

			approvalCheck := passCheck.
				WithExpEvents(authorization.EventTypeIBCTransferAuthorization)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, approveArgs, approvalCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			s.chainA.NextBlock()

			// check GetAuthorizations is returning the record
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(1), "expected one authorization")
			Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
			transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
			Expect(transferAuthz.Allocations[0].SpendLimit).To(Equal(defaultCoins))
		})
	})

	Describe("Execute revoke transaction", func() {
		var defaultRevokeArgs contracts.CallArgs
		BeforeEach(func() {
			// create authorization
			s.setTransferApproval(defaultCallArgs, s.differentAddr, defaultSingleAlloc)
			defaultRevokeArgs = defaultCallArgs.WithMethodName(authorization.RevokeMethod)
		})

		It("should revoke authorization", func() {
			revokeArgs := defaultRevokeArgs.WithArgs(
				s.differentAddr,
			)
			revokeCheck := passCheck.
				WithExpEvents(authorization.EventTypeIBCTransferAuthorization)

			_, _, err := contracts.CallContractAndCheckLogs(
				s.chainA.GetContext(),
				s.app,
				revokeArgs,
				revokeCheck,
			)
			Expect(err).To(BeNil(), "error while calling the precompile")

			s.chainA.NextBlock()

			// check no authorization exist
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(0), "expected no authorization")
		})
	})

	Describe("Execute increase allowance transaction", func() {
		BeforeEach(func() {
			s.setTransferApproval(defaultCallArgs, s.differentAddr, defaultSingleAlloc)
		})

		// TODO uncomment when enforcing grantee != origin
		// this is a copy of a different test but for a different method
		// It("should return an error if the origin is same as the spender", func() {
		// 	increaseAllowanceArgs := defaultCallArgs.
		// 		WithMethodName(authorization.IncreaseAllowanceMethod).
		// 		WithArgs(
		// 			s.address,
		// 			s.transferPath.EndpointA.ChannelConfig.PortID,
		// 			s.transferPath.EndpointA.ChannelID,
		// 			utils.BaseDenom,
		// 			big.NewInt(1e18),
		// 		)

		// 	differentOriginCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.address, s.differentAddr)

		// 	_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, increaseAllowanceArgs, differentOriginCheck)
		// 	Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

		// 	// check no authorization exist
		// 	auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.address.Bytes(), s.address.Bytes())
		// 	Expect(err).To(BeNil(), "error while getting authorizations")
		// 	Expect(auths).To(BeNil())
		// })

		It("should return an error if the allocation denom is not present", func() { //nolint:dupl
			increaseAllowanceArgs := defaultCallArgs.
				WithMethodName(authorization.IncreaseAllowanceMethod).
				WithArgs(
					s.differentAddr,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					"urandom",
					big.NewInt(1e18),
				)

			noMatchingAllocation := defaultLogCheck.WithErrContains(
				ics20.ErrNoMatchingAllocation,
				s.transferPath.EndpointA.ChannelConfig.PortID,
				s.transferPath.EndpointA.ChannelID,
				"urandom",
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, increaseAllowanceArgs, noMatchingAllocation)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			Expect(err.Error()).To(ContainSubstring(ics20.ErrNoMatchingAllocation, "transfer", "channel-0", "urandom"))

			// check authorization didn't change
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(1), "expected one authorization")
			Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
			transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
			Expect(transferAuthz.Allocations[0].SpendLimit).To(Equal(defaultCoins))
		})

		It("should increase allowance by 1 EVMOS", func() {
			s.setTransferApproval(defaultCallArgs, s.differentAddr, defaultSingleAlloc)

			increaseAllowanceArgs := defaultCallArgs.
				WithMethodName(authorization.IncreaseAllowanceMethod).
				WithArgs(
					s.differentAddr,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				)

			allowanceCheck := passCheck.WithExpEvents(authorization.EventTypeIBCTransferAuthorization)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, increaseAllowanceArgs, allowanceCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			s.chainA.NextBlock()

			// check auth was updated
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(1), "expected one authorization")
			Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
			transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
			Expect(transferAuthz.Allocations[0].SpendLimit).To(Equal(defaultCoins.Add(sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)})))
		})
	})

	Describe("Execute decrease allowance transaction", func() {
		BeforeEach(func() {
			s.setTransferApproval(defaultCallArgs, s.differentAddr, defaultSingleAlloc)
		})

		It("should fail if decreased amount is more than the total spend limit left", func() {
			decreaseAllowance := defaultCallArgs.
				WithMethodName(authorization.DecreaseAllowanceMethod).
				WithArgs(
					s.differentAddr,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(2e18),
				)

			allowanceCheck := defaultLogCheck.WithErrContains("negative amount")

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, decreaseAllowance, allowanceCheck)
			Expect(err).To(HaveOccurred(), "error while calling the precompile")
			Expect(err.Error()).To(ContainSubstring("negative amount"))
		})

		// TODO uncomment when enforcing grantee != origin
		// //nolint:dupl // this is a copy of a different test but for a different method
		// It("should return an error if the origin same the spender", func() {
		// 	decreaseAllowance := defaultCallArgs.
		// 		WithMethodName(authorization.DecreaseAllowanceMethod).
		// 		WithArgs(
		// 			s.address,
		// 			s.transferPath.EndpointA.ChannelConfig.PortID,
		// 			s.transferPath.EndpointA.ChannelID,
		// 			utils.BaseDenom,
		// 			big.NewInt(1e18),
		// 		)

		// 	differentOriginCheck := defaultLogCheck.WithErrContains(cmn.ErrDifferentOrigin, s.address, s.differentAddr)

		// 	_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, decreaseAllowance, differentOriginCheck)
		// 	Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

		// 	// check authorization does not exist
		// 	auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.address.Bytes(), s.address.Bytes())
		// 	Expect(err).To(BeNil(), "error while getting authorizations")
		// 	Expect(auths).To(BeNil())
		// })

		It("should return an error if the allocation denom is not present", func() { //nolint:dupl
			decreaseAllowance := defaultCallArgs.
				WithMethodName(authorization.DecreaseAllowanceMethod).
				WithArgs(
					s.differentAddr,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					"urandom",
					big.NewInt(1e18),
				)

			noMatchingAllocation := defaultLogCheck.WithErrContains(
				ics20.ErrNoMatchingAllocation,
				s.transferPath.EndpointA.ChannelConfig.PortID,
				s.transferPath.EndpointA.ChannelID,
				"urandom",
			)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, decreaseAllowance, noMatchingAllocation)
			Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
			Expect(err.Error()).To(ContainSubstring(ics20.ErrNoMatchingAllocation, "transfer", "channel-0", "urandom"))

			// check authorization didn't change
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(1), "expected one authorization")
			Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
			transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
			Expect(transferAuthz.Allocations[0].SpendLimit).To(Equal(defaultCoins))
		})

		It("should delete grant if allowance is decreased to 0", func() {
			decreaseAllowance := defaultCallArgs.
				WithMethodName(authorization.DecreaseAllowanceMethod).
				WithArgs(
					s.differentAddr,
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					utils.BaseDenom,
					big.NewInt(1e18),
				)

			allowanceCheck := passCheck.WithExpEvents(authorization.EventTypeIBCTransferAuthorization)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, decreaseAllowance, allowanceCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			s.chainA.NextBlock()

			// check auth record
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), s.differentAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(HaveLen(1), "expected one authorization")
			Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
			transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
			Expect(transferAuthz.Allocations[0].SpendLimit).To(HaveLen(0))
		})
	})

	Describe("Execute transfer transaction", func() {
		var defaultTransferArgs contracts.CallArgs

		BeforeEach(func() {
			// populate the default transfer args
			defaultTransferArgs = defaultCallArgs.
				WithMethodName(ics20.TransferMethod).
				WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					s.bondDenom,
					defaultCmnCoins[0].Amount,
					s.address,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)
		})

		Context("without authorization", func() {
			It("owner should transfer without authorization", func() {
				initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

				logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

				res, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				// check the sender balance was deducted
				fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
				finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
				Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees).Sub(defaultCoins[0].Amount)))
			})

			It("should succeed in transfer transaction but should timeout and refund sender", func() {
				initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

				logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)
				timeoutHeight := clienttypes.NewHeight(clienttypes.ParseChainID(s.chainB.ChainID), uint64(s.chainB.GetContext().BlockHeight())+1)

				transferArgs := defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					s.bondDenom,
					defaultCmnCoins[0].Amount,
					s.address,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					timeoutHeight,
					uint64(0), // disable timeout timestamp
					"memo",
				)

				res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking response: %v", err)
				// check sequence in returned data
				sequence, ok := out[0].(uint64)
				Expect(ok).To(BeTrue())
				Expect(sequence).To(Equal(uint64(1)))

				// check the sender balance was deducted
				fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
				finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
				Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees).Sub(defaultCoins[0].Amount)))

				// the transfer is reverted because the packet times out
				// build the sent packet
				// this is the packet sent
				packet := s.makePacket(
					sdk.AccAddress(s.address.Bytes()).String(),
					s.chainB.SenderAccount.GetAddress().String(),
					s.bondDenom,
					"memo",
					defaultCmnCoins[0].Amount,
					sequence,
					timeoutHeight,
				)

				// packet times out and the OnTimeoutPacket callback is executed
				s.chainA.NextBlock()
				// increment block height on chainB to make the packet timeout
				s.chainB.NextBlock()

				// increment sequence for successful transaction execution
				err = s.chainA.SenderAccount.SetSequence(s.chainA.SenderAccount.GetSequence() + 1)
				s.Require().NoError(err)

				err = s.transferPath.EndpointA.UpdateClient()
				Expect(err).To(BeNil())

				// Receive timeout
				err = s.transferPath.EndpointA.TimeoutPacket(packet)
				Expect(err).To(BeNil())

				// To submit a timeoutMsg, the TimeoutPacket function
				// uses a default fee amount
				timeoutMsgFee := math.NewInt(evmostesting.DefaultFeeAmt * 2)
				fees = fees.Add(timeoutMsgFee)

				finalBalance = s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
				Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))
			})

			It("should not transfer other account's balance", func() {
				// initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

				// fund senders account
				err := evmosutil.FundAccountWithBaseDenom(s.chainA.GetContext(), s.app.BankKeeper, s.differentAddr.Bytes(), amt)
				Expect(err).To(BeNil())
				senderInitialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.differentAddr.Bytes(), s.bondDenom)
				Expect(senderInitialBalance.Amount).To(Equal(math.NewInt(amt)))

				transferArgs := defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					s.bondDenom,
					defaultCmnCoins[0].Amount,
					s.differentAddr,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)

				logCheckArgs := defaultLogCheck.WithErrContains(ics20.ErrDifferentOriginFromSender, s.address, s.differentAddr)

				_, _, err = contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring(ics20.ErrDifferentOriginFromSender, s.address, s.differentAddr))

				// check the sender only paid for the fees
				// and funds were not transferred
				// TODO: fees are not calculated correctly with this logic
				// fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
				// finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
				// Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

				senderFinalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.differentAddr.Bytes(), s.bondDenom)
				Expect(senderFinalBalance.Amount).To(Equal(senderInitialBalance.Amount))
			})
		})

		Context("with authorization", func() {
			BeforeEach(func() {
				expTime := s.chainA.GetContext().BlockTime().Add(s.precompile.ApprovalExpiration)

				allocations := []transfertypes.Allocation{
					{
						SourcePort:    s.transferPath.EndpointA.ChannelConfig.PortID,
						SourceChannel: s.transferPath.EndpointA.ChannelID,
						SpendLimit:    defaultCoins,
					},
				}

				// create grant to allow s.address to spend differentAddr funds
				err := s.app.AuthzKeeper.SaveGrant(
					s.chainA.GetContext(),
					s.address.Bytes(),
					s.differentAddr.Bytes(),
					&transfertypes.TransferAuthorization{Allocations: allocations},
					&expTime,
				)
				Expect(err).To(BeNil())

				// fund the account from which funds will be sent
				err = evmosutil.FundAccountWithBaseDenom(s.chainA.GetContext(), s.app.BankKeeper, s.differentAddr.Bytes(), amt)
				Expect(err).To(BeNil())
				senderInitialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.differentAddr.Bytes(), s.bondDenom)
				Expect(senderInitialBalance.Amount).To(Equal(math.NewInt(amt)))
			})

			It("should not transfer other account's balance", func() {
				// ATM it is not allowed for another EOA to spend other EOA
				// funds via EVM extensions.
				// However, it is allowed for a contract to spend an EOA's account and
				// an EOA account to spend a contract's balance
				// if the required authorization exist
				// initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

				transferArgs := defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					s.bondDenom,
					defaultCmnCoins[0].Amount,
					s.differentAddr,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)

				logCheckArgs := defaultLogCheck.WithErrContains(ics20.ErrDifferentOriginFromSender, s.address, s.differentAddr)

				_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
				Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
				Expect(err.Error()).To(ContainSubstring(ics20.ErrDifferentOriginFromSender, s.address, s.differentAddr))

				// check the sender only paid for the fees
				// and funds from the other account were not transferred
				// TODO: fees are not calculated correctly with this logic
				// fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
				// finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
				// Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

				senderFinalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.differentAddr.Bytes(), s.bondDenom)
				Expect(senderFinalBalance.Amount).To(Equal(math.NewInt(amt)))
			})
		})

		Context("sending ERC20 coins", func() {
			var (
				// erc20Addr is the address of the ERC20 contract
				erc20Addr common.Address
				// sentAmount is the amount of tokens to send for testing
				sentAmount               = big.NewInt(1000)
				tokenPair                *erc20types.TokenPair
				defaultErc20TransferArgs contracts.CallArgs
				err                      error
			)

			BeforeEach(func() {
				erc20Addr = s.setupERC20ContractTests(sentAmount)
				// register the token pair
				tokenPair, err = s.app.Erc20Keeper.RegisterERC20(s.chainA.GetContext(), erc20Addr)
				Expect(err).To(BeNil(), "error while registering the token pair: %v", err)

				defaultErc20TransferArgs = defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					tokenPair.Denom,
					sentAmount,
					s.address,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)
			})

			Context("without authorization", func() {
				It("should transfer registered ERC20s", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultErc20TransferArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					Expect(out[0]).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// check only fees were deducted from sending account
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(preBalance.Amount.Sub(fees)))

					// check Erc20 balance was reduced by sent amount
					balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(balance.Int64()).To(BeZero(), "address does not have the expected amount of tokens")
				})

				It("should not transfer other account's balance", func() {
					// mint some ERC20 to the sender's account
					defaultERC20CallArgs := contracts.CallArgs{
						ContractAddr: erc20Addr,
						ContractABI:  evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						PrivKey:      s.privKey,
						GasPrice:     gasPrice,
					}

					// mint coins to the address
					mintCoinsArgs := defaultERC20CallArgs.
						WithMethodName("mint").
						WithArgs(s.differentAddr, defaultCmnCoins[0].Amount)

					mintCheck := testutil.LogCheckArgs{
						ABIEvents: evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI.Events,
						ExpEvents: []string{erc20.EventTypeTransfer}, // upon minting the tokens are sent to the receiving address
						ExpPass:   true,
					}

					_, _, err = contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, mintCoinsArgs, mintCheck)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					// try to transfer other account's erc20 tokens
					transferArgs := defaultTransferArgs.WithArgs(
						s.transferPath.EndpointA.ChannelConfig.PortID,
						s.transferPath.EndpointA.ChannelID,
						tokenPair.Denom,
						defaultCmnCoins[0].Amount,
						s.differentAddr,
						s.chainB.SenderAccount.GetAddress().String(), // receiver
						s.chainB.GetTimeoutHeight(),
						uint64(0), // disable timeout timestamp
						"memo",
					)

					logCheckArgs := defaultLogCheck.WithErrContains(ics20.ErrDifferentOriginFromSender, s.address, s.differentAddr)

					_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
					Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)
					Expect(err.Error()).To(ContainSubstring(ics20.ErrDifferentOriginFromSender, s.address, s.differentAddr))

					// check funds were not transferred
					balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.differentAddr,
					)
					Expect(balance).To(Equal(defaultCmnCoins[0].Amount), "address does not have the expected amount of tokens")
				})

				It("should succeed in transfer transaction but should error on packet destination if the receiver address is wrong", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					invalidReceiverAddr := "invalid_address"
					transferArgs := defaultTransferArgs.WithArgs(
						s.transferPath.EndpointA.ChannelConfig.PortID,
						s.transferPath.EndpointA.ChannelID,
						tokenPair.Denom,
						sentAmount,
						s.address,
						invalidReceiverAddr, // invalid receiver
						s.chainB.GetTimeoutHeight(),
						uint64(0), // disable timeout timestamp
						"memo",
					)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					sequence, ok := out[0].(uint64)
					Expect(ok).To(BeTrue())
					Expect(sequence).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// check only fees were deducted from sending account
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(preBalance.Amount.Sub(fees)))

					// check Erc20 balance was reduced by sent amount (escrowed on ibc escrow account)
					balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(balance.Int64()).To(BeZero(), "address does not have the expected amount of tokens")

					// the transfer is reverted because fails checks on the receiving chain
					// this is the packet sent
					packet := s.makePacket(
						sdk.AccAddress(s.address.Bytes()).String(),
						invalidReceiverAddr,
						tokenPair.Denom,
						"memo",
						sentAmount,
						sequence,
						s.chainB.GetTimeoutHeight(),
					)

					// increment sequence for successful transaction execution
					err = s.chainA.SenderAccount.SetSequence(s.chainA.SenderAccount.GetSequence() + 3)
					s.Require().NoError(err)

					err = s.transferPath.EndpointA.UpdateClient()
					Expect(err).To(BeNil())

					// Relay packet
					err = s.transferPath.RelayPacket(packet)
					Expect(err).To(BeNil())

					// check escrowed funds are refunded to sender
					finalERC20balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(finalERC20balance).To(Equal(sentAmount), "address does not have the expected amount of tokens")
				})

				It("should succeed in transfer transaction but should timeout", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					timeoutHeight := clienttypes.NewHeight(clienttypes.ParseChainID(s.chainB.ChainID), uint64(s.chainB.GetContext().BlockHeight())+1)

					transferArgs := defaultTransferArgs.WithArgs(
						s.transferPath.EndpointA.ChannelConfig.PortID,
						s.transferPath.EndpointA.ChannelID,
						tokenPair.Denom,
						sentAmount,
						s.address,
						s.chainB.SenderAccount.GetAddress().String(), // receiver
						timeoutHeight,
						uint64(0), // disable timeout timestamp
						"memo",
					)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					sequence, ok := out[0].(uint64)
					Expect(ok).To(BeTrue())
					Expect(sequence).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// check only fees were deducted from sending account
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(preBalance.Amount.Sub(fees)))

					// check Erc20 balance was reduced by sent amount
					balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(balance.Int64()).To(BeZero(), "address does not have the expected amount of tokens")

					// the transfer is reverted because the packet times out
					// this is the packet sent
					packet := s.makePacket(
						sdk.AccAddress(s.address.Bytes()).String(),
						s.chainB.SenderAccount.GetAddress().String(),
						tokenPair.Denom,
						"memo",
						sentAmount,
						sequence,
						timeoutHeight,
					)

					// packet times out and the OnTimeoutPacket callback is executed
					s.chainA.NextBlock()
					// increment block height on chainB to make the packet timeout
					s.chainB.NextBlock()

					// increment sequence for successful transaction execution
					err = s.chainA.SenderAccount.SetSequence(s.chainA.SenderAccount.GetSequence() + 3)
					s.Require().NoError(err)

					err = s.transferPath.EndpointA.UpdateClient()
					Expect(err).To(BeNil())

					// Receive timeout
					err = s.transferPath.EndpointA.TimeoutPacket(packet)
					Expect(err).To(BeNil())

					// check escrowed funds are refunded to sender
					finalERC20balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(finalERC20balance).To(Equal(sentAmount), "address does not have the expected amount of tokens")
				})
			})
		})
	})

	Context("Queries", func() {
		var (
			path     string
			expTrace transfertypes.DenomTrace
		)

		BeforeEach(func() {
			path = fmt.Sprintf(
				"%s/%s/%s/%s",
				s.transferPath.EndpointA.ChannelConfig.PortID,
				s.transferPath.EndpointA.ChannelID,
				s.transferPath.EndpointB.ChannelConfig.PortID,
				s.transferPath.EndpointB.ChannelID,
			)
			expTrace = transfertypes.DenomTrace{
				Path:      path,
				BaseDenom: utils.BaseDenom,
			}
		})

		It("should query denom trace", func() {
			// setup - create a denom trace to get it on the query result
			method := ics20.DenomTraceMethod
			s.app.TransferKeeper.SetDenomTrace(s.chainA.GetContext(), expTrace)

			args := defaultCallArgs.
				WithMethodName(method).
				WithArgs(expTrace.IBCDenom())

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			var out ics20.DenomTraceResponse
			err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
			Expect(out.DenomTrace.Path).To(Equal(expTrace.Path))
			Expect(out.DenomTrace.BaseDenom).To(Equal(expTrace.BaseDenom))
		})

		Context("denom traces query", func() {
			var (
				method    string
				expTraces []transfertypes.DenomTrace
			)
			BeforeEach(func() {
				method = ics20.DenomTracesMethod
				// setup - create some denom traces to get on the query result
				expTraces = []transfertypes.DenomTrace{
					{Path: "", BaseDenom: utils.BaseDenom},
					{Path: fmt.Sprintf("%s/%s", s.transferPath.EndpointA.ChannelConfig.PortID, s.transferPath.EndpointA.ChannelID), BaseDenom: utils.BaseDenom},
					expTrace,
				}

				for _, trace := range expTraces {
					s.app.TransferKeeper.SetDenomTrace(s.chainA.GetContext(), trace)
				}
			})
			It("should query denom traces - w/all results on page", func() {
				args := defaultCallArgs.
					WithMethodName(method).
					WithArgs(
						query.PageRequest{
							Limit:      3,
							CountTotal: true,
						})
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out ics20.DenomTracesResponse
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
				Expect(out.DenomTraces).To(HaveLen(3), "expected 3 denom traces to be returned")
				Expect(out.PageResponse.Total).To(Equal(uint64(3)))
				Expect(out.PageResponse.NextKey).To(BeEmpty())

				for i, dt := range out.DenomTraces {
					// order can change
					Expect(dt.Path).To(Equal(expTraces[i].Path))
					Expect(dt.BaseDenom).To(Equal(expTraces[i].BaseDenom))
				}
			})

			It("should query denom traces - w/pagination", func() {
				args := defaultCallArgs.
					WithMethodName(method).
					WithArgs(
						query.PageRequest{
							Limit:      1,
							CountTotal: true,
						})
				_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out ics20.DenomTracesResponse
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
				Expect(out.DenomTraces).To(HaveLen(1), "expected 1 denom traces to be returned")
				Expect(out.PageResponse.Total).To(Equal(uint64(3)))
				Expect(out.PageResponse.NextKey).NotTo(BeEmpty())
			})
		})

		It("should query denom hash", func() {
			method := ics20.DenomHashMethod
			// setup - create a denom expTrace
			s.app.TransferKeeper.SetDenomTrace(s.chainA.GetContext(), expTrace)

			args := defaultCallArgs.
				WithMethodName(method).
				WithArgs(expTrace.GetFullDenomPath())

			_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, passCheck)
			Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

			out, err := s.precompile.Unpack(method, ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
			Expect(out).To(HaveLen(1))
			Expect(out[0]).To(Equal(expTrace.Hash().String()))
		})
	})

	Context("query allowance", func() {
		Context("No authorization", func() {
			It("should return empty array", func() {
				method := authorization.AllowanceMethod

				args := defaultCallArgs.
					WithMethodName(method).
					WithArgs(s.address, s.differentAddr)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out []cmn.ICS20Allocation
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
				Expect(out).To(HaveLen(0))
			})
		})

		Context("with authorization", func() {
			BeforeEach(func() {
				s.setTransferApproval(defaultCallArgs, s.differentAddr, defaultSingleAlloc)
			})

			It("should return the allowance", func() {
				method := authorization.AllowanceMethod

				args := defaultCallArgs.
					WithMethodName(method).
					WithArgs(s.differentAddr, s.address)

				_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, passCheck)
				Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

				var out []cmn.ICS20Allocation
				err = s.precompile.UnpackIntoInterface(&out, method, ethRes.Ret)
				Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
				Expect(out).To(HaveLen(1))
				Expect(len(out)).To(Equal(len(defaultSingleAlloc)))
				Expect(out[0].SourcePort).To(Equal(defaultSingleAlloc[0].SourcePort))
				Expect(out[0].SourceChannel).To(Equal(defaultSingleAlloc[0].SourceChannel))
				Expect(out[0].SpendLimit).To(Equal(defaultSingleAlloc[0].SpendLimit))
				Expect(out[0].AllowList).To(HaveLen(0))
				Expect(out[0].AllowedPacketData).To(HaveLen(1))
				Expect(out[0].AllowedPacketData[0]).To(Equal("memo"))
			})
		})
	})
})

var _ = Describe("Calling ICS20 precompile from another contract", func() {
	var (

		// interchainSenderCallerContract is the compiled contract calling the interchain functionality
		interchainSenderCallerContract evmtypes.CompiledContract
		// contractAddr is the address of the smart contract that will be deployed
		contractAddr common.Address
		// senderCallerContractAddr is the address of the InterchainSenderCaller smart contract that will be deployed
		senderCallerContractAddr common.Address
		// execRevertedCheck defines the default log checking arguments which includes the
		// standard revert message.
		execRevertedCheck testutil.LogCheckArgs
		// err is a basic error type
		err error
	)

	BeforeEach(func() {
		s.suiteIBCTesting = true
		s.SetupTest()
		s.setupAllocationsForTesting()

		// Deploy InterchainSender contract
		interchainSenderContract, err = contracts.LoadInterchainSenderContract()
		Expect(err).To(BeNil(), "error while loading the interchain sender contract: %v", err)

		contractAddr, err = DeployContract(
			s.chainA.GetContext(),
			s.app,
			s.privKey,
			gasPrice,
			s.queryClientEVM,
			interchainSenderContract,
		)
		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)

		// NextBlock the smart contract
		s.chainA.NextBlock()

		// Deploy InterchainSenderCaller contract
		interchainSenderCallerContract, err = contracts.LoadInterchainSenderCallerContract()
		Expect(err).To(BeNil(), "error while loading the interchain sender contract: %v", err)

		senderCallerContractAddr, err = DeployContract(
			s.chainA.GetContext(),
			s.app,
			s.privKey,
			gasPrice,
			s.queryClientEVM,
			interchainSenderCallerContract,
			contractAddr,
		)
		Expect(err).To(BeNil(), "error while deploying the smart contract: %v", err)

		// NextBlock the smart contract
		s.chainA.NextBlock()

		// check contracts were correctly deployed
		cAcc := s.app.EvmKeeper.GetAccount(s.chainA.GetContext(), contractAddr)
		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

		cAcc = s.app.EvmKeeper.GetAccount(s.chainA.GetContext(), senderCallerContractAddr)
		Expect(cAcc).ToNot(BeNil(), "contract account should exist")
		Expect(cAcc.IsContract()).To(BeTrue(), "account should be a contract")

		// populate default call args
		defaultCallArgs = contracts.CallArgs{
			ContractAddr: contractAddr,
			ContractABI:  interchainSenderContract.ABI,
			PrivKey:      s.privKey,
			GasPrice:     gasPrice,
		}
		defaultApproveArgs = defaultCallArgs.
			WithMethodName("testApprove").
			WithArgs(defaultSingleAlloc)

		// default log check arguments
		defaultLogCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		execRevertedCheck = defaultLogCheck.WithErrContains("execution reverted")
		passCheck = defaultLogCheck.WithExpPass(true)
	})

	Context("approving methods", func() {
		Context("with valid input", func() {
			It("should approve one allocation", func() {
				approvalCheck := passCheck.
					WithExpEvents(authorization.EventTypeIBCTransferAuthorization)

				_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultApproveArgs, approvalCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")

				s.chainA.NextBlock()

				// check GetAuthorizations is returning the record
				auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes())
				Expect(err).To(BeNil(), "error while getting authorizations")
				Expect(auths).To(HaveLen(1), "expected one authorization")
				Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
				transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
				Expect(transferAuthz.Allocations[0].SpendLimit).To(Equal(defaultCoins))
			})
		})
	})

	Context("revoke method", func() {
		var defaultRevokeArgs contracts.CallArgs
		BeforeEach(func() {
			s.setTransferApprovalForContract(defaultApproveArgs)
			defaultRevokeArgs = defaultCallArgs.WithMethodName(
				"testRevoke",
			)
		})

		It("should revoke authorization", func() {
			// used to check if the corresponding event is emitted
			revokeCheck := passCheck.
				WithExpEvents(authorization.EventTypeIBCTransferAuthorization)

			_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultRevokeArgs, revokeCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			s.chainA.NextBlock()

			// check authorization was removed
			auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes())
			Expect(err).To(BeNil(), "error while getting authorizations")
			Expect(auths).To(BeNil())
		})
	})

	Context("update allowance methods", func() {
		var (
			amt                       *big.Int
			allowanceChangeCheck      testutil.LogCheckArgs
			defaultChangeAllowanceArg contracts.CallArgs
		)

		BeforeEach(func() {
			amt = big.NewInt(1e10)
			allowanceChangeCheck = passCheck.
				WithExpEvents(authorization.EventTypeIBCTransferAuthorization)
			s.setTransferApprovalForContract(defaultApproveArgs)
			defaultChangeAllowanceArg = defaultCallArgs.
				WithMethodName("testIncreaseAllowance").
				WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					utils.BaseDenom,
					amt,
				)
		})

		Context("Increase allowance", func() {
			It("should increase allowance", func() { //nolint:dupl
				_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultChangeAllowanceArg, allowanceChangeCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")

				s.chainA.NextBlock()

				// check authorization spend limit increased
				auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes())
				Expect(err).To(BeNil(), "error while getting authorizations")
				Expect(auths).To(HaveLen(1), "expected one authorization")
				Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
				transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
				Expect(transferAuthz.Allocations[0].SpendLimit.AmountOf(utils.BaseDenom)).To(Equal(defaultCoins.AmountOf(utils.BaseDenom).Add(math.NewIntFromBigInt(amt))))
			})
		})

		Context("Decrease allowance", func() {
			var defaultDecreaseAllowanceArg contracts.CallArgs

			BeforeEach(func() {
				defaultDecreaseAllowanceArg = defaultChangeAllowanceArg.
					WithMethodName("testDecreaseAllowance")
			})

			It("should decrease allowance", func() { //nolint:dupl
				_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultDecreaseAllowanceArg, allowanceChangeCheck)
				Expect(err).To(BeNil(), "error while calling the precompile")

				s.chainA.NextBlock()

				// check authorization spend limit decreased
				auths, err := s.app.AuthzKeeper.GetAuthorizations(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes())
				Expect(err).To(BeNil(), "error while getting authorizations")
				Expect(auths).To(HaveLen(1), "expected one authorization")
				Expect(auths[0].MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
				transferAuthz := auths[0].(*transfertypes.TransferAuthorization)
				Expect(transferAuthz.Allocations[0].SpendLimit.AmountOf(utils.BaseDenom)).To(Equal(defaultCoins.AmountOf(utils.BaseDenom).Sub(math.NewIntFromBigInt(amt))))
			})
		})
	})

	Context("transfer method", func() {
		var defaultTransferArgs contracts.CallArgs
		BeforeEach(func() {
			defaultTransferArgs = defaultCallArgs.WithMethodName(
				"testTransferUserFunds",
			)
		})

		Context("'aevmos' coin", func() {
			Context("with authorization", func() {
				BeforeEach(func() {
					// set approval to transfer 'aevmos'
					s.setTransferApprovalForContract(defaultApproveArgs)
				})

				It("should transfer funds", func() {
					initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					transferArgs := defaultTransferArgs.WithArgs(
						s.transferPath.EndpointA.ChannelConfig.PortID,
						s.transferPath.EndpointA.ChannelID,
						s.bondDenom,
						defaultCmnCoins[0].Amount,
						s.chainB.SenderAccount.GetAddress().String(), // receiver
						s.chainB.GetTimeoutHeight(),
						uint64(0), // disable timeout timestamp
						"memo",
					)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, transferArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					Expect(out[0]).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// The allowance is spent after the transfer thus the authorization is deleted
					authz, _ := s.app.AuthzKeeper.GetAuthorization(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
					Expect(authz).To(BeNil())

					// check sent tokens were deducted from sending account
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(defaultCoins.AmountOf(s.bondDenom)).Sub(fees)))
				})

				Context("Calling the InterchainSender caller contract", func() {
					It("should perform 2 transfers and revert 2 transfers", func() {
						// setup approval to send transfer without memo
						alloc := defaultSingleAlloc
						alloc[0].AllowedPacketData = []string{""}
						appArgs := defaultApproveArgs.WithArgs(alloc)
						s.setTransferApprovalForContract(appArgs)
						// Send some funds to the InterchainSender
						// to perform internal transfers
						initialContractBal := math.NewInt(1e18)
						err := evmosutil.FundAccountWithBaseDenom(s.chainA.GetContext(), s.app.BankKeeper, contractAddr.Bytes(), initialContractBal.Int64())
						Expect(err).To(BeNil(), "error while funding account")

						// get initial balances
						initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

						// use half of the allowance when calling the fn
						// because in total we'll try to send (2 * amt)
						// with 4 IBC transfers (2 will succeed & 2 will revert)
						amt := defaultCmnCoins[0].ToSDKType().Amount.QuoRaw(2)
						args := contracts.CallArgs{
							PrivKey:      s.privKey,
							ContractAddr: senderCallerContractAddr,
							ContractABI:  interchainSenderCallerContract.ABI,
							MethodName:   "transfersWithRevert",
							GasPrice:     gasPrice,
							Args: []interface{}{
								s.address,
								s.transferPath.EndpointA.ChannelConfig.PortID,
								s.transferPath.EndpointA.ChannelID,
								s.bondDenom,
								amt.BigInt(),
								s.chainB.SenderAccount.GetAddress().String(), // receiver
							},
						}

						logCheckArgs := passCheck.WithExpEvents([]string{ics20.EventTypeIBCTransfer, ics20.EventTypeIBCTransfer}...)

						res, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, logCheckArgs)
						Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
						Expect(res.IsOK()).To(BeTrue())
						fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)

						// the response should have two IBC transfer cosmos events (required for the relayer)
						expIBCPackets := 2
						ibcTransferCount := 0
						sendPacketCount := 0
						for _, event := range res.Events {
							if event.Type == transfertypes.EventTypeTransfer {
								ibcTransferCount++
							}
							if event.Type == channeltypes.EventTypeSendPacket {
								sendPacketCount++
							}
						}
						Expect(ibcTransferCount).To(Equal(expIBCPackets))
						Expect(sendPacketCount).To(Equal(expIBCPackets))

						// Check that 2 packages were created
						pkgs := s.app.IBCKeeper.ChannelKeeper.GetAllPacketCommitments(s.chainA.GetContext())
						Expect(pkgs).To(HaveLen(expIBCPackets))

						// check that the escrow amount corresponds to the 2 transfers
						coinsEscrowed := s.app.TransferKeeper.GetTotalEscrowForDenom(s.chainA.GetContext(), s.bondDenom)
						Expect(coinsEscrowed.Amount).To(Equal(amt))

						amtTransferredFromContract := math.NewInt(45)
						finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
						Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(amt).Sub(fees).Add(amtTransferredFromContract)))

						contractFinalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), contractAddr.Bytes(), s.bondDenom)
						Expect(contractFinalBalance.Amount).To(Equal(initialContractBal.Sub(amtTransferredFromContract)))
					})
				})
			})
		})

		Context("IBC coin", func() {
			var (
				ibcDenom                   = teststypes.UosmoIbcdenom
				amt, _                     = math.NewIntFromString("1000000000000000000000")
				sentAmt, _                 = math.NewIntFromString("100000000000000000000")
				coinOsmo                   = sdk.NewCoin(ibcDenom, amt)
				coins                      = sdk.NewCoins(coinOsmo)
				initialOsmoBalance         sdk.Coin
				defaultTransferIbcCoinArgs contracts.CallArgs
			)
			BeforeEach(func() {
				// set IBC denom trace
				s.app.TransferKeeper.SetDenomTrace(
					s.chainA.GetContext(),
					transfertypes.DenomTrace{
						Path:      teststypes.UosmoDenomtrace.Path,
						BaseDenom: teststypes.UosmoDenomtrace.BaseDenom,
					},
				)

				// Mint IBC coins and add them to sender balance
				err = s.app.BankKeeper.MintCoins(s.chainA.GetContext(), inflationtypes.ModuleName, coins)
				s.Require().NoError(err)
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), inflationtypes.ModuleName, s.chainA.SenderAccount.GetAddress(), coins)
				s.Require().NoError(err)

				initialOsmoBalance = s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), ibcDenom)
				Expect(initialOsmoBalance.Amount).To(Equal(amt))

				defaultTransferIbcCoinArgs = defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					ibcDenom,
					sentAmt.BigInt(),
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)
			})

			Context("without authorization", func() {
				It("should not transfer IBC coin", func() {
					// initialEvmosBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferIbcCoinArgs, execRevertedCheck)
					Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

					// check only fees were deducted from sending account
					// TODO: fees are not calculated correctly with this logic
					// fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					// finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					// Expect(finalBalance.Amount).To(Equal(initialEvmosBalance.Amount.Sub(fees)))

					// check IBC coins balance remains unchanged
					finalOsmoBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), ibcDenom)
					Expect(finalOsmoBalance.Amount).To(Equal(initialOsmoBalance.Amount))
				})
			})

			Context("with authorization", func() {
				BeforeEach(func() {
					// create grant to allow spending the ibc coins
					args := defaultApproveArgs.WithArgs([]cmn.ICS20Allocation{
						{
							SourcePort:        ibctesting.TransferPort,
							SourceChannel:     s.transferPath.EndpointA.ChannelID,
							SpendLimit:        []cmn.Coin{{Denom: ibcDenom, Amount: amt.BigInt()}},
							AllowList:         []string{},
							AllowedPacketData: []string{"memo"},
						},
					})
					s.setTransferApprovalForContract(args)
				})

				It("should transfer IBC coin", func() {
					initialEvmosBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferIbcCoinArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					Expect(out[0]).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// Check the allowance spend limit is updated
					authz, _ := s.app.AuthzKeeper.GetAuthorization(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
					Expect(authz).NotTo(BeNil(), "expected one authorization")
					Expect(authz.MsgTypeURL()).To(Equal(ics20.TransferMsgURL))
					transferAuthz := authz.(*transfertypes.TransferAuthorization)
					Expect(transferAuthz.Allocations[0].SpendLimit.AmountOf(ibcDenom)).To(Equal(amt.Sub(sentAmt)))

					// check only fees were deducted from sending account
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(initialEvmosBalance.Amount.Sub(fees)))

					// check sent tokens were deducted from sending account
					finalOsmoBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), ibcDenom)
					Expect(finalOsmoBalance.Amount).To(Equal(initialOsmoBalance.Amount.Sub(sentAmt)))
				})
			})
		})

		Context("transfer ERC20", func() {
			var (
				// denom is the registered token pair denomination
				denom string
				// erc20Addr is the address of the ERC20 contract
				erc20Addr                common.Address
				defaultTransferERC20Args contracts.CallArgs
				// sentAmount is the amount of tokens to send for testing
				sentAmount = big.NewInt(1000)
			)

			BeforeEach(func() {
				erc20Addr = s.setupERC20ContractTests(sentAmount)

				// Register ERC20 token pair to send via IBC
				_, err := s.app.Erc20Keeper.RegisterERC20(s.chainA.GetContext(), erc20Addr)
				Expect(err).To(BeNil(), "error while registering the token pair: %v", err)

				denom = fmt.Sprintf("erc20/%s", erc20Addr.String())

				defaultTransferERC20Args = defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					denom,
					sentAmount,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)
			})

			Context("without authorization", func() {
				tryERC20Transfer := func() {
					// initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferERC20Args, execRevertedCheck)
					Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

					// check only fees were deducted from sending account
					// TODO: fees are not calculated correctly with this logic
					// fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					// finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					// Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

					// check Erc20 balance remained unchanged by sent amount
					balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(balance).To(Equal(sentAmount), "address does not have the expected amount of tokens")
				}

				It("should not transfer registered ERC-20 token", func() {
					tryERC20Transfer()
				})

				Context("with authorization, but not for ERC20 token", func() {
					BeforeEach(func() {
						// create grant to allow spending the ibc coins
						args := defaultApproveArgs.WithArgs([]cmn.ICS20Allocation{
							{
								SourcePort:    ibctesting.TransferPort,
								SourceChannel: s.transferPath.EndpointA.ChannelID,
								SpendLimit:    []cmn.Coin{{Denom: teststypes.UosmoIbcdenom, Amount: big.NewInt(10000)}},
								AllowList:     []string{},
							},
						})
						s.setTransferApprovalForContract(args)
					})

					It("should not transfer registered ERC-20 token", func() {
						tryERC20Transfer()
					})
				})
			})

			Context("with authorization", func() {
				BeforeEach(func() {
					// create grant to allow spending the erc20 tokens
					args := defaultApproveArgs.WithArgs([]cmn.ICS20Allocation{
						{
							SourcePort:        ibctesting.TransferPort,
							SourceChannel:     s.transferPath.EndpointA.ChannelID,
							SpendLimit:        []cmn.Coin{{Denom: denom, Amount: sentAmount}},
							AllowList:         []string{},
							AllowedPacketData: []string{"memo"},
						},
					})
					s.setTransferApprovalForContract(args)
				})

				It("should transfer registered ERC-20 token", func() {
					initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferERC20Args, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					Expect(out[0]).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// The allowance is spent after the transfer thus the authorization is deleted
					authz, _ := s.app.AuthzKeeper.GetAuthorization(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
					Expect(authz).To(BeNil())

					// check only fees were deducted from sending account
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(initialBalance.Amount.Sub(fees)))

					// check Erc20 balance was reduced by sent amount
					balance := s.app.Erc20Keeper.BalanceOf(
						s.chainA.GetContext(),
						evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
						erc20Addr,
						s.address,
					)
					Expect(balance.Int64()).To(BeZero(), "address does not have the expected amount of tokens")
				})
			})
		})
	})

	Context("transfer a contract's funds", func() {
		var defaultTransferArgs contracts.CallArgs

		BeforeEach(func() {
			defaultTransferArgs = defaultCallArgs.WithMethodName(
				"testTransferContractFunds",
			)
		})

		Context("transfer 'aevmos", func() {
			var defaultTransferEvmosArgs contracts.CallArgs
			BeforeEach(func() {
				// send some funds to the contract from which the funds will be sent
				err = evmosutil.FundAccountWithBaseDenom(s.chainA.GetContext(), s.app.BankKeeper, contractAddr.Bytes(), amt)
				Expect(err).To(BeNil())
				senderInitialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), contractAddr.Bytes(), s.bondDenom)
				Expect(senderInitialBalance.Amount).To(Equal(math.NewInt(amt)))

				defaultTransferEvmosArgs = defaultTransferArgs.WithArgs(
					s.transferPath.EndpointA.ChannelConfig.PortID,
					s.transferPath.EndpointA.ChannelID,
					s.bondDenom,
					defaultCmnCoins[0].Amount,
					s.chainB.SenderAccount.GetAddress().String(), // receiver
					s.chainB.GetTimeoutHeight(),
					uint64(0), // disable timeout timestamp
					"memo",
				)
			})

			Context("without authorization", func() {
				It("should not transfer funds", func() {
					initialBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), contractAddr.Bytes(), s.bondDenom)

					_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferEvmosArgs, execRevertedCheck)
					Expect(err).To(HaveOccurred(), "error while calling the smart contract: %v", err)

					// check sent tokens remained unchanged from sending account (contract)
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), contractAddr.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(initialBalance.Amount))
				})
			})

			Context("with authorization", func() {
				BeforeEach(func() {
					// set approval to transfer 'aevmos'
					s.setTransferApprovalForContract(defaultApproveArgs)
				})

				It("should transfer funds", func() {
					initialSignerBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)

					logCheckArgs := passCheck.WithExpEvents(ics20.EventTypeIBCTransfer)

					res, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultTransferEvmosArgs, logCheckArgs)
					Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)

					out, err := s.precompile.Unpack(ics20.TransferMethod, ethRes.Ret)
					Expect(err).To(BeNil(), "error while unpacking response: %v", err)
					// check sequence in returned data
					Expect(out[0]).To(Equal(uint64(1)))

					s.chainA.NextBlock()

					// The allowance is spent after the transfer thus the authorization is deleted
					authz, _ := s.app.AuthzKeeper.GetAuthorization(s.chainA.GetContext(), contractAddr.Bytes(), s.address.Bytes(), ics20.TransferMsgURL)
					Expect(authz).To(BeNil())

					// check sent tokens were deducted from sending account
					finalBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), contractAddr.Bytes(), s.bondDenom)
					Expect(finalBalance.Amount).To(Equal(math.ZeroInt()))

					// tx fees are paid by the tx signer
					fees := math.NewIntFromBigInt(gasPrice).MulRaw(res.GasUsed)
					finalSignerBalance := s.app.BankKeeper.GetBalance(s.chainA.GetContext(), s.address.Bytes(), s.bondDenom)
					Expect(finalSignerBalance.Amount).To(Equal(initialSignerBalance.Amount.Sub(fees)))
				})
			})
		})
	})

	// ===============================================
	// 					QUERIES
	// ===============================================

	Context("allowance query method", func() {
		var defaultAllowanceArgs contracts.CallArgs
		BeforeEach(func() {
			s.setTransferApprovalForContract(defaultApproveArgs)
			defaultAllowanceArgs = defaultCallArgs.
				WithMethodName("testAllowance").
				WithArgs(contractAddr, s.address)
		})

		It("should return allocations", func() {
			_, ethRes, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, defaultAllowanceArgs, passCheck)
			Expect(err).To(BeNil(), "error while calling the precompile")

			var out []cmn.ICS20Allocation
			err = interchainSenderContract.ABI.UnpackIntoInterface(&out, "testAllowance", ethRes.Ret)
			Expect(err).To(BeNil(), "error while unpacking the output: %v", err)
			Expect(out).To(HaveLen(1))
			Expect(len(out)).To(Equal(len(defaultSingleAlloc)))
			Expect(out[0].SourcePort).To(Equal(defaultSingleAlloc[0].SourcePort))
			Expect(out[0].SourceChannel).To(Equal(defaultSingleAlloc[0].SourceChannel))
			Expect(out[0].SpendLimit).To(Equal(defaultSingleAlloc[0].SpendLimit))
			Expect(out[0].AllowList).To(HaveLen(0))
			Expect(out[0].AllowedPacketData).To(HaveLen(1))
			Expect(out[0].AllowedPacketData[0]).To(Equal("memo"))
		})
	})
})
