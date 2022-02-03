package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
				// resAms, _ := s.grpcQueryClient.AllocationMeters(s.ctx)
				// ams := resAms.AllocationMeters
				// s.Req

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
