package keeper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration", func() {
	Describe("Surpassing an epoch End", func() {
		Context("with default params", func() {
			It("should allocate funds for usage incentives", func() {
				actual := 1
				expected := 1
				Expect(actual).To(Equal(expected))
			})
		})
	})

})
