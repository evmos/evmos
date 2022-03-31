package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/tests"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

func (suite *KeeperTestSuite) TestRegisterFeeContract() {
	addr1 := tests.GenerateAddress()
	factory1 := crypto.CreateAddress(addr1, 1)
	factory2 := crypto.CreateAddress(factory1, 0)

	testCases := []struct {
		name     string
		deployer sdk.AccAddress
		withdraw sdk.AccAddress
		contract common.Address
		nonces   []uint64
		malleate func()
		expPass  bool
	}{
		{
			"ok - contract deployed by EOA",
			sdk.AccAddress(addr1.Bytes()),
			sdk.AccAddress(addr1.Bytes()),
			crypto.CreateAddress(addr1, 1),
			[]uint64{1},
			func() {},
			true,
		},
		{
			"ok - contract deployed by factory in factory",
			sdk.AccAddress(addr1.Bytes()),
			sdk.AccAddress(addr1.Bytes()),
			crypto.CreateAddress(factory2, 1),
			[]uint64{1, 0, 1},
			func() {},
			true,
		},
		{
			"not ok - contract already registered",
			sdk.AccAddress(addr1.Bytes()),
			sdk.AccAddress(addr1.Bytes()),
			factory1,
			[]uint64{1},
			func() {
				msg := types.NewMsgRegisterFeeContract(
					factory1,
					sdk.AccAddress(addr1.Bytes()),
					sdk.AccAddress(addr1.Bytes()),
					[]uint64{1},
				)
				ctx := sdk.WrapSDKContext(suite.ctx)
				suite.app.FeesKeeper.RegisterFeeContract(ctx, msg)
			},
			false,
		},
		{
			"not ok - not contract deployer",
			sdk.AccAddress(tests.GenerateAddress().Bytes()),
			sdk.AccAddress(addr1.Bytes()),
			crypto.CreateAddress(addr1, 1),
			[]uint64{1},
			func() {},
			false,
		},
		// TODO: not ok - contract not deployed
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			tc.malleate()

			ctx := sdk.WrapSDKContext(suite.ctx)
			msg := types.NewMsgRegisterFeeContract(tc.contract, tc.deployer, tc.withdraw, tc.nonces)

			res, err := suite.app.FeesKeeper.RegisterFeeContract(ctx, msg)
			expRes := &types.MsgRegisterFeeContractResponse{}
			suite.Commit()

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res, tc.name)

				fee, ok := suite.app.FeesKeeper.GetFee(suite.ctx, tc.contract)
				suite.Require().True(ok, "unregistered fee")
				suite.Require().Equal(tc.contract.String(), fee.ContractAddress, "wrong contract")
				suite.Require().Equal(tc.deployer.String(), fee.DeployerAddress, "wrong deployer")
				suite.Require().Equal(tc.withdraw.String(), fee.WithdrawAddress, "wrong withdraw address")
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}
