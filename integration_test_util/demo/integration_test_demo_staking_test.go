package demo

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *DemoTestSuite) Test_TX_Delegate() {
	validators := suite.CITS.ChainApp.StakingKeeper().GetAllValidators(suite.Ctx())
	fmt.Println("Validators:", len(validators))
	for _, validator := range validators {
		fmt.Println(validator.OperatorAddress)
	}

	delegator := suite.CITS.WalletAccounts.Number(1)
	validatorAddr := suite.CITS.GetValidatorAddress(1).String()
	delegationAmount := suite.CITS.NewBaseCoin(1).Amount
	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.GetCosmosAddress().String(),
		ValidatorAddress: validatorAddr,
		Amount: sdk.Coin{
			Denom:  suite.CITS.ChainConstantsConfig.GetMinDenom(),
			Amount: delegationAmount,
		},
	}

	_, _, err := suite.CITS.DeliverTx(suite.Ctx(), delegator, nil, msgDelegate)
	suite.Require().NoError(err)

	suite.Commit()

	delegations, err := suite.CITS.QueryClients.Staking.DelegatorDelegations(suite.Ctx(), &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator.GetCosmosAddress().String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(delegations)
	suite.NotEmpty(delegations.DelegationResponses)
}
