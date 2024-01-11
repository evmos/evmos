package integration_test_util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
)

// TxPrepareContextWithdrawDelegatorAndValidatorReward prepares context for withdraw delegator and validator reward.
// It does create delegation, allocate reward, commit state and wait a few blocks for reward to increase.
func (suite *ChainIntegrationTestSuite) TxPrepareContextWithdrawDelegatorAndValidatorReward(delegator *itutiltypes.TestAccount, delegate uint8, waitXBlocks uint8) (valAddr sdk.ValAddress) {
	validatorAddr := suite.GetValidatorAddress(1)

	val := suite.ChainApp.StakingKeeper().Validator(suite.CurrentContext, validatorAddr)

	valReward := suite.NewBaseCoin(1)
	delegationAmount := suite.NewBaseCoin(int64(delegate))

	distAcc := suite.ChainApp.DistributionKeeper().GetDistributionAccount(suite.CurrentContext)
	suite.MintCoinToModuleAccount(distAcc, suite.NewBaseCoin(int64(int(delegate)*int(10+waitXBlocks))))
	suite.ChainApp.AccountKeeper().SetModuleAccount(suite.CurrentContext, distAcc)

	suite.MintCoin(delegator, delegationAmount)

	suite.Commit()

	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.GetCosmosAddress().String(),
		ValidatorAddress: validatorAddr.String(),
		Amount:           delegationAmount,
	}
	_, _, err := suite.DeliverTx(suite.CurrentContext, delegator, nil, msgDelegate)
	suite.Require().NoError(err)
	suite.Commit()

	for c := 1; c <= int(waitXBlocks); c++ {
		suite.ChainApp.DistributionKeeper().AllocateTokensToValidator(suite.CurrentContext, val, sdk.NewDecCoinsFromCoins(valReward))
		suite.Commit()
	}

	return validatorAddr
}

// GetValidatorAddress returns the validator address of the validator with the given number.
// Due to there is a bug that the validator address is delivered from tendermint pubkey instead of cosmos pubkey in tendermint mode.
// So this function is used to correct the validator address in tendermint mode.
func (suite *ChainIntegrationTestSuite) GetValidatorAddress(number int) sdk.ValAddress {
	validator := suite.ValidatorAccounts.Number(number)

	if suite.HasTendermint() {
		return sdk.ValAddress(validator.GetTmPubKey().Address())
	}

	return validator.GetValidatorAddress()
}
