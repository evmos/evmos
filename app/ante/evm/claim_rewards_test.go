package evm_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/testutil"
	"github.com/evmos/evmos/v11/utils"
)

// TestClaimSufficientStakingRewards tests the ClaimSufficientStakingRewards function
func (suite *AnteTestSuite) TestClaimSufficientStakingRewards() {
	suite.SetupTest()

	// ----------------------------------------
	// Set up first account
	//
	addr, _ := testutil.NewAccAddressAndKey()
	initialBalance := sdk.Coins{sdk.Coin{Amount: sdk.NewInt(1e18), Denom: utils.BaseDenom}}
	err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, initialBalance)
	suite.Require().NoError(err, "failed to fund account")

	// ----------------------------------------
	// Set up validator
	//
	addr2, priv2 := testutil.NewAccAddressAndKey()
	valAddr := sdk.ValAddress(addr2.Bytes())
	fivePercent := sdk.NewDecWithPrec(5, 2)

	stakingHelper := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(fivePercent, fivePercent, fivePercent)
	stakingHelper.Denom = utils.BaseDenom

	stakingHelper.CreateValidator(valAddr, priv2.PubKey(), sdk.NewInt(1e18), true)
	stakingHelper.Delegate(addr, valAddr, sdk.NewInt(123456789))

	// Get all delegations
	delegations := suite.app.StakingKeeper.GetAllDelegations(suite.ctx)
	suite.T().Logf("Delegations: %v", delegations)

	suite.Require().True(false)
}
