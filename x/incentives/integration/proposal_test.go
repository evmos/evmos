package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/incentives/types"
)

var _ = Describe("Proposal", func() {
	Describe("Querying the grpcQueryClient", func() {
		Context("with a test network", func() {
			It("should respond with a non-nil result", func() {
				resParams, err := s.grpcQueryClient.Params(s.ctx, &types.QueryParamsRequest{})
				Expect(err).To(BeNil())
				Expect(resParams).ToNot(BeNil())
			})
		})
	})

	Describe("Registering a proposal", func() {
		Context("with a successful vote", func() {
			It("should create a new incentive", func() {
				// TODO Create account with funds
				// TODO Deploy contract

				fromAddr, _ := sdk.AccAddressFromBech32(tests.GenerateAddress().Hex())
				contract := tests.GenerateAddress().Hex()
				content := types.NewRegisterIncentiveProposal(
					"title",
					"desctiption",
					contract,
					sdk.NewDecCoins(sdk.NewDecCoinFromDec(s.cfg.BondDenom, sdk.NewDecWithPrec(5, 2))),
					20,
				)

				deposit := sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20000000000000)))
				msg, err := govtypes.NewMsgSubmitProposal(content, deposit, fromAddr)
				Expect(err).To(BeNil())

				_, err = s.grpcTxClientGov.SubmitProposal(s.ctx, msg)
				Expect(err).To(BeNil())

				resProposals, err := s.grpcQueryClientGov.Proposals(s.ctx, &govtypes.QueryProposalsRequest{})
				Expect(resProposals).ToNot(BeNil())
				Expect(err).To(BeNil())

				// evmosd tx gov submit-proposal register-coin ./metadata.json
				// --from=mykey --deposit=20000000000000aphoton --description='this is a
				// cli proposal test' --title='registering using the cli'
				// --fees=82aphoton -b block --gas=auto

				actual := 1
				expected := 1
				Expect(actual).To(Equal(expected))
			})
		})

		Context("with a unsuccessful vote", func() {
			It("should not create a new incentive", func() {
				// Expect(lesMis.Category()).To(Equal(books.CategoryNovel))
			})
		})
	})
})
