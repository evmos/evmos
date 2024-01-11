package demo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	tmrpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/evmos/evmos/v16/integration_test_util"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	rpctypes "github.com/evmos/evmos/v16/rpc/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

//goland:noinspection SpellCheckingInspection

type resultGetBlockStruct struct {
	BaseFeePerGas    string        `json:"baseFeePerGas"`
	Difficulty       string        `json:"difficulty"`
	ExtraData        string        `json:"extraData"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	Hash             string        `json:"hash"`
	LogsBloom        string        `json:"logsBloom"`
	Miner            string        `json:"miner"`
	MixHash          string        `json:"mixHash"`
	Nonce            string        `json:"nonce"`
	Number           string        `json:"number"`
	ParentHash       string        `json:"parentHash"`
	ReceiptsRoot     string        `json:"receiptsRoot"`
	Sha3Uncles       string        `json:"sha3Uncles"`
	Size             string        `json:"size"`
	StateRoot        string        `json:"stateRoot"`
	Timestamp        string        `json:"timestamp"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	Transactions     []interface{} `json:"transactions"`
	TransactionsRoot string        `json:"transactionsRoot"`
	Uncles           []string      `json:"uncles"`
}

func (suite *EthRpcTestSuite) Test_BlockNumber() {
	randomShiftingBlocksCount := int(rand.Uint32()%5 + 2)

	for i := 0; i < randomShiftingBlocksCount; i++ {
		suite.Commit()
	}

	latestBlockHeight := suite.CITS.GetLatestBlockHeight()
	suite.Require().Equal(latestBlockHeight, suite.Ctx().BlockHeight())

	blockNumber, err := suite.GetEthPublicAPI().BlockNumber()
	suite.Require().NoError(err)

	suite.Equal(uint64(latestBlockHeight), uint64(blockNumber))
}

