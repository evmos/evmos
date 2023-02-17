package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v11/contracts"
	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/x/erc20/types"
)

// ensureHooksSet tries to set the hooks on EVMKeeper, this will fail if the erc20 hook is already set
func (suite *KeeperTestSuite) ensureHooksSet() {
	// TODO: PR to Ethermint to add the functionality `GetHooks` or `areHooksSet` to avoid catching a panic
	defer func() {
		err := recover()
		suite.Require().NotNil(err)
	}()
	suite.app.EvmKeeper.SetHooks(suite.app.Erc20Keeper.Hooks())
}

func (suite *KeeperTestSuite) TestEvmHooksRegisteredERC20() {
	testCases := []struct {
		name     string
		malleate func(common.Address)
		result   bool
	}{
		{
			"correct execution",
			func(contractAddr common.Address) {
				_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				// Mint 10 tokens to suite.address (owner)
				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				// Burn the 10 tokens of suite.address (owner)
				_ = suite.TransferERC20TokenToModule(contractAddr, suite.address, big.NewInt(10))
			},
			true,
		},
		{
			"unregistered pair",
			func(contractAddr common.Address) {
				// Mint 10 tokens to suite.address (owner)
				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				// Burn the 10 tokens of suite.address (owner)
				_ = suite.TransferERC20TokenToModule(contractAddr, suite.address, big.NewInt(10))
			},
			false,
		},
		{
			"wrong event",
			func(contractAddr common.Address) {
				_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				// Mint 10 tokens to suite.address (owner)
				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
			},
			false,
		},
		{
			"Pair is disabled",
			func(contractAddr common.Address) {
				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				pair.Enabled = false
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *pair)
				// Mint 10 tokens to suite.address (owner)
				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				// Burn the 10 tokens of suite.address (owner)
				_ = suite.TransferERC20TokenToModule(contractAddr, suite.address, big.NewInt(10))
			},
			false,
		},
		{
			"Pair is incorrectly loaded",
			func(contractAddr common.Address) {
				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				suite.app.Erc20Keeper.DeleteTokenPair(suite.ctx, *pair)

				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), pair.GetID())
				// Mint 10 tokens to suite.address (owner)
				_ = suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(10))
				suite.Commit()

				// Burn the 10 tokens of suite.address (owner)
				_ = suite.TransferERC20TokenToModule(contractAddr, suite.address, big.NewInt(10))
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()

			suite.ensureHooksSet()

			contractAddr, err := suite.DeployContract("coin test erc20", "token", erc20Decimals)
			suite.Require().NoError(err)

			tc.malleate(contractAddr)

			balance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(suite.address.Bytes()), types.CreateDenom(contractAddr.String()))
			suite.Commit()
			if tc.result {
				// Check if the execution was successful
				suite.Require().Equal(int64(10), balance.Amount.Int64())
			} else {
				// Check that no changes were made to the account
				suite.Require().Equal(int64(0), balance.Amount.Int64())
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestEvmHooksRegisteredCoin() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64

		result bool
	}{
		{"correct execution", 100, 10, 5, true},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()

			suite.ensureHooksSet()

			pair := suite.setupRegisterCoin(metadataCoin)
			suite.Require().NotNil(metadataCoin)
			suite.Require().NotNil(pair)

			sender := sdk.AccAddress(suite.address.Bytes())
			contractAddr := common.HexToAddress(pair.Erc20Address)

			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.mint)))
			err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.Require().NoError(err, tc.name)
			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			suite.Require().NoError(err, tc.name)

			convertCoin := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err = suite.app.Erc20Keeper.ConvertCoin(ctx, convertCoin)
			suite.Require().NoError(err, tc.name)
			suite.Commit()

			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadataCoin.Base)
			suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn).Int64())
			suite.Require().Equal(balance, big.NewInt(tc.burn))

			// Burn the 10 tokens of suite.address (owner)
			_ = suite.TransferERC20TokenToModule(contractAddr, suite.address, big.NewInt(tc.reconvert))

			balance = suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadataCoin.Base)

			if tc.result {
				suite.Require().Equal(balance, big.NewInt(tc.burn-tc.reconvert))
				// Check if the execution was successful
				suite.Require().NoError(err)
				suite.Require().Equal(cosmosBalance.Amount, sdk.NewInt(tc.mint-tc.burn+tc.reconvert))
			} else {
				// Check that no changes were made to the account
				suite.Require().Error(err)
				suite.Require().Equal(cosmosBalance.Amount, sdk.NewInt(tc.mint-tc.burn))
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestPostTxProcessing() {
	var (
		receipt *ethtypes.Receipt
		pair    *types.TokenPair
	)

	msg := ethtypes.NewMessage(
		types.ModuleAddress,
		&common.Address{},
		0,
		big.NewInt(0), // amount
		uint64(0),     // gasLimit
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		big.NewInt(0), // gasPrice
		[]byte{},
		ethtypes.AccessList{}, // AccessList
		true,                  // checkNonce
	)

	account := utiltx.GenerateAddress()

	transferData := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	transferData[31] = uint8(10)
	erc20 := contracts.ERC20BurnableContract.ABI

	transferEvent := erc20.Events["Transfer"]

	testCases := []struct {
		name          string
		malleate      func()
		expConversion bool
	}{
		{
			"Empty logs",
			func() {
				log := ethtypes.Log{}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"No log data",
			func() {
				topics := []common.Hash{transferEvent.ID, account.Hash(), types.ModuleAddress.Hash()}
				log := ethtypes.Log{
					Topics: topics,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"Non recognized event",
			func() {
				topics := []common.Hash{{}, account.Hash(), account.Hash()}
				log := ethtypes.Log{
					Topics: topics,
					Data:   transferData,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"Non transfer event",
			func() {
				aprovalEvent := erc20.Events["Approval"]
				topics := []common.Hash{aprovalEvent.ID, account.Hash(), account.Hash()}
				log := ethtypes.Log{
					Topics: topics,
					Data:   transferData,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"No log address",
			func() {
				topics := []common.Hash{transferEvent.ID, account.Hash(), types.ModuleAddress.Hash()}
				log := ethtypes.Log{
					Topics: topics,
					Data:   transferData,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"No data on topic",
			func() {
				topics := []common.Hash{transferEvent.ID}
				log := ethtypes.Log{
					Topics: topics,
					Data:   transferData,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"transfer to non-evm-module account",
			func() {
				contractAddr, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)

				_, err = suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				topics := []common.Hash{transferEvent.ID, account.Hash(), account.Hash()}
				log := ethtypes.Log{
					Topics:  topics,
					Data:    transferData,
					Address: contractAddr,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"correct burn",
			func() {
				contractAddr, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)

				pair, err = suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				topics := []common.Hash{transferEvent.ID, account.Hash(), types.ModuleAddress.Hash()}
				log := ethtypes.Log{
					Topics:  topics,
					Data:    transferData,
					Address: contractAddr,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			true,
		},
		{
			"Unspecified Owner",
			func() {
				contractAddr, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				pair.ContractOwner = types.OWNER_UNSPECIFIED
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *pair)

				topics := []common.Hash{transferEvent.ID, account.Hash(), types.ModuleAddress.Hash()}
				log := ethtypes.Log{
					Topics:  topics,
					Data:    transferData,
					Address: contractAddr,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
		{
			"Fail Evm",
			func() {
				contractAddr, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)

				pair, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
				suite.Require().NoError(err)

				pair.ContractOwner = types.OWNER_MODULE
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *pair)

				topics := []common.Hash{transferEvent.ID, account.Hash(), types.ModuleAddress.Hash()}
				log := ethtypes.Log{
					Topics:  topics,
					Data:    transferData,
					Address: contractAddr,
				}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			suite.ensureHooksSet()
			suite.Commit()

			tc.malleate()

			err := suite.app.Erc20Keeper.Hooks().PostTxProcessing(suite.ctx, msg, receipt)
			suite.Require().NoError(err)

			if tc.expConversion {
				sender := sdk.AccAddress(account.Bytes())
				cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, pair.Denom)

				transferEvent, err := erc20.Unpack("Transfer", transferData)
				suite.Require().NoError(err)

				tokens, _ := transferEvent[0].(*big.Int)
				suite.Require().Equal(cosmosBalance.Amount.String(), tokens.String())
			}
		})
	}
	suite.mintFeeCollector = false
}
