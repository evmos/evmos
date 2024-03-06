// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/vesting"
	"github.com/evmos/evmos/v16/precompiles/vesting/testdata"
	evmosutil "github.com/evmos/evmos/v16/testutil"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v16/x/vesting/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// CallType is a struct that represents the type of call to be made to the
// precompile - either direct or through a smart contract.
type CallType struct {
	// name is the name of the call type
	name string
	// directCall is true if the call is to be made directly to the precompile
	directCall bool
}

// BuildCallArgs builds the call arguments for the integration test suite
// depending on the type of interaction.
func (s *PrecompileTestSuite) BuildCallArgs(
	callType CallType,
	contractAddr common.Address,
) (factory.CallArgs, evmtypes.EvmTxArgs) {
	var (
		to       common.Address
		callArgs = factory.CallArgs{}
		txArgs   = evmtypes.EvmTxArgs{}
	)

	if callType.directCall {
		callArgs.ContractABI = s.precompile.ABI
		to = s.precompile.Address()
	} else {
		to = contractAddr
		callArgs.ContractABI = testdata.VestingCallerContract.ABI
	}
	txArgs.To = &to
	return callArgs, txArgs
}

// FundTestClawbackVestingAccount funds the clawback vesting account with some tokens
func (s *PrecompileTestSuite) FundTestClawbackVestingAccount() {
	method := s.precompile.Methods[vesting.FundVestingAccountMethod]
	createArgs := []interface{}{s.keyring.GetAddr(0), toAddr, uint64(time.Now().Unix()), lockupPeriods, vestingPeriods}
	msg, _, _, _, _, err := vesting.NewMsgFundVestingAccount(createArgs, &method)
	s.Require().NoError(err)
	_, err = s.network.App.VestingKeeper.FundVestingAccount(s.network.GetContext(), msg)
	s.Require().NoError(err)
	vestingAcc, err := s.network.App.VestingKeeper.Balances(s.network.GetContext(), &vestingtypes.QueryBalancesRequest{Address: sdk.AccAddress(toAddr.Bytes()).String()})
	s.Require().NoError(err)
	s.Require().Equal(vestingAcc.Locked, balancesSdkCoins)
	s.Require().Equal(vestingAcc.Unvested, balancesSdkCoins)
}

// CreateTestClawbackVestingAccount creates a vesting account that can clawback.
// Useful for unit tests only
func (s *PrecompileTestSuite) CreateTestClawbackVestingAccount(ctx sdk.Context, funder, vestingAddr common.Address) {
	msgArgs := []interface{}{funder, vestingAddr, false}
	msg, _, _, err := vesting.NewMsgCreateClawbackVestingAccount(msgArgs)
	s.Require().NoError(err)
	err = evmosutil.FundAccount(ctx, s.network.App.BankKeeper, vestingAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
	s.Require().NoError(err)
	_, err = s.network.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
	s.Require().NoError(err)
}

// ExpectSimpleVestingAccount checks that the vesting account has the expected funder address
func (s *PrecompileTestSuite) ExpectSimpleVestingAccount(vestingAddr, funderAddr common.Address) {
	vestingAcc := s.GetVestingAccount(vestingAddr)
	funder, err := sdk.AccAddressFromBech32(vestingAcc.FunderAddress)
	Expect(err).ToNot(HaveOccurred(), "vesting account should have a valid funder address")
	Expect(funder.Bytes()).To(Equal(funderAddr.Bytes()), "vesting account should have the correct funder address")
}

// ExpectVestingAccount checks that the vesting account has the expected lockup and vesting periods.
func (s *PrecompileTestSuite) ExpectVestingAccount(vestingAddr common.Address, lockupPeriods, vestingPeriods []vesting.Period) {
	vestingAcc := s.GetVestingAccount(vestingAddr)
	// TODO: check for multiple lockup or vesting periods
	Expect(vestingAcc.LockupPeriods).To(HaveLen(len(lockupPeriods)), "vesting account should have the correct number of lockup periods")
	Expect(vestingAcc.LockupPeriods[0].Length).To(Equal(lockupPeriods[0].Length), "vesting account should have the correct lockup period length")
	Expect(vestingAcc.LockupPeriods[0].Amount[0].Denom).To(Equal(lockupPeriods[0].Amount[0].Denom), "vesting account should have the correct vestingPeriod amount")
	Expect(vestingAcc.LockupPeriods[0].Amount[0].Amount.BigInt()).To(Equal(lockupPeriods[0].Amount[0].Amount), "vesting account should have the correct vesting amount")
	Expect(vestingAcc.VestingPeriods).To(HaveLen(len(vestingPeriods)), "vesting account should have the correct number of vesting periods")
	Expect(vestingAcc.VestingPeriods[0].Length).To(Equal(vestingPeriods[0].Length), "vesting account should have the correct vesting period length")
	Expect(vestingAcc.VestingPeriods[0].Amount[0].Denom).To(Equal(vestingPeriods[0].Amount[0].Denom), "vesting account should have the correct vesting amount")
	Expect(vestingAcc.VestingPeriods[0].Amount[0].Amount.BigInt()).To(Equal(vestingPeriods[0].Amount[0].Amount), "vesting account should have the correct vesting amount")
}

// ExpectVestingFunder checks that the vesting funder of a given vesting account address is the given one.
func (s *PrecompileTestSuite) ExpectVestingFunder(vestingAddr common.Address, funderAddr common.Address) {
	vestingAcc := s.GetVestingAccount(vestingAddr)
	Expect(vestingAcc.FunderAddress).To(Equal(sdk.AccAddress(funderAddr.Bytes()).String()), "expected a different funder for the vesting account")
}

// GetVestingAccount returns the vesting account for the given address.
func (s *PrecompileTestSuite) GetVestingAccount(addr common.Address) *vestingtypes.ClawbackVestingAccount {
	acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), addr.Bytes())
	Expect(acc).ToNot(BeNil(), "vesting account should exist")
	vestingAcc, ok := acc.(*vestingtypes.ClawbackVestingAccount)
	Expect(ok).To(BeTrue(), "vesting account should be of type VestingAccount")
	return vestingAcc
}
