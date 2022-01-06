package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/ethermint/tests"

	"github.com/tharsis/evmos/x/erc20/types"
)

const (
	erc20Name          = "Coin Token"
	erc20Symbol        = "CTKN"
	cosmosTokenBase    = "acoin"
	cosmosTokenDisplay = "coin"
	defaultExponent    = uint32(18)
	zeroExponent       = uint32(0)
)

func (suite *KeeperTestSuite) setupRegisterERC20Pair() common.Address {
	suite.SetupTest()
	contractAddr := suite.DeployContract(erc20Name, erc20Symbol)
	suite.Commit()
	_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
	suite.Require().NoError(err)
	return contractAddr
}
func (suite *KeeperTestSuite) setupRegisterERC20PairMaliciousDelayed() common.Address {
	suite.SetupTest()
	contractAddr := suite.DeployContractMaliciousDelayed(erc20Name, erc20Symbol)
	suite.Commit()
	_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
	suite.Require().NoError(err)
	return contractAddr
}

func (suite *KeeperTestSuite) setupRegisterCoin() (banktypes.Metadata, *types.TokenPair) {
	suite.SetupTest()
	validMetadata := banktypes.Metadata{
		Description: "description of the token",
		Base:        cosmosTokenBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    cosmosTokenBase,
				Exponent: 0,
			},
			{
				Denom:    cosmosTokenBase[1:],
				Exponent: uint32(18),
			},
		},
		Name:    cosmosTokenBase,
		Symbol:  erc20Symbol,
		Display: cosmosTokenBase,
	}

	err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
	suite.Require().NoError(err)

	// pair := types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
	pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, validMetadata)
	suite.Require().NoError(err)
	suite.Commit()
	return validMetadata, pair
}