func (suite *EthRpcTestSuite) Test_GetBlockByNumberAndHash() {
	suite.Run("basic test", func() {
		suite.Commit() // require at least 2 blocks

		previousBlockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(1))
		suite.Require().NoError(err)
		suite.Require().NotNil(previousBlockResult)

		gotBlockByNumber, err := suite.GetEthPublicAPI().GetBlockByNumber(2, false)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByNumber)

		currentBlockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(2))
		suite.Require().NoError(err)
		suite.Require().NotNil(currentBlockResult)
		suite.Require().Equal(int64(2), currentBlockResult.Block.Height)

		gotBlockByHash, err := suite.GetEthPublicAPI().GetBlockByHash(common.BytesToHash(currentBlockResult.Block.Hash()), false)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByHash)

		suite.Equal(gotBlockByNumber, gotBlockByHash, "result of eth_getBlockByNumber and eth_getBlockByHash must be same")

		suite.Equal(hexutil.Uint64(currentBlockResult.Block.Height), gotBlockByNumber["number"])
		suite.Equal(hexutil.Bytes(currentBlockResult.Block.Hash()), gotBlockByNumber["hash"], "hash must be Tendermint block hash")
		suite.Equal(common.BytesToHash(previousBlockResult.Block.Hash()), gotBlockByNumber["parentHash"], "parentHash must be previous Tendermint block hash")
		suite.Equal(hexutil.Bytes(currentBlockResult.Block.AppHash), gotBlockByNumber["stateRoot"], "stateRoot must be Tendermint AppHash")
		suite.Equal([]common.Hash{}, gotBlockByNumber["uncles"], "uncles must be empty since it is not possible in PoS chain")
	})

	// this is response-based testing so the test is json based, that's why it is not usually unmarshal to object.
	// The ultimate target is ensured response data to end user.
	deepTestGetBlockByNumberAndHash := func(fullTxs bool) {
		var err error

		// shift some blocks
		randomShiftingBlocksCount := int(rand.Uint32()%3 + 3)
		for i := 0; i < randomShiftingBlocksCount; i++ {
			suite.Commit()
		}

		// prepare txs
		const evmTxsCount = 2
		const nonEvmTxsCount = 1
		var senderEvmTxs, senderNonEvmTxs []*itutiltypes.TestAccount
		// prepare senders and fund them
		for num := 1; num <= evmTxsCount; num++ {
			sender := integration_test_util.NewTestAccount(suite.T(), nil)
			suite.CITS.MintCoin(sender, suite.CITS.NewBaseCoin(10))
			senderEvmTxs = append(senderEvmTxs, sender)
		}
		for num := 1; num <= nonEvmTxsCount; num++ {
			sender := integration_test_util.NewTestAccount(suite.T(), nil)
			suite.CITS.MintCoin(sender, suite.CITS.NewBaseCoin(10))
			senderNonEvmTxs = append(senderNonEvmTxs, sender)
		}

		// wait new block then send some txs to ensure all txs are included in the same block
		suite.CITS.WaitNextBlockOrCommit()

		testBlockHeight := suite.CITS.GetLatestBlockHeight()

		receiver := suite.CITS.WalletAccounts.Number(1)

		msgEvmTxs := make(map[string]*evmtypes.MsgEthereumTx)

		startTime := time.Now().UTC().UnixMilli()
		for num := 1; num <= evmTxsCount; num++ {
			// Txs must be sent async to ensure same block height

			msgEthereumTx, err := suite.CITS.TxSendViaEVMAsync(senderEvmTxs[num-1], receiver, 1)
			suite.Require().NoError(err, "failed to send tx to create test data")

			msgEvmTxs[msgEthereumTx.Hash] = msgEthereumTx
		}

		for num := 1; num <= nonEvmTxsCount; num++ {
			// Txs must be sent async to ensure same block height
			err = suite.CITS.TxSendAsync(senderNonEvmTxs[num-1], receiver, 1)
			suite.Require().NoError(err, "failed to send tx to create test data")
		}
		fmt.Println("Broadcast takes", time.Now().UTC().UnixMilli()-startTime, "ms")

		suite.CITS.WaitNextBlockOrCommit() // finalize the test block

		if testBlockHeight+1 != suite.CITS.GetLatestBlockHeight() {
			suite.T().Skip("test skipped because the expected context block number does not matches")
		}

		testBlockHeight++ // since txs go to mempool and only included in the next block

		fmt.Println("testBlockHeight", testBlockHeight)

		suite.CITS.Commit() // commit to passive trigger EVM Tx indexer

		balance := suite.CITS.QueryBalance(0, receiver.GetCosmosAddress().String())
		suite.Require().False(balance.IsZero(), "receiver must received some balance")

		previousBlockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(testBlockHeight-1))
		suite.Require().NoError(err)
		suite.Require().NotNil(previousBlockResult)

		gotBlockByNumber, err := suite.GetEthPublicAPI().GetBlockByNumber(rpctypes.BlockNumber(testBlockHeight), fullTxs)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByNumber)

		blockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(testBlockHeight))
		suite.Require().NoError(err)
		suite.Require().NotNil(blockResult)
		suite.Require().Equal(evmTxsCount+nonEvmTxsCount, len(blockResult.Block.Txs), "must be same as sent txs count for both EVM & non-EVM txs")

		gotBlockByHash, err := suite.GetEthPublicAPI().GetBlockByHash(common.BytesToHash(blockResult.Block.Hash()), fullTxs)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByHash)

		suite.Equal(gotBlockByNumber, gotBlockByHash, "result of eth_getBlockByNumber and eth_getBlockByHash must be same")

		resultBlockResult, err := suite.CITS.RpcBackend.TendermintBlockResultByNumber(ptrInt64(testBlockHeight))
		suite.Require().NoError(err)
		blockBloom, err := suite.CITS.RpcBackend.BlockBloom(resultBlockResult)
		suite.Require().NoError(err, "failed to fetch block bloom")

		baseFee := suite.App().FeeMarketKeeper().GetParams(suite.Ctx()).BaseFee
		consensusParams, err := suite.CITS.QueryClients.ClientQueryCtx.Client.(tmrpcclient.NetworkClient).ConsensusParams(context.Background(), ptrInt64(testBlockHeight))
		suite.Require().NoError(err, "failed to fetch consensus params of test block")
		suite.Equal(int64(40_000_000), consensusParams.ConsensusParams.Block.MaxGas, "invalid setup?")

		bzGotBlockByNumber, err := json.Marshal(gotBlockByNumber)
		suite.Require().NoError(err)

		var textResultStruct resultGetBlockStruct
		err = json.Unmarshal(bzGotBlockByNumber, &textResultStruct)
		suite.Require().NoError(err)

		if len(textResultStruct.Transactions) > 0 {
			if fullTxs {
				for i, mTx := range textResultStruct.Transactions {
					txData, ok := mTx.(map[string]interface{})
					suite.True(ok, "when full-txs mode, tx list must be the tx data itself")

					txHash, ok := txData["hash"].(string)
					suite.True(ok, "invalid tx content")

					msgEvmTx, ok := msgEvmTxs[txHash]
					suite.Truef(ok, "tx %s could not be found", txHash)

					bz, err := json.Marshal(txData)
					suite.Require().NoError(err)

					var ethRpcTxs rpctypes.RPCTransaction
					err = json.Unmarshal(bz, &ethRpcTxs)
					suite.Require().NoError(err, "failed to unmarshal to RPCTransaction")

					tx := msgEvmTx.AsTransaction()

					if suite.NotNil(txData["blockHash"]) {
						if suite.IsType("string", txData["blockHash"]) {
							suite.Equal("0x"+strings.ToLower(blockResult.Block.Hash().String()), txData["blockHash"])
						}
					}
					if suite.NotNil(txData["blockNumber"]) {
						if suite.IsType("string", txData["blockNumber"]) {
							suite.Equal(fmt.Sprintf("0x%x", blockResult.Block.Height), txData["blockNumber"])
						}
					}
					if suite.NotNil(txData["from"]) {
						if suite.IsType("string", txData["from"]) {
							suite.Len(txData["from"].(string), 42)
							suite.NotEqual("0x0000000000000000000000000000000000000000", txData["from"])
						}
					}
					if suite.NotNil(txData["gas"]) {
						if suite.IsType("string", txData["gas"]) {
							suite.Equal(fmt.Sprintf("0x%x", tx.Gas()), txData["gas"])
						}
					}
					if suite.NotNil(txData["gasPrice"]) {
						if suite.IsType("string", txData["gasPrice"]) {
							suite.Contains(txData["gasPrice"].(string), "0x")
						}
					}
					if suite.NotNil(txData["hash"]) {
						if suite.IsType("string", txData["hash"]) {
							suite.Equal(strings.ToLower(tx.Hash().String()), txData["hash"])
						}
					}
					if suite.NotNil(txData["to"]) {
						if suite.IsType("string", txData["to"]) {
							suite.Equal(strings.ToLower(receiver.GetEthAddress().String()), txData["to"])
						}
					}
					if suite.NotNil(txData["nonce"]) {
						if suite.IsType("string", txData["nonce"]) {
							suite.Equal("0x0", txData["nonce"])
						}
					}
					if suite.NotNil(txData["input"]) {
						if suite.IsType("string", txData["input"]) {
							suite.Contains(txData["input"].(string), "0x")
							suite.Equal("0x", txData["input"])
						}
					}
					if suite.NotNil(txData["transactionIndex"]) {
						if suite.IsType("string", txData["transactionIndex"]) {
							suite.Equal(fmt.Sprintf("0x%x", i), txData["transactionIndex"])
						}
					}
					if suite.NotNil(txData["value"]) {
						if suite.IsType("string", txData["value"]) {
							suite.Equal(fmt.Sprintf("0x%x", suite.CITS.NewBaseCoin(1).Amount.Int64()), txData["value"])
						}
					}
					if suite.NotNil(txData["type"]) {
						if suite.IsType("string", txData["type"]) {
							suite.Equal(fmt.Sprintf("0x%x", tx.Type()), txData["type"])
						}
					}
					if accessList, found := txData["accessList"]; found {
						suite.Empty(accessList)
					}
					if suite.NotNil(txData["chainId"]) {
						if suite.IsType("string", txData["chainId"]) {
							suite.Equal(fmt.Sprintf("0x%x", suite.App().EvmKeeper().ChainID().Int64()), txData["chainId"])
						}
					}
					v, r, s := tx.RawSignatureValues()
					if suite.NotNil(txData["v"]) {
						if suite.IsType("string", txData["v"]) {
							if v.Sign() == 0 {
								suite.Equal("0x0", txData["v"])
							} else {
								suite.Equal(fmt.Sprintf("0x%x", v), txData["v"])
							}
						}
					}
					if suite.NotNil(txData["r"]) {
						if suite.IsType("string", txData["r"]) {
							if r.Sign() == 0 {
								suite.Equal("0x0", txData["r"])
							} else {
								suite.Equal(fmt.Sprintf("0x%x", r), txData["r"])
							}
						}
					}
					if suite.NotNil(txData["s"]) {
						if suite.IsType("string", txData["s"]) {
							if s.Sign() == 0 {
								suite.Equal("0x0", txData["s"])
							} else {
								suite.Equal(fmt.Sprintf("0x%x", s), txData["s"])
							}
						}
					}
				}
			} else {
				for _, tx := range textResultStruct.Transactions {
					txHash, ok := tx.(string)
					suite.True(ok, "when Not full-txs mode, tx list must be tx hash")
					_, ok = msgEvmTxs[txHash]
					suite.Truef(ok, "tx %s could not be found", txHash)
				}
			}
		}

		suite.Equal(fmt.Sprintf("0x%x", baseFee.Int64()), textResultStruct.BaseFeePerGas)
		suite.Equal("0x0", textResultStruct.Difficulty, "difficulty must be zero since PoS chain does not have this")
		suite.Equal("0x", textResultStruct.ExtraData)
		suite.Equal(fmt.Sprintf("0x%x", consensusParams.ConsensusParams.Block.MaxGas), textResultStruct.GasLimit)
		suite.NotEqual("0x0", textResultStruct.GasUsed, "gasUsed must not be zero since there are some txs")
		suite.Equal("0x"+hex.EncodeToString(blockResult.Block.Hash()), textResultStruct.Hash, "hash must be Tendermint block hash")
		suite.Equal(fmt.Sprintf("0x%x", blockBloom.Bytes()), textResultStruct.LogsBloom)
		suite.Equal(strings.ToLower(suite.CITS.ValidatorAccounts.Number(1).GetEthAddress().String()), textResultStruct.Miner, "mis-match validator address as miner or must be lower-case") // Tendermint node uses the first pre-defined validator
		suite.Equal("0x0000000000000000000000000000000000000000000000000000000000000000", textResultStruct.MixHash, "mixHash must be zero since PoS chain does not have this")
		suite.Equal("0x0000000000000000", textResultStruct.Nonce, "nonce must be zero since PoS chain does not have this")
		suite.Equal(fmt.Sprintf("0x%x", testBlockHeight), textResultStruct.Number)
		suite.Equal("0x"+hex.EncodeToString(previousBlockResult.Block.Hash()), textResultStruct.ParentHash, "parentHash must be previous Tendermint block hash")
		suite.Equal(func() string { // TODO ES fix the RPC to return correct receipt root
			var receipts ethtypes.Receipts
			for _, tx := range textResultStruct.Transactions {
				var transaction *ethtypes.Transaction
				if fullTxs {
					mTx, ok := tx.(map[string]interface{})
					suite.Require().True(ok)
					txHash, ok := mTx["hash"].(string)
					suite.Require().True(ok)
					transaction = msgEvmTxs[txHash].AsTransaction()
				} else {
					txHash, ok := tx.(string)
					suite.Require().True(ok)
					transaction = msgEvmTxs[txHash].AsTransaction()
				}
				receipts = append(receipts, suite.GetTxReceipt(transaction.Hash()))
			}

			return ethtypes.DeriveSha(receipts, trie.NewStackTrie(nil)).String()
		}(), textResultStruct.ReceiptsRoot, "mis-match receipt root")
		suite.Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", textResultStruct.Sha3Uncles, "sha3Uncles must be value of EmptyUncleHash")
		suite.Equal(fmt.Sprintf("0x%x", blockResult.Block.Size()), textResultStruct.Size, "must must be Tendermint block size")
		suite.Equal("0x"+hex.EncodeToString(blockResult.Block.AppHash.Bytes()), textResultStruct.StateRoot, "stateRoot must be Tendermint AppHash")
		suite.Equal(fmt.Sprintf("0x%x", blockResult.Block.Time.UTC().Unix()), textResultStruct.Timestamp, "timestamp must be block UTC epoch seconds")
		suite.Equal("0x0", textResultStruct.TotalDifficulty, "total difficulty must be zero since PoS chain does not have this")
		suite.Len(textResultStruct.Transactions, evmTxsCount, "transaction list must be same as sent EVM txs")
		suite.Equal(func() string { // TODO ES fix the RPC to return correct transaction root
			var transactions ethtypes.Transactions
			for _, tx := range textResultStruct.Transactions {
				var transaction *ethtypes.Transaction
				if fullTxs {
					mTx, ok := tx.(map[string]interface{})
					suite.Require().True(ok)
					txHash, ok := mTx["hash"].(string)
					suite.Require().True(ok)
					transaction = msgEvmTxs[txHash].AsTransaction()
				} else {
					txHash, ok := tx.(string)
					suite.Require().True(ok)
					transaction = msgEvmTxs[txHash].AsTransaction()
				}
				transactions = append(transactions, transaction)
			}

			return ethtypes.DeriveSha(transactions, trie.NewStackTrie(nil)).String()
		}(), textResultStruct.TransactionsRoot, "mis-match transaction root")
		suite.Empty(textResultStruct.Uncles, "uncles must be empty since it is not possible in PoS chain")
	}

	suite.Run("deep test, not full txs", func() {
		deepTestGetBlockByNumberAndHash(false)
	})

	suite.Run("deep test, full txs", func() {
		deepTestGetBlockByNumberAndHash(true)
	})

	suite.Run("test access list & nonce of txs in full-tx mode", func() {
		sender := suite.CITS.WalletAccounts.Number(1)
		receiver := suite.CITS.WalletAccounts.Number(2)

		suite.CITS.MintCoin(sender, suite.CITS.NewBaseCoin(100))

		evmKeeper := suite.App().EvmKeeper()
		nonceSender := evmKeeper.GetNonce(suite.Ctx(), sender.GetEthAddress())

		suite.Commit()

		err := suite.CITS.TxSend(sender, receiver, 1)
		suite.Require().NoError(err)
		nonceSender++

		suite.Commit()

		evmTxArgs := &evmtypes.EvmTxArgs{
			ChainID:   evmKeeper.ChainID(),
			Nonce:     nonceSender,
			GasLimit:  300_000,
			GasFeeCap: suite.App().FeeMarketKeeper().GetBaseFee(suite.Ctx()),
			GasTipCap: big.NewInt(1),
			To: func() *common.Address {
				ethAddr := receiver.GetEthAddress()
				return &ethAddr
			}(),
			Amount: suite.CITS.NewBaseCoin(1).Amount.BigInt(),
			Input:  nil,
			Accesses: &ethtypes.AccessList{
				{
					Address: sender.GetEthAddress(),
				},
				{
					Address: receiver.GetEthAddress(),
				},
			},
		}
		suite.Require().Len(*evmTxArgs.Accesses, 2)

		msgEthereumTx := evmtypes.NewTx(evmTxArgs)
		msgEthereumTx.From = sender.GetEthAddress().String()

		_, err = suite.CITS.DeliverEthTx(sender, msgEthereumTx)
		suite.Require().NoError(err)

		suite.Commit() // trigger EVM Tx indexer to index block

		txByHash, err := suite.CITS.RpcBackend.GetTransactionByHash(msgEthereumTx.AsTransaction().Hash())
		suite.Require().NoError(err)
		suite.Require().NotNil(txByHash, "failed to find tx by hash")
		suite.Require().NotNil(txByHash.BlockNumber)
		suite.Require().NotNil(txByHash.BlockHash)

		gotBlockByNumber, err := suite.GetEthPublicAPI().GetBlockByNumber(rpctypes.BlockNumber(txByHash.BlockNumber.ToInt().Int64()), true)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByNumber)

		gotBlockByHash, err := suite.GetEthPublicAPI().GetBlockByHash(*txByHash.BlockHash, true)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByHash)

		suite.Equal(gotBlockByNumber, gotBlockByHash, "result of eth_getBlockByNumber and eth_getBlockByHash must be same")

		bzGotBlockByNumber, err := json.Marshal(gotBlockByNumber)
		suite.Require().NoError(err)

		var textResultStruct resultGetBlockStruct
		err = json.Unmarshal(bzGotBlockByNumber, &textResultStruct)
		suite.Require().NoError(err)

		suite.Require().NotNil(textResultStruct.Transactions)
		suite.Require().Len(textResultStruct.Transactions, 1)

		txData, ok := textResultStruct.Transactions[0].(map[string]interface{})
		suite.True(ok)

		bz, err := json.Marshal(txData)
		suite.Require().NoError(err)

		var ethRpcTxs rpctypes.RPCTransaction
		err = json.Unmarshal(bz, &ethRpcTxs)
		suite.Require().NoError(err, "failed to unmarshal to RPCTransaction")

		if suite.NotNil(txData["from"]) {
			if suite.IsType("string", txData["from"]) {
				suite.Equal(strings.ToLower(sender.GetEthAddress().String()), txData["from"])
			}
		}

		if suite.NotNil(txData["to"]) {
			if suite.IsType("string", txData["to"]) {
				suite.Equal(strings.ToLower(receiver.GetEthAddress().String()), txData["to"])
			}
		}

		if suite.NotNil(txData["nonce"]) {
			if suite.IsType("string", txData["nonce"]) {
				suite.Equal(fmt.Sprintf("0x%x", evmTxArgs.Nonce), txData["nonce"])
			}
		}

		if suite.NotNil(txData["accessList"]) {
			suite.Len(txData["accessList"].([]interface{}), 2)
		}
	})

	suite.Run("test input data of txs in full-txs mode", func() {
		deployer := suite.CITS.WalletAccounts.Number(1)

		evmKeeper := suite.App().EvmKeeper()
		nonce := evmKeeper.GetNonce(suite.Ctx(), deployer.GetEthAddress())

		_, evmTxsMsg, _, err := suite.CITS.TxDeploy1StorageContract(deployer)
		suite.Require().NoError(err)
		nonce++

		suite.Commit() // trigger EVM Tx indexer to index block

		txByHash, err := suite.CITS.RpcBackend.GetTransactionByHash(evmTxsMsg.AsTransaction().Hash())
		suite.Require().NoError(err)
		suite.Require().NotNil(txByHash, "failed to find tx by hash")
		suite.Require().NotNil(txByHash.BlockNumber)
		suite.Require().NotNil(txByHash.BlockHash)

		gotBlockByNumber, err := suite.GetEthPublicAPI().GetBlockByNumber(rpctypes.BlockNumber(txByHash.BlockNumber.ToInt().Int64()), true)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByNumber)

		gotBlockByHash, err := suite.GetEthPublicAPI().GetBlockByHash(*txByHash.BlockHash, true)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByHash)

		suite.Equal(gotBlockByNumber, gotBlockByHash, "result of eth_getBlockByNumber and eth_getBlockByHash must be same")

		bzGotBlockByNumber, err := json.Marshal(gotBlockByNumber)
		suite.Require().NoError(err)

		var textResultStruct resultGetBlockStruct
		err = json.Unmarshal(bzGotBlockByNumber, &textResultStruct)
		suite.Require().NoError(err)

		suite.Require().NotNil(textResultStruct.Transactions)
		suite.Require().Len(textResultStruct.Transactions, 1)

		txData, ok := textResultStruct.Transactions[0].(map[string]interface{})
		suite.True(ok)

		bz, err := json.Marshal(txData)
		suite.Require().NoError(err)

		var ethRpcTxs rpctypes.RPCTransaction
		err = json.Unmarshal(bz, &ethRpcTxs)
		suite.Require().NoError(err, "failed to unmarshal to RPCTransaction")

		if suite.NotNil(txData["from"]) {
			if suite.IsType("string", txData["from"]) {
				suite.Equal(strings.ToLower(deployer.GetEthAddress().String()), txData["from"])
			}
		}

		suite.Nil(txData["to"]) // it is contract deployment

		if suite.NotNil(txData["input"]) {
			if suite.IsType("string", txData["input"]) {
				suite.Equal("0x"+strings.ToLower(hex.EncodeToString(evmTxsMsg.AsTransaction().Data())), txData["input"])
			}
		}
	})

	suite.Run("txs index must be unique and ordered ascending in EVM block", func() {
		suite.Commit()

		receiver := integration_test_util.NewTestAccount(suite.T(), nil)

		var allSenders []*itutiltypes.TestAccount
		var msgEvmTxs []*evmtypes.MsgEthereumTx
		var evmTxSender []*itutiltypes.TestAccount

		for n := 1; n <= 6; n++ {
			sender := integration_test_util.NewTestAccount(suite.T(), nil)
			suite.CITS.MintCoin(sender, suite.CITS.NewBaseCoin(10))
			allSenders = append(allSenders, sender)
		}

		// wait new block then send some txs to ensure all txs are included in the same block
		suite.CITS.WaitNextBlockOrCommit()

		actionBlockHeight := suite.CITS.GetLatestBlockHeight()

		for i, sender := range allSenders {
			// create interleaved transactions Evm => Cosmos => Evm => Cosmos => ...

			if i%2 == 0 {
				// Txs must be sent async to ensure same block height
				msgEthereumTx, err := suite.CITS.TxSendViaEVMAsync(sender, receiver, 1)
				suite.Require().NoError(err, "failed to send tx to create test data")

				msgEvmTxs = append(msgEvmTxs, msgEthereumTx)
				evmTxSender = append(evmTxSender, sender)
			} else {
				// Txs must be sent async to ensure same block height
				err := suite.CITS.TxSendAsync(sender, receiver, 1) // bank sent
				suite.Require().NoError(err, "failed to send tx to create test data")
			}
		}

		suite.CITS.WaitNextBlockOrCommit() // finalize the test block

		suite.Require().Equal(actionBlockHeight+1, suite.CITS.GetLatestBlockHeight(), "be one block later")

		suite.CITS.Commit() // commit to passive trigger EVM Tx indexer

		balance := suite.CITS.QueryBalance(0, receiver.GetCosmosAddress().String())
		suite.Require().False(balance.IsZero(), "receiver must received some balance")

		var uniqueBlockNumber int64

		for _, sentEvmTx := range msgEvmTxs {
			sentTxHash := sentEvmTx.AsTransaction().Hash()
			txByHash, err := suite.GetEthPublicAPI().GetTransactionByHash(sentTxHash)
			suite.Require().NoError(err)
			suite.Require().NotNil(txByHash)
			suite.Equal(sentTxHash, txByHash.Hash)
			if suite.NotNil(txByHash.BlockHash) {
				suite.Equal(1, txByHash.BlockHash.Big().Sign()) // positive
			}
			if suite.NotNil(txByHash.BlockNumber) {
				if suite.Equal(1, txByHash.BlockNumber.ToInt().Sign()) { // positive
					blockNumber := txByHash.BlockNumber.ToInt().Int64()
					if uniqueBlockNumber == 0 {
						uniqueBlockNumber = blockNumber
					} else {
						suite.Require().Equal(uniqueBlockNumber, blockNumber, "expected all test txs must be in the same block")
					}
				}
			}
		}

		gotBlockByNumber, err := suite.GetEthPublicAPI().GetBlockByNumber(rpctypes.BlockNumber(uniqueBlockNumber), true)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByNumber)

		bzGotBlockByNumber, err := json.Marshal(gotBlockByNumber)
		suite.Require().NoError(err)

		var textResultStruct resultGetBlockStruct
		err = json.Unmarshal(bzGotBlockByNumber, &textResultStruct)
		suite.Require().NoError(err)

		gotBlockByHash, err := suite.GetEthPublicAPI().GetBlockByHash(common.HexToHash(textResultStruct.Hash), true)
		suite.Require().NoError(err)
		suite.Require().NotNil(gotBlockByHash)

		suite.Equal(gotBlockByNumber, gotBlockByHash, "result of eth_getBlockByNumber and eth_getBlockByHash must be same")

		suite.Require().NotNil(textResultStruct.Transactions)
		suite.Require().Len(textResultStruct.Transactions, len(msgEvmTxs), "must be same as sent EVM txs")

		txIndexTracker := make([]bool, len(msgEvmTxs))

		for _, tx := range textResultStruct.Transactions {
			txData, ok := tx.(map[string]interface{})
			suite.Require().True(ok, "when full-txs mode, tx list must be the tx data itself")

			bz, err := json.Marshal(txData)
			suite.Require().NoError(err)

			var ethRpcTxs rpctypes.RPCTransaction
			err = json.Unmarshal(bz, &ethRpcTxs)
			suite.Require().NoError(err, "failed to unmarshal to RPCTransaction")

			if suite.NotNil(txData["transactionIndex"]) {
				if suite.IsType("string", txData["transactionIndex"]) {
					txIndex, err := strconv.ParseInt(strings.TrimPrefix(txData["transactionIndex"].(string), "0x"), 16, 64)
					suite.Require().NoError(err)

					reserved := txIndexTracker[int(txIndex)]
					if reserved {
						suite.Failf("tx index must be unique", "tx index %d is already reserved", txIndex)
					} else {
						txIndexTracker[txIndex] = true
					}
				}
			}
		}

		for i, reserved := range txIndexTracker {
			if !reserved {
				suite.Failf("lacking tx tracker", "where is tx index %d?", i)
			}
		}
	})
}

