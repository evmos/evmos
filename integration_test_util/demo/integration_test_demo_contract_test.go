package demo

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/integration_test_util"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"math/big"
)

//goland:noinspection SpellCheckingInspection

func (suite *DemoTestSuite) Test_Contract_DeployContracts() {
	suite.Run("ERC-20", func() {
		deployer := suite.CITS.WalletAccounts.Number(1)

		newContractAddress, _, resDeliver, err := suite.CITS.TxDeployErc20Contract(deployer, "coin", "token", 18)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(0), newContractAddress)
		suite.assertContractCode(newContractAddress)
	})

	suite.Run("1-create.sol", func() {
		deployer := suite.CITS.WalletAccounts.Number(2)

		newContractAddress, _, resDeliver, err := suite.CITS.TxDeploy1StorageContract(deployer)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(0), newContractAddress)
		suite.assertContractCode(newContractAddress)

		suite.Run("send a success tx", func() {
			data, err := integration_test_util.Contract1Storage.ABI.Pack("store", big.NewInt(7))
			suite.Require().NoError(err)
			_, resDeliver, err = suite.CITS.TxSendEvmTx(suite.Ctx(), deployer, &newContractAddress, nil, data)
			suite.Require().NoError(err)
			suite.Require().NotNil(resDeliver)
			suite.NotEmpty(resDeliver.CosmosTxHash)
			suite.NotEmpty(resDeliver.EthTxHash)
			suite.Empty(resDeliver.EvmError)
			suite.Commit()
		})

		suite.Run("send a failed tx", func() {
			data, err := integration_test_util.Contract1Storage.ABI.Pack("store", big.NewInt(3))
			suite.Require().NoError(err)
			_, resDeliver, err = suite.CITS.TxSendEvmTx(suite.Ctx(), deployer, &newContractAddress, nil, data)
			suite.Require().Error(err)
			suite.Require().NotNil(resDeliver)
			suite.NotEmpty(resDeliver.CosmosTxHash)
			suite.NotEmpty(resDeliver.EthTxHash)
			suite.NotEmpty(resDeliver.EvmError)
			suite.Commit()
		})
	})

	suite.Run("2-wevmos.sol", func() {
		deployer := suite.CITS.WalletAccounts.Number(3)

		newContractAddress, _, resDeliver, err := suite.CITS.TxDeploy2WEvmosContract(deployer, nil)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(0), newContractAddress)
		suite.assertContractCode(newContractAddress)
	})

	suite.Run("3-nft721.sol", func() {
		deployer := suite.CITS.WalletAccounts.Number(4)

		newContractAddress, _, resDeliver, err := suite.CITS.TxDeploy3Nft721Contract(deployer, nil)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(0), newContractAddress)
		suite.assertContractCode(newContractAddress)
	})

	suite.Run("4-nft1155.sol", func() {
		deployer := suite.CITS.WalletAccounts.Number(5)

		newContractAddress, _, resDeliver, err := suite.CITS.TxDeploy4Nft1155Contract(deployer, nil)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(0), newContractAddress)
		suite.assertContractCode(newContractAddress)
	})

	suite.Run("5-create.sol", func() {
		deployer := suite.CITS.WalletAccounts.Number(1)

		accDeployer := suite.CITS.ChainApp.AccountKeeper().GetAccount(suite.Ctx(), deployer.GetCosmosAddress())
		var nonce uint64 = accDeployer.GetSequence()

		fooContractAddress, _, resDeliver, err := suite.CITS.TxDeploy5CreateFooContract(deployer)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(nonce), fooContractAddress)
		suite.assertContractCode(fooContractAddress)

		nonce++

		barContractAddress, _, resDeliver, err := suite.CITS.TxDeploy5CreateBarContract(deployer)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(nonce), barContractAddress)
		suite.assertContractCode(barContractAddress)

		nonce++

		barInteractionContractAddress, _, resDeliver, err := suite.CITS.TxDeploy5CreateBarInteractionContract(deployer, barContractAddress)
		suite.Commit()
		suite.Require().NoError(err)
		suite.Require().NotNil(resDeliver)
		suite.NotEmpty(resDeliver.CosmosTxHash)
		suite.NotEmpty(resDeliver.EthTxHash)
		suite.Empty(resDeliver.EvmError)
		suite.Require().Equal(deployer.ComputeContractAddress(nonce), barInteractionContractAddress)
		suite.assertContractCode(barInteractionContractAddress)
	})
}

func (suite *DemoTestSuite) assertContractCode(contractAddress common.Address) {
	res, err := suite.CITS.QueryClients.EVM.Code(suite.Ctx(), &evmtypes.QueryCodeRequest{
		Address: contractAddress.String(),
	})
	if suite.NoError(err) {
		suite.Require().NotEmpty(res.Code)
		suite.Require().True(len(res.Code) >= 45)
	}
}
