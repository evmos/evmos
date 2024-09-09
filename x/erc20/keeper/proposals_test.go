package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/contracts"
	testfactory "github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	testutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v20/x/erc20/keeper"
	"github.com/evmos/evmos/v20/x/erc20/types"
	erc20mocks "github.com/evmos/evmos/v20/x/erc20/types/mocks"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/stretchr/testify/mock"
)

const (
	contractMinterBurner = iota + 1
	contractDirectBalanceManipulation
	contractMaliciousDelayed
)

const (
	erc20Name          = "Coin Token"
	erc20Symbol        = "CTKN"
	erc20Decimals      = uint8(18)
	cosmosTokenBase    = "acoin"
	cosmosTokenDisplay = "coin"
	cosmosDecimals     = uint8(6)
	defaultExponent    = uint32(18)
	zeroExponent       = uint32(0)
	ibcBase            = "ibc/7B2A4F6E798182988D77B6B884919AF617A73503FDAC27C916CD7A69A69013CF"
)

var metadataIbc = banktypes.Metadata{
	Description: "ATOM IBC voucher (channel 14)",
	Base:        ibcBase,
	// NOTE: Denom units MUST be increasing
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    ibcBase,
			Exponent: 0,
		},
	},
	Name:    "ATOM channel-14",
	Symbol:  "ibcATOM-14",
	Display: ibcBase,
}

// setupRegisterERC20Pair deploys an ERC20 smart contract and
// registers it as ERC20.
func (suite *KeeperTestSuite) setupRegisterERC20Pair(contractType int) (common.Address, error) {
	var (
		contract common.Address
		err      error
	)
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contract, err = suite.DeployContractDirectBalanceManipulation()
	case contractMaliciousDelayed:
		contract, err = suite.DeployContractMaliciousDelayed()
	default:
		contract, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	}

	if err != nil {
		return common.Address{}, err
	}
	if err := suite.network.NextBlock(); err != nil {
		return common.Address{}, err
	}

	// submit gov proposal to register ERC20 token pair
	_, err = testutils.RegisterERC20(suite.factory, suite.network, testutils.ERC20RegistrationData{
		Addresses:    []string{contract.Hex()},
		ProposerPriv: suite.keyring.GetPrivKey(0),
	})

	return contract, err
}

func (suite *KeeperTestSuite) TestRegisterERC20() {
	var (
		ctx          sdk.Context
		contractAddr common.Address
		pair         types.TokenPair
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"token ERC20 already registered",
			func() {
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, pair.GetERC20Contract(), pair.GetID())
			},
			false,
		},
		{
			"denom already registered",
			func() {
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, pair.Denom, pair.GetID())
			},
			false,
		},
		{
			"meta data already stored",
			func() {
				suite.network.App.Erc20Keeper.CreateCoinMetadata(ctx, contractAddr) //nolint:errcheck
			},
			false,
		},
		{
			"ok",
			func() {},
			true,
		},
		{
			"force fail evm",
			func() {
				mockEVMKeeper := &erc20mocks.EVMKeeper{}

				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					suite.network.App.BankKeeper, mockEVMKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
				)

				mockEVMKeeper.On("EstimateGasInternal", mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced CallEVM error"))
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			var err error
			suite.SetupTest() // reset

			contractAddr, err = suite.factory.DeployContract(
				suite.keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{},
				testfactory.ContractDeploymentData{
					Contract:        contracts.ERC20MinterBurnerDecimalsContract,
					ConstructorArgs: []interface{}{erc20Name, erc20Symbol, cosmosDecimals},
				},
			)
			suite.Require().NoError(err, "failed to deploy contract")
			suite.Require().NoError(suite.network.NextBlock(), "failed to advance block")

			coinName := types.CreateDenom(contractAddr.String())
			pair = types.NewTokenPair(contractAddr, coinName, types.OWNER_EXTERNAL)

			ctx = suite.network.GetContext()

			tc.malleate()

			_, err = suite.network.App.Erc20Keeper.RegisterERC20(ctx, &types.MsgRegisterERC20{
				Authority:      authtypes.NewModuleAddress("gov").String(),
				Erc20Addresses: []string{contractAddr.Hex()},
			})
			metadata, found := suite.network.App.BankKeeper.GetDenomMetaData(ctx, coinName)
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
				suite.Require().Equal(zeroExponent, metadata.DenomUnits[0].Exponent)
				suite.Require().Equal(types.SanitizeERC20Name(erc20Name), metadata.DenomUnits[1].Denom)
				// Custom exponent at contract creation matches coin with token
				suite.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(cosmosDecimals))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestToggleConverision() {
	var (
		ctx          sdk.Context
		err          error
		contractAddr common.Address
		id           []byte
		pair         types.TokenPair
	)

	testCases := []struct {
		name              string
		malleate          func()
		expPass           bool
		conversionEnabled bool
	}{
		{
			"token not registered",
			func() {
				contractAddr, err = suite.factory.DeployContract(
					suite.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{
						Contract:        contracts.ERC20MinterBurnerDecimalsContract,
						ConstructorArgs: []interface{}{erc20Name, erc20Symbol, erc20Decimals},
					},
				)
				suite.Require().NoError(err, "failed to deploy contract")
				suite.Require().NoError(suite.network.NextBlock(), "failed to advance block")

				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, types.OWNER_MODULE)
			},
			false,
			false,
		},
		{
			"token not registered - pair not found",
			func() {
				contractAddr, err = suite.factory.DeployContract(
					suite.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{
						Contract:        contracts.ERC20MinterBurnerDecimalsContract,
						ConstructorArgs: []interface{}{erc20Name, erc20Symbol, erc20Decimals},
					},
				)
				suite.Require().NoError(err, "failed to deploy contract")
				suite.Require().NoError(suite.network.NextBlock(), "failed to advance block")

				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, types.OWNER_MODULE)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
			false,
		},
		{
			"disable conversion",
			func() {
				contractAddr, err = suite.setupRegisterERC20Pair(contractMinterBurner)
				suite.Require().NoError(err, "failed to register pair")
				ctx = suite.network.GetContext()
				id = suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
				pair, _ = suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
			},
			true,
			false,
		},
		{
			"disable and enable conversion",
			func() {
				contractAddr, err = suite.setupRegisterERC20Pair(contractMinterBurner)
				suite.Require().NoError(err, "failed to register pair")
				ctx = suite.network.GetContext()
				id = suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
				pair, _ = suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
				res, err := suite.network.App.Erc20Keeper.ToggleConversion(ctx, &types.MsgToggleConversion{Authority: authtypes.NewModuleAddress("gov").String(), Token: contractAddr.String()})
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				pair, _ = suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx = suite.network.GetContext()

			tc.malleate()

			_, err = suite.network.App.Erc20Keeper.ToggleConversion(ctx, &types.MsgToggleConversion{Authority: authtypes.NewModuleAddress("gov").String(), Token: contractAddr.String()})
			// Request the pair using the GetPairToken func to make sure that is updated on the db
			pair, _ = suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				if tc.conversionEnabled {
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
