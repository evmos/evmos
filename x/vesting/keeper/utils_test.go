package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v19/testutil/integration/common/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/x/vesting/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
<<<<<<< HEAD
=======

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"

	evmante "github.com/evmos/evmos/v19/app/ante/evm"
	"github.com/evmos/evmos/v19/contracts"
	"github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/x/vesting/types"
>>>>>>> main
)

func (suite *KeeperTestSuite) setupClawbackVestingAccount(vestingAccount, funder keyring.Key, vestingPeriods, lockupPeriods sdkvesting.Periods, enableGovClawback bool) *types.ClawbackVestingAccount {
	// send a create vesting account tx
	createAccMsg := types.NewMsgCreateClawbackVestingAccount(funder.AccAddr, vestingAccount.AccAddr, enableGovClawback)
	res, err := suite.factory.ExecuteCosmosTx(vestingAccount.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{createAccMsg}, Gas: &gasLimit, GasPrice: &gasPrice})
	Expect(err).To(BeNil())
	Expect(res.IsOK()).To(BeTrue())
	Expect(suite.network.NextBlock()).To(BeNil())

	// Fund the clawback vesting accounts
	vestingStart := suite.network.GetContext().BlockTime()
	fundMsg := types.NewMsgFundVestingAccount(funder.AccAddr, vestingAccount.AccAddr, vestingStart, lockupPeriods, vestingPeriods)
	res, err = suite.factory.ExecuteCosmosTx(funder.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{fundMsg}})
	Expect(err).To(BeNil())
	Expect(res.IsOK()).To(BeTrue())
	Expect(suite.network.NextBlock()).To(BeNil())

	acc, err := suite.handler.GetAccount(vestingAccount.AccAddr.String())
	Expect(err).To(BeNil())
	var ok bool
	clawbackAccount, ok := acc.(*types.ClawbackVestingAccount)
	Expect(ok).To(BeTrue())

	return clawbackAccount
}
