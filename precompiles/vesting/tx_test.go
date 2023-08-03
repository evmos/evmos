// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting_test

import (
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v13/precompiles/common"
	"github.com/evmos/evmos/v13/precompiles/testutil"
	"github.com/evmos/evmos/v13/precompiles/vesting"
	evmosutil "github.com/evmos/evmos/v13/testutil"
	evmosutiltx "github.com/evmos/evmos/v13/testutil/tx"
	evmostypes "github.com/evmos/evmos/v13/types"
	"github.com/evmos/evmos/v13/utils"
	vestingtypes "github.com/evmos/evmos/v13/x/vesting/types"
)

var (
	balances         = []cmn.Coin{{Denom: utils.BaseDenom, Amount: big.NewInt(1000)}}
	quarter          = []cmn.Coin{{Denom: utils.BaseDenom, Amount: big.NewInt(250)}}
	balancesSdkCoins = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseDenom, 1000))
	toAddr           = evmosutiltx.GenerateAddress()
	funderAddr       = evmosutiltx.GenerateAddress()
	diffFunderAddr   = evmosutiltx.GenerateAddress()
	lockupPeriods    = []vesting.Period{{Length: 5000, Amount: balances}}
	vestingPeriods   = []vesting.Period{
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
	}
)

func (s *PrecompileTestSuite) TestCreateClawbackVestingAccount() {
	method := s.precompile.Methods[vesting.CreateClawbackVestingAccountMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			func(data []byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			name: "fail - different origin than vesting address",
			malleate: func() []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					funderAddr,
					differentAddr,
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "does not match the from address",
		},
		{
			"success",
			func() []interface{} {
				return []interface{}{
					funderAddr,
					s.address,
				}
			},
			20000,
			func(data []byte) {
				success, err := s.precompile.Unpack(vesting.CreateClawbackVestingAccountMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				// Check if the vesting account was created
				_, err = s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: sdk.AccAddress(s.address.Bytes()).String()})
				s.Require().NoError(err)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.CreateClawbackVestingAccount(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestFundVestingAccount() {
	method := s.precompile.Methods[vesting.FundVestingAccountMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			func(data []byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 5, 0),
		},
		{
			name: "fail - different origin than funder address",
			malleate: func() []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					differentAddr,
					toAddr,
					uint64(time.Now().Unix()),
					lockupPeriods,
					vestingPeriods,
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "does not match the from address",
		},
		{
			"success",
			func() []interface{} {
				s.CreateTestClawbackVestingAccount()
				err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, toAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(100))))
				return []interface{}{
					s.address,
					toAddr,
					uint64(time.Now().Unix()),
					lockupPeriods,
					vestingPeriods,
				}
			},
			20000,
			func(data []byte) {
				success, err := s.precompile.Unpack(vesting.CreateClawbackVestingAccountMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				// Check if the vesting account was created
				vestingAcc, err := s.app.VestingKeeper.Balances(s.ctx, &vestingtypes.QueryBalancesRequest{Address: sdk.AccAddress(toAddr.Bytes()).String()})
				s.Require().NoError(err)
				s.Require().Equal(vestingAcc.Locked, balancesSdkCoins)
				s.Require().Equal(vestingAcc.Unvested, balancesSdkCoins)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.FundVestingAccount(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestClawback() {
	method := s.precompile.Methods[vesting.ClawbackMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			func(data []byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		{
			name: "fail - different origin than funder address",
			malleate: func() []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					differentAddr,
					toAddr,
					s.address,
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "does not match the funder address",
		},
		{
			"success",
			func() []interface{} {
				s.CreateTestClawbackVestingAccount()
				s.FundTestClawbackVestingAccount()
				return []interface{}{
					s.address,
					toAddr,
					s.address,
				}
			},
			20000,
			func(data []byte) {
				success, err := s.precompile.Unpack(vesting.ClawbackMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.Clawback(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestUpdateVestingFunder() {
	method := s.precompile.Methods[vesting.ClawbackMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			func(data []byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		{
			name: "fail - different origin than funder address",
			malleate: func() []interface{} {
				differentAddr := evmosutiltx.GenerateAddress()
				return []interface{}{
					differentAddr,
					toAddr,
					s.address,
				}
			},
			gas:         200000,
			expError:    true,
			errContains: "does not match the funder address",
		},
		{
			"success",
			func() []interface{} {
				s.CreateTestClawbackVestingAccount()
				vestingAcc := s.app.AccountKeeper.GetAccount(s.ctx, toAddr.Bytes())
				va, ok := vestingAcc.(*vestingtypes.ClawbackVestingAccount)
				s.Require().True(ok)
				s.Require().Equal(va.FunderAddress, sdk.AccAddress(s.address.Bytes()).String())
				return []interface{}{
					s.address,
					diffFunderAddr,
					toAddr,
				}
			},
			20000,
			func(data []byte) {
				success, err := s.precompile.Unpack(vesting.UpdateVestingFunderMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				// Check if the vesting account has a new funder address
				vestingAcc := s.app.AccountKeeper.GetAccount(s.ctx, toAddr.Bytes())
				va, ok := vestingAcc.(*vestingtypes.ClawbackVestingAccount)
				s.Require().True(ok)
				s.Require().Equal(va.FunderAddress, sdk.AccAddress(diffFunderAddr.Bytes()).String())
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var contract *vm.Contract
			contract, s.ctx = testutil.NewPrecompileContract(s.T(), s.ctx, s.address, s.precompile, tc.gas)

			bz, err := s.precompile.UpdateVestingFunder(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestConvertVestingAccount() {
	method := s.precompile.Methods[vesting.ConvertVestingAccountMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			func(data []byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - incorrect address",
			func() []interface{} {
				return []interface{}{
					"asda412",
				}
			},
			200000,
			func(data []byte) {},
			true,
			"invalid type for vestingAddress",
		},
		{
			"success",
			func() []interface{} {
				s.CreateTestClawbackVestingAccount()
				return []interface{}{
					toAddr,
				}
			},
			20000,
			func(data []byte) {
				success, err := s.precompile.Unpack(vesting.ConvertVestingAccountMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(success[0], true)

				// Check if the vesting account was converted back to an EthAccountI
				account := s.app.AccountKeeper.GetAccount(s.ctx, toAddr.Bytes())
				_, ok := account.(evmostypes.EthAccountI)
				s.Require().True(ok)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			bz, err := s.precompile.ConvertVestingAccount(s.ctx, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz)
			}
		})
	}
}
