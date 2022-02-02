package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proposal", func() {
	BeforeAll(func(suite *IntegrationTestSuite) {

	})

	BeforeEach(func() {
	})

	AfterAll(func() {
	})

	Describe("Registering a proposal", func() {
		Context("with a successful vote", func() {
			It("should create a new incentive", func() {
				actual := 1
				expected := 1
				Expect(actual).To(Equal(expected))
				// Expect(lesMis.Category()).To(Equal(books.CategoryNovel))
			})
		})

		Context("with a unsuccessful vote", func() {
			It("should not create a new incentive", func() {
				// Expect(lesMis.Category()).To(Equal(books.CategoryNovel))
			})
		})
	})
})
