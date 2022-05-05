package keeper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Performing EVM transactions", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()

		params := s.app.Erc20Keeper.GetParams(s.ctx)
		params.EnableEVMHook = true
		params.EnableErc20 = true
		s.app.Erc20Keeper.SetParams(s.ctx, params)
	})

	// Epoch mechanism for triggering allocation and distribution
	Context("with the ERC20 module and EVM Hook disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableEVMHook = false
			params.EnableErc20 = false
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the ERC20 module disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableErc20 = false
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the EVMHook disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableEVMHook = false
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the ERC20 module and EVM Hook enabled", func() {
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})
})
