package demo

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/evmos/v16/integration_test_util"
	rpctypes "github.com/evmos/evmos/v16/rpc/types"
	etherminttypes "github.com/evmos/evmos/v16/types"
	"math/big"
)

//goland:noinspection SpellCheckingInspection

func (suite *EthRpcTestSuite) Test_Accounts() {
	suite.CITS.UseKeyring()
	suite.Commit() // ensure keyring is used

	accounts, err := suite.GetEthPublicAPI().Accounts()
	suite.Require().NoError(err)
	suite.Require().Len(accounts, len(suite.CITS.WalletAccounts))

	expectedAccounts := make(map[string]bool)
	for _, account := range suite.CITS.WalletAccounts {
		expectedAccounts[account.GetEthAddress().String()] = false
	}

	for _, account := range accounts {
		_, ok := expectedAccounts[account.String()]
		suite.Require().True(ok, "unexpected account %s", account.String())
		expectedAccounts[account.String()] = true
	}

	for account, ok := range expectedAccounts {
		suite.True(ok, "expected account %s not found", account)
	}
}

func (suite *EthRpcTestSuite) Test_GetBalance() {
	type historicalTestcase struct {
		rpctypes.BlockNumberOrHash
		originalBlockHeight int64
		expectedBalance     *big.Int
	}

	sender := suite.CITS.WalletAccounts.Number(1)
	receiver := suite.CITS.WalletAccounts.Number(2)

	suite.CITS.MintCoin(sender, suite.CITS.NewBaseCoin(100)) // prepare for multi txs

	var historicalTestcases []historicalTestcase

	for i := 0; i < 5; i++ {
		suite.CITS.WaitNextBlockOrCommit()

		beforeReceive := suite.CITS.QueryBalanceFromStore(0, receiver.GetCosmosAddress())

		err := suite.CITS.TxSend(sender, receiver, 1)
		suite.Require().NoError(err)

		suite.CITS.Commit() // ensure tx is included in block

		afterReceived := suite.CITS.QueryBalanceFromStore(0, receiver.GetCosmosAddress())

		suite.Require().False((*afterReceived).IsEqual(*beforeReceive), "receiver balance must be changed")

		currentHeight := suite.CITS.GetLatestBlockHeight()
		currentBlockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), nil)
		suite.Require().NoError(err)
		suite.Require().NotNil(currentBlockResult)

		blockNum := rpctypes.BlockNumber(currentHeight)
		blockHash := common.BytesToHash(currentBlockResult.Block.Hash().Bytes())
		historicalTestcases = append(historicalTestcases,
			historicalTestcase{
				originalBlockHeight: currentHeight,
				BlockNumberOrHash: rpctypes.BlockNumberOrHash{
					BlockNumber: &blockNum,
					BlockHash:   nil,
				},
				expectedBalance: afterReceived.Amount.BigInt(),
			}, historicalTestcase{
				originalBlockHeight: currentHeight,
				BlockNumberOrHash: rpctypes.BlockNumberOrHash{
					BlockNumber: nil,
					BlockHash:   &blockHash,
				},
				expectedBalance: afterReceived.Amount.BigInt(),
			},
		)
	}

	suite.Commit()

	finalBalance := suite.CITS.QueryBalanceFromStore(0, receiver.GetCosmosAddress())

	for _, tt := range historicalTestcases {
		if tt.BlockNumberOrHash.BlockNumber != nil {
			suite.Run(fmt.Sprintf("block height %d, GetBalance using BlockNumber", tt.originalBlockHeight), func() {
				gotBalance, err := suite.GetEthPublicAPIAt(tt.originalBlockHeight).GetBalance(receiver.GetEthAddress(), tt.BlockNumberOrHash)
				suite.Require().NoError(err)
				suite.Require().NotNil(gotBalance)

				suite.Zerof(tt.expectedBalance.Cmp(gotBalance.ToInt()), "expected balance %s at block %d, got %s", tt.expectedBalance.String(), tt.originalBlockHeight, gotBalance.ToInt().String())
			})
		}
		if tt.BlockNumberOrHash.BlockHash != nil {
			suite.Run(fmt.Sprintf("block height %d, GetBalance using BlockHash", tt.originalBlockHeight), func() {
				gotBalance, err := suite.GetEthPublicAPIAt(tt.originalBlockHeight).GetBalance(receiver.GetEthAddress(), tt.BlockNumberOrHash)
				suite.Require().NoError(err)
				suite.Require().NotNil(gotBalance)

				suite.Zerof(tt.expectedBalance.Cmp(gotBalance.ToInt()), "expected balance %s at block %d, got %s", tt.expectedBalance.String(), tt.originalBlockHeight, gotBalance.ToInt().String())
			})
		}
	}

	latestBlockNumber := rpctypes.EthLatestBlockNumber
	gotBalance, err := suite.GetEthPublicAPI().GetBalance(receiver.GetEthAddress(), rpctypes.BlockNumberOrHash{
		BlockNumber: &latestBlockNumber,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(gotBalance)
	suite.Zerof(finalBalance.Amount.BigInt().Cmp(gotBalance.ToInt()), "expected balance %s at latest, got %s", finalBalance.Amount.String(), gotBalance.ToInt().String())
}

func (suite *EthRpcTestSuite) Test_GetStorage() {
	type historicalTestcase struct {
		rpctypes.BlockNumberOrHash
		originalBlockHeight int64
		key                 string
		expectedValue       string
	}

	deployer := suite.CITS.WalletAccounts.Number(1)

	suite.CITS.MintCoin(deployer, suite.CITS.NewBaseCoin(100)) // prepare for multi txs

	contractAddr, _, _, err := suite.CITS.TxDeploy1StorageContract(deployer)
	suite.Require().NoError(err)

	storages := suite.App().EvmKeeper().GetAccountStorage(suite.Ctx(), contractAddr)
	suite.Require().Empty(storages, "new contract shouldn't have storage")

	var historicalTestcases []historicalTestcase

	for number := 5; number <= 10; number++ {
		suite.CITS.WaitNextBlockOrCommit()

		storagesBefore := suite.App().EvmKeeper().GetAccountStorage(suite.Ctx(), contractAddr)

		callData, err := integration_test_util.Contract1Storage.ABI.Pack("store", big.NewInt(int64(number)))
		suite.Require().NoError(err)
		_, _, err = suite.CITS.TxSendEvmTx(suite.Ctx(), deployer, &contractAddr, nil, callData)
		suite.Require().NoError(err)

		suite.CITS.Commit() // ensure tx is included in block

		storagesLater := suite.App().EvmKeeper().GetAccountStorage(suite.Ctx(), contractAddr)
		suite.Require().NotEmpty(storagesLater, "contract should have storage at this point")

		suite.Require().NotEqual(storagesBefore, storagesLater, "storage of contract must be changed")

		currentHeight := suite.CITS.GetLatestBlockHeight()
		currentBlockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), nil)
		suite.Require().NoError(err)
		suite.Require().NotNil(currentBlockResult)

		blockNum := rpctypes.BlockNumber(currentHeight)
		blockHash := common.BytesToHash(currentBlockResult.Block.Hash().Bytes())

		for _, storage := range storagesLater {
			historicalTestcases = append(historicalTestcases,
				historicalTestcase{
					originalBlockHeight: currentHeight,
					BlockNumberOrHash: rpctypes.BlockNumberOrHash{
						BlockNumber: &blockNum,
					},
					key:           storage.Key,
					expectedValue: storage.Value,
				},
				historicalTestcase{
					originalBlockHeight: currentHeight,
					BlockNumberOrHash: rpctypes.BlockNumberOrHash{
						BlockHash: &blockHash,
					},
					key:           storage.Key,
					expectedValue: storage.Value,
				},
			)
		}
	}

	suite.Commit()

	finalStorage := suite.App().EvmKeeper().GetAccountStorage(suite.Ctx(), contractAddr)

	for _, tt := range historicalTestcases {
		if tt.BlockNumberOrHash.BlockNumber != nil {
			suite.Run(fmt.Sprintf("block height %d, GetAccountStorage using BlockNumber", tt.originalBlockHeight), func() {
				gotStorageValue, err := suite.GetEthPublicAPIAt(tt.originalBlockHeight).GetStorageAt(contractAddr, tt.key, tt.BlockNumberOrHash)
				suite.Require().NoError(err)
				suite.Require().NotNil(gotStorageValue)

				suite.Equal(tt.expectedValue, gotStorageValue.String(), "expected value %s at block %d, got %s, key %s", tt.expectedValue, tt.originalBlockHeight, gotStorageValue.String(), tt.key)
			})
		}
		if tt.BlockNumberOrHash.BlockHash != nil {
			suite.Run(fmt.Sprintf("block height %d, GetAccountStorage using BlockHash", tt.originalBlockHeight), func() {
				gotStorageValue, err := suite.GetEthPublicAPIAt(tt.originalBlockHeight).GetStorageAt(contractAddr, tt.key, tt.BlockNumberOrHash)
				suite.Require().NoError(err)
				suite.Require().NotNil(gotStorageValue)

				suite.Equal(tt.expectedValue, gotStorageValue.String(), "expected value %s at block %d, got %s, key %s", tt.expectedValue, tt.originalBlockHeight, gotStorageValue.String(), tt.key)
			})
		}
	}

	latestBlockNumber := rpctypes.EthLatestBlockNumber
	for _, state := range finalStorage {
		gotStorageValue, err := suite.GetEthPublicAPI().GetStorageAt(contractAddr, state.GetKey(), rpctypes.BlockNumberOrHash{
			BlockNumber: &latestBlockNumber,
		})
		suite.Require().NoError(err)
		suite.Equal(state.GetValue(), gotStorageValue.String())
	}
}

