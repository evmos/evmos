package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/tharsis/evmos/x/inflation/types"
)

var _ = Describe("Inflation", func() {
	Describe("Querying the grpcQueryClient", func() {
		Context("with a test network", func() {
			It("should respond with a non-nil result", func() {
				resParams, err := s.grpcQueryClient.Params(s.ctx, &types.QueryParamsRequest{})
				Expect(err).To(BeNil())
				Expect(resParams).ToNot(BeNil())
			})
		})
	})

	Describe("Surpassing an epoch end", func() {
		Context("with default params", func() {
			// TODO time travel
			It("should allocate coins to the inflation module", func() {
				// sdk.AccAddressFromHex()
				// moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

				req := &banktypes.QuerySupplyOfRequest{Denom: s.cfg.BondDenom}
				resBalance, err := s.grpcQueryClientBank.SupplyOf(s.ctx, req)
				Expect(err).To(BeNil())

				err = s.network.WaitForNextBlock()
				Expect(err).To(BeNil())

				nextResBalance, err := s.grpcQueryClientBank.SupplyOf(s.ctx, req)
				Expect(err).To(BeNil())
				Expect(resBalance.Amount).ToNot(Equal((nextResBalance.Amount)))
			})
		})
	})
})