func (suite *EthRpcTestSuite) Test_GetBlockTransactionCountByNumberAndHash() {
	var err error

	// shift some blocks
	randomShiftingBlocksCount := int(rand.Uint32()%3 + 3)
	for i := 0; i < randomShiftingBlocksCount; i++ {
		suite.Commit()
	}

	// prepare txs
	var nonEvmTxsCount = 1
	var evmTxsCount = len(suite.CITS.WalletAccounts) - nonEvmTxsCount
	var senderEvmTxs, senderNonEvmTxs []*itutiltypes.TestAccount

	// prepare senders
	for _, sender := range suite.CITS.WalletAccounts {
		if len(senderEvmTxs) < evmTxsCount {
			senderEvmTxs = append(senderEvmTxs, sender)
		} else {
			senderNonEvmTxs = append(senderNonEvmTxs, sender)
		}
	}

	// wait new block then send some txs to ensure all txs are included in the same block
	suite.CITS.WaitNextBlockOrCommit()

	testBlockHeight := suite.CITS.GetLatestBlockHeight()

	receiver := integration_test_util.NewTestAccount(suite.T(), nil)

	msgEvmTxs := make(map[string]*evmtypes.MsgEthereumTx)

	startTime := time.Now().UTC().UnixMilli()
	for num := 1; num <= evmTxsCount; num++ {
		// Txs must be sent async to ensure same block height

		msgEthereumTx, err := suite.CITS.TxSendViaEVMAsync(senderEvmTxs[num-1], receiver, 1)
		suite.Require().NoError(err, "failed to send tx to create test data")

		msgEvmTxs[msgEthereumTx.Hash] = msgEthereumTx
	}

	for num := 1; num <= nonEvmTxsCount; num++ {
		// Txs must be sent async to ensure same block height
		err = suite.CITS.TxSendAsync(senderNonEvmTxs[num-1], receiver, 1)
		suite.Require().NoError(err, "failed to send tx to create test data")
	}
	fmt.Println("Broadcast takes", time.Now().UTC().UnixMilli()-startTime, "ms")

	suite.CITS.WaitNextBlockOrCommit() // finalize the test block

	if testBlockHeight+1 != suite.CITS.GetLatestBlockHeight() {
		suite.T().Skip("test skipped because the expected context block number does not matches")
	}

	testBlockHeight++ // since txs go to mempool and only included in the next block

	fmt.Println("testBlockHeight", testBlockHeight)

	suite.CITS.Commit() // commit to passive trigger EVM Tx indexer

	balance := suite.CITS.QueryBalance(0, receiver.GetCosmosAddress().String())
	suite.Require().False(balance.IsZero(), "receiver must received some balance")

	blockResult, err := suite.CITS.QueryClients.TendermintRpcHttpClient.Block(context.Background(), ptrInt64(testBlockHeight))
	suite.Require().NoError(err)
	suite.Require().NotNil(blockResult)
	suite.Require().Equal(evmTxsCount+nonEvmTxsCount, len(blockResult.Block.Txs), "must be same as sent txs count for both EVM & non-EVM txs")

	gotCountByBlockNumber := suite.GetEthPublicAPI().GetBlockTransactionCountByNumber(rpctypes.BlockNumber(testBlockHeight))
	suite.Require().NotNil(gotCountByBlockNumber)
	gotCountByBlockHash := suite.GetEthPublicAPI().GetBlockTransactionCountByHash(common.BytesToHash(blockResult.Block.Hash()))
	suite.Require().NotNil(gotCountByBlockHash)

	if suite.Equal(uint(evmTxsCount), uint(*gotCountByBlockNumber), "must be same as sent EVM txs") {
		if suite.Equal(uint(evmTxsCount), uint(*gotCountByBlockHash), "must be same as sent EVM txs") {
			suite.Equal(gotCountByBlockNumber, gotCountByBlockHash, "result of eth_getBlockTransactionCountByNumber and eth_getBlockTransactionCountByHash must be same")
		}
	}
}
