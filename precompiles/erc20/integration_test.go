package erc20_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20 Extension - ", func() {
	BeforeEach(func() {
		s.SetupTest()
	})

	When("querying balance", func() {
		It("should return an existing balance", func() {
			Expect(true).To(BeFalse(), "Not implemented")
		})

		It("should return zero if balance only exists for other tokens", func() {
			Expect(true).To(BeFalse(), "Not implemented")
		})

		It("should fail if the account does not exist", func() {
			Expect(true).To(BeFalse(), "Not implemented")
		})

		It("should fail for an invalid number of arguments", func() {
			Expect(true).To(BeFalse(), "Not implemented")
		})

		It("should fail for an invalid address", func() {
			Expect(true).To(BeFalse(), "Not implemented")
		})
	})

})