func (suite KeeperTestSuite) TestRegisterCoin() {
	metadata := banktypes.Metadata{
		Description: "description",
		Base:        cosmosTokenBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    cosmosTokenBase,
				Exponent: 0,
			},
			{
				Denom:    cosmosTokenDisplay,
				Exponent: defaultExponent,
			},
		},
		Name:    cosmosTokenBase,
		Symbol:  erc20Symbol,
		Display: cosmosTokenDisplay,
	}

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"denom already registered",
			func() {
				regPair := types.NewTokenPair(tests.GenerateAddress(), metadata.Base, true, types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, regPair.Denom, regPair.GetID())
				suite.Commit()
			},
			false,
		},
		{
			"token doesn't have supply",
			func() {
			},
			false,
		},
		{
			"metadata different that stored",
			func() {
				metadata.Base = cosmosTokenBase
				validMetadata := banktypes.Metadata{
					Description: "description",
					Base:        cosmosTokenBase,
					// NOTE: Denom units MUST be increasing
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    cosmosTokenBase,
							Exponent: 0,
						},
						{
							Denom:    cosmosTokenDisplay,
							Exponent: uint32(18),
						},
					},
					Name:    erc20Name,
					Symbol:  erc20Symbol,
					Display: cosmosTokenDisplay,
				}

				err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
				suite.Require().NoError(err)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, validMetadata)
			},
			false,
		},
		{
			"ok",
			func() {
				err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(metadata.Base, 1)})
				suite.Require().NoError(err)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, metadata)
			suite.Commit()

			expPair := &types.TokenPair{
				Erc20Address:  "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd",
				Denom:         "acoin",
				Enabled:       true,
				ContractOwner: 1,
			}

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(pair, expPair)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestRegisterERC20() {
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
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token ERC20 already registered",
			func() {
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), pair.GetID())
			},
			false,
		},
		{
			"denom already registered",
			func() {
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
			},
			false,
		},
		{
			"meta data already stored",
			func() {
				suite.app.Erc20Keeper.CreateCoinMetadata(suite.ctx, contractAddr)
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
			coinName := types.CreateDenom(contractAddr.String())
			pair = types.NewTokenPair(contractAddr, coinName, true, types.OWNER_EXTERNAL)

			tc.malleate()

			_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
			metadata, found := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				// Metadata variables
				suite.Require().True(found)
				suite.Require().Equal(coinName, metadata.Base)
				suite.Require().Equal(coinName, metadata.Name)
				suite.Require().Equal(types.SanitizeERC20Name(erc20Name), metadata.Display)
				suite.Require().Equal(erc20Symbol, metadata.Symbol)
				// Denom units
				suite.Require().Equal(len(metadata.DenomUnits), 2)
				suite.Require().Equal(coinName, metadata.DenomUnits[0].Denom)
				suite.Require().Equal(uint32(zeroExponent), metadata.DenomUnits[0].Exponent)
				suite.Require().Equal(types.SanitizeERC20Name(erc20Name), metadata.DenomUnits[1].Denom)
				// Default exponent at contract creation is 18
				suite.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(defaultExponent))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestToggleRelay() {
	var (
		contractAddr common.Address
		id           []byte
		pair         types.TokenPair
	)

	testCases := []struct {
		name         string
		malleate     func()
		expPass      bool
		relayEnabled bool
	}{
		{
			"token not registered",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
			},
			false,
			false,
		},
		{
			"token not registered - pair not found",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
			false,
		},
		{
			"disable relay",
			func() {
				contractAddr = suite.setupRegisterERC20Pair()
				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			},
			true,
			false,
		},
		{
			"disable and enable relay",
			func() {
				contractAddr = suite.setupRegisterERC20Pair()
				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				pair, _ = suite.app.Erc20Keeper.ToggleRelay(suite.ctx, contractAddr.String())
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var err error
			pair, err = suite.app.Erc20Keeper.ToggleRelay(suite.ctx, contractAddr.String())
			// Request the pair using the GetPairToken func to make sure that is updated on the db
			pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				if tc.relayEnabled {
					suite.Require().True(pair.Enabled)
				} else {
					suite.Require().False(pair.Enabled)
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestUpdateTokenPairERC20() {
	var (
		contractAddr    common.Address
		pair            types.TokenPair
		metadata        banktypes.Metadata
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
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
			},
			false,
		},
		{
			"token not registered - pair not found",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)

				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
		},
		{
			"token not registered - Metadata not found",
			func() {
				contractAddr = suite.DeployContract(erc20Name, erc20Symbol)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)

				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
		},
		{
			"newErc20 not found",
			func() {
				contractAddr = suite.setupRegisterERC20Pair()
				newContractAddr = common.Address{}
			},
			false,
		},
		{
			"empty denom units",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair()
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, banktypes.Metadata{Base: pair.Denom})
				suite.Commit()

				// Deploy a new contrat with the same values
				newContractAddr = suite.DeployContract(erc20Name, erc20Symbol)
			},
			false,
		},
		{
			"metadata ERC20 details mismatch",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair()
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, DenomUnits: []*banktypes.DenomUnit{{}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contrat with the same values
				newContractAddr = suite.DeployContract(erc20Name, erc20Symbol)
			},
			false,
		},
		{
			"no denom unit with ERC20 name",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair()
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contrat with the same values
				newContractAddr = suite.DeployContract(erc20Name, erc20Symbol)
			},
			false,
		},
		{
			"denom unit and ERC20 decimals mismatch",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair()
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contrat with the same values
				newContractAddr = suite.DeployContract(erc20Name, erc20Symbol)
			},
			false,
		},
		{
			"ok",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair()
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name, Exponent: 18}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
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
			pair, err = suite.app.Erc20Keeper.UpdateTokenPairERC20(suite.ctx, contractAddr, newContractAddr)
			metadata, _ = suite.app.BankKeeper.GetDenomMetaData(suite.ctx, types.CreateDenom(contractAddr.String()))

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(newContractAddr.String(), pair.Erc20Address)
				suite.Require().Equal(types.CreateDenomDescription(newContractAddr.String()), metadata.Description)
			} else {
				suite.Require().Error(err, tc.name)
				if suite.app.Erc20Keeper.IsTokenPairRegistered(suite.ctx, pair.GetID()) {
					suite.Require().Equal(contractAddr.String(), pair.Erc20Address, "check pair")
					suite.Require().Equal(types.CreateDenomDescription(contractAddr.String()), metadata.Description, "check metadata")
				}
			}
		})
	}
}
