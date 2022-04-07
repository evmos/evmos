package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/statedb"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

func (suite *KeeperTestSuite) TestRegisterDevFeeInfo() {
	deployer := tests.GenerateAddress()
	fakeDeployer := tests.GenerateAddress()
	contract1 := crypto.CreateAddress(deployer, 1)
	factory1 := contract1
	factory2 := crypto.CreateAddress(factory1, 0)
	codeHash := common.Hex2Bytes("fa98cd094c09bb300de0037ba34e94f569b145ce8baa36ed863a08d7b7433f8d")
	contractAccount := statedb.Account{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: codeHash,
	}
	deployerAccount := statedb.Account{
		Balance:  big.NewInt(0),
		CodeHash: crypto.Keccak256(nil),
	}

	testCases := []struct {
		name         string
		deployer     sdk.AccAddress
		withdraw     sdk.AccAddress
		contract     common.Address
		nonces       []uint64
		malleate     func()
		expPass      bool
		errorMessage string
	}{
		{
			"ok - contract deployed by EOA",
			sdk.AccAddress(deployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set deployer and contract accounts
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
			},
			true,
			"",
		},
		{
			"ok - contract deployed by factory in factory",
			sdk.AccAddress(deployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			crypto.CreateAddress(factory2, 1),
			[]uint64{1, 0, 1},
			func() {
				// set deployer and contract accounts
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, crypto.CreateAddress(factory2, 1), contractAccount)
				s.Require().NoError(err)
			},
			true,
			"",
		},
		{
			"ok - omit withdraw address, it is stored as empty string",
			sdk.AccAddress(deployer.Bytes()),
			nil,
			contract1,
			[]uint64{1},
			func() {
				// set deployer and contract accounts
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
			},
			true,
			"",
		},
		{
			"ok - deployer == withdraw, withdraw is stored as empty string",
			sdk.AccAddress(deployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set deployer and contract accounts
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
			},
			true,
			"",
		},
		{
			"not ok - deployer account not found",
			sdk.AccAddress(deployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set only contract account
				err := s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
			},
			false,
			"deployer account not found",
		},
		{
			"not ok - deployer cannot be a contract",
			sdk.AccAddress(contract1.Bytes()),
			sdk.AccAddress(contract1.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set contract account
				err := s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
			},
			false,
			"deployer cannot be a contract",
		},
		{
			"not ok - contract is already registered",
			sdk.AccAddress(deployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set deployer and contract accounts
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
				msg := types.NewMsgRegisterDevFeeInfo(
					contract1,
					sdk.AccAddress(deployer.Bytes()),
					sdk.AccAddress(deployer.Bytes()),
					[]uint64{1},
				)
				ctx := sdk.WrapSDKContext(suite.ctx)
				suite.app.FeesKeeper.RegisterDevFeeInfo(ctx, msg)
			},
			false,
			"contract is already registered",
		},
		{
			"not ok - not contract deployer",
			sdk.AccAddress(fakeDeployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set deployer, fakeDeployer and contract accounts
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, fakeDeployer, deployerAccount)
				s.Require().NoError(err)
				err = s.app.EvmKeeper.SetAccount(s.ctx, contract1, contractAccount)
				s.Require().NoError(err)
			},
			false,
			"not contract deployer",
		},
		{
			"not ok - contract not deployed",
			sdk.AccAddress(deployer.Bytes()),
			sdk.AccAddress(deployer.Bytes()),
			contract1,
			[]uint64{1},
			func() {
				// set deployer account
				err := s.app.EvmKeeper.SetAccount(s.ctx, deployer, deployerAccount)
				s.Require().NoError(err)
			},
			false,
			"contract has no code",
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			tc.malleate()

			ctx := sdk.WrapSDKContext(suite.ctx)
			msg := types.NewMsgRegisterDevFeeInfo(tc.contract, tc.deployer, tc.withdraw, tc.nonces)

			res, err := suite.app.FeesKeeper.RegisterDevFeeInfo(ctx, msg)
			expRes := &types.MsgRegisterDevFeeInfoResponse{}
			suite.Commit()

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res, tc.name)

				fee, ok := suite.app.FeesKeeper.GetFeeInfo(suite.ctx, tc.contract)
				suite.Require().True(ok, "unregistered fee")
				suite.Require().Equal(tc.contract.String(), fee.ContractAddress, "wrong contract")
				suite.Require().Equal(tc.deployer.String(), fee.DeployerAddress, "wrong deployer")
				if tc.withdraw.String() != tc.deployer.String() {
					suite.Require().Equal(tc.withdraw.String(), fee.WithdrawAddress, "wrong withdraw address")
				} else {
					suite.Require().Equal("", fee.WithdrawAddress, "wrong withdraw address")
				}
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().Contains(err.Error(), tc.errorMessage)
			}
		})
	}
}
