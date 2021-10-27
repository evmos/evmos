package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/keeper"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

const (
	erc20Name       = "coin"
	erc20Symbol     = "token"
	cosmosTokenName = "coinevm"
	defaultExponent = uint32(18)
	zeroExponent    = uint32(0)
)

func (suite *KeeperTestSuite) setupNewTokenPair() common.Address {
	suite.SetupTest()
	contractAddr := suite.DeployContract(erc20Name, erc20Symbol)
	suite.Commit()
	pair := types.NewTokenPair(contractAddr, cosmosTokenName, true)
	err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
	suite.Require().NoError(err)
	return contractAddr
}

func (suite KeeperTestSuite) TestRegisterTokenPair() {
	var (
		contractAddr common.Address
		pair         types.TokenPair
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token ERC20 already registered",
			func() {
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), pair.GetID())
			},
			false,
		},
		{
			"denom already registered",
			func() {
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
			},
			false,
		},
		{
			"meta data already stored",
			func() {
				suite.app.IntrarelayerKeeper.CreateMetadata(suite.ctx, pair)
			},
			false,
		},
		{
			"ok",
			func() {},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
			suite.Commit()
			pair = types.NewTokenPair(contractAddr, cosmosTokenName, true)

			tc.malleate()

			err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
			metadata, found := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, cosmosTokenName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				// Metadata variables
				suite.Require().True(found)
				suite.Require().Equal(cosmosTokenName, metadata.Base)
				suite.Require().Equal(contractAddr.String(), metadata.Name)
				suite.Require().Equal(erc20Name, metadata.Display)
				suite.Require().Equal(erc20Symbol, metadata.Symbol)
				// Denom units
				suite.Require().Equal(len(metadata.DenomUnits), 2)
				suite.Require().Equal(cosmosTokenName, metadata.DenomUnits[0].Denom)
				suite.Require().Equal(uint32(zeroExponent), metadata.DenomUnits[0].Exponent)
				suite.Require().Equal(erc20Name, metadata.DenomUnits[1].Denom)
				// Default exponent at contract creation is 18
				suite.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(defaultExponent))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEnableRelayWithContext() {
	// Default enabled value is True
	contractAddr := suite.setupNewTokenPair()
	id := suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, contractAddr.String())
	suite.Require().True(len(id) > 0)
	pair, found := suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
	suite.Require().True(found)
	suite.Require().True(pair.Enabled)

	// Dissable it
	pair, err := suite.app.IntrarelayerKeeper.ToggleRelay(suite.ctx, contractAddr.String())
	suite.Require().NoError(err)
	suite.Require().False(pair.Enabled)

	// Request the pair using the GetPairToken func to make sure that is updated on the db
	pair, found = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
	suite.Require().True(found)
	suite.Require().False(pair.Enabled)

	// Reenable it
	pair, err = suite.app.IntrarelayerKeeper.ToggleRelay(suite.ctx, contractAddr.String())
	suite.Require().NoError(err)
	suite.Require().True(pair.Enabled)

	// Request the pair using the GetPairToken func to make sure that is updated on the db
	pair, found = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
	suite.Require().True(found)
	suite.Require().True(pair.Enabled)

	// Try to toggle a not registered token
	pair, found = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, make([]byte, 0))
	suite.Require().False(found)
}

func (suite KeeperTestSuite) TestUpdateTokenPairERC20() {
	var (
		contractAddr    common.Address
		pair            types.TokenPair
		metadata        sdk.Metadata
		newContractAddr common.Address
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"token not registered",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenName, true)
			},
			false,
		},
		{
			"token not registered - pair not found",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenName, true)

				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
		},
		{
			"token not registered - Metadata not found",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenName, true)

				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
		},
		{
			"newErc20 not found",
			func() {
				contractAddr = suite.setupNewTokenPair()
				newContractAddr = common.Address{}
			},
			false,
		},
		{
			"ok",
			func() {
				contractAddr = suite.setupNewTokenPair()
				id := suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)
				metadata, _ = suite.app.BankKeeper.GetDenomMetaData(suite.ctx, cosmosTokenName)
				suite.Commit()

				// Deploy a new contrat with the same values
				newContractAddr = suite.DeployContract(erc20Name, erc20Symbol)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var err error
			pair, err = suite.app.IntrarelayerKeeper.UpdateTokenPairERC20(suite.ctx, contractAddr, newContractAddr)
			metadata, _ = suite.app.BankKeeper.GetDenomMetaData(suite.ctx, cosmosTokenName)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(newContractAddr.String(), pair.Erc20Address)
				suite.Require().Equal(keeper.CreateDenomDescription(newContractAddr.String()), metadata.Description)
			} else {
				suite.Require().Error(err, tc.name)
				if suite.app.IntrarelayerKeeper.IsTokenPairRegistered(suite.ctx, pair.GetID()) {
					suite.Require().Equal(contractAddr.String(), pair.Erc20Address, "check pair")
					suite.Require().Equal(keeper.CreateDenomDescription(contractAddr.String()), metadata.Description, "check metadata")
				}
			}
		})
	}
}