func (suite *EthRpcTestSuite) Test_GetCode() {
	deployer := suite.CITS.WalletAccounts.Number(1)

	suite.CITS.MintCoin(deployer, suite.CITS.NewBaseCoin(100)) // prepare for multi txs

	suite.Commit()

	heightWithoutCode := suite.CITS.GetLatestBlockHeight()

	suite.Commit()
	suite.Commit()

	contractAddr, _, _, err := suite.CITS.TxDeploy1StorageContract(deployer)
	suite.Require().NoError(err)

	heightWithCode := suite.CITS.GetLatestBlockHeight()

	suite.Commit()

	accountI := suite.App().AccountKeeper().GetAccount(suite.Ctx(), contractAddr.Bytes())
	suite.Require().NotNil(accountI)
	contractAccount, ok := accountI.(*etherminttypes.EthAccount)
	suite.Require().True(ok)

	codeHash := common.HexToHash(contractAccount.CodeHash)

	code := suite.App().EvmKeeper().GetCode(suite.Ctx(), codeHash)
	suite.Require().NotEmptyf(code, "not found code for contract %s", contractAddr.String())

	blockWithoutCode, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(heightWithoutCode))
	suite.Require().NoError(err)
	suite.Require().NotNil(blockWithoutCode)

	blockWithCode, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(heightWithCode))
	suite.Require().NoError(err)
	suite.Require().NotNil(blockWithCode)

	suite.Run("context contract not deployed, fetch by block number", func() {
		blockNumberWithoutCode := rpctypes.BlockNumber(heightWithoutCode)
		bz, err := suite.GetEthPublicAPIAt(heightWithoutCode).GetCode(contractAddr, rpctypes.BlockNumberOrHash{
			BlockNumber: &blockNumberWithoutCode,
		})
		suite.Require().NoError(err)
		suite.Empty(bz, "code must be empty at this context")
	})

	suite.Run("context contract not deployed, fetch by block hash", func() {
		blockHashWithoutCode := common.BytesToHash(blockWithoutCode.Block.Hash().Bytes())
		bz, err := suite.GetEthPublicAPIAt(heightWithoutCode).GetCode(contractAddr, rpctypes.BlockNumberOrHash{
			BlockHash: &blockHashWithoutCode,
		})
		suite.Require().NoError(err)
		suite.Empty(bz, "code must be empty at this context")
	})

	suite.Run("context contract deployed, fetch by block number", func() {
		blockNumberWithCode := rpctypes.BlockNumber(heightWithCode)
		bz, err := suite.GetEthPublicAPIAt(heightWithCode).GetCode(contractAddr, rpctypes.BlockNumberOrHash{
			BlockNumber: &blockNumberWithCode,
		})
		suite.Require().NoError(err)
		suite.Equal(hexutil.Bytes(code), bz)
	})

	suite.Run("context contract deployed, fetch by block hash", func() {
		blockHashWithCode := common.BytesToHash(blockWithCode.Block.Hash().Bytes())
		bz, err := suite.GetEthPublicAPIAt(heightWithCode).GetCode(contractAddr, rpctypes.BlockNumberOrHash{
			BlockHash: &blockHashWithCode,
		})
		suite.Require().NoError(err)
		suite.Equal(hexutil.Bytes(code), bz)
	})
}
