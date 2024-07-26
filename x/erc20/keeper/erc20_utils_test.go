package keeper_test

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/evmos/evmos/v18/contracts"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	evm "github.com/evmos/evmos/v18/x/evm/types"
)

func (suite *KeeperTestSuite) MintERC20Token(contractAddr, to common.Address, amount *big.Int) (abcitypes.ExecTxResult, error) {
	res, err := suite.factory.ExecuteContractCall(
		suite.keyring.GetPrivKey(0),
		evm.EvmTxArgs{
			To: &contractAddr,
		},
		factory.CallArgs{
			ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
			MethodName:  "mint",
			Args:        []interface{}{to, amount},
		},
	)

	if err != nil {
		return res, err
	}

	return res, suite.network.NextBlock()
}

// func (suite *KeeperTestSuite) TransferERC20TokenToModule(contractAddr, from common.Address, amount *big.Int) *evm.MsgEthereumTx {
// 	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", types.ModuleAddress, amount)
// 	suite.Require().NoError(err)
// 	return suite.sendTx(contractAddr, from, transferData)
// }

// func (suite *KeeperTestSuite) GrantERC20Token(contractAddr, from, to common.Address, roleString string) *evm.MsgEthereumTx {
// 	// 0xCc508cD0818C85b8b8a1aB4cEEef8d981c8956A6 MINTER_ROLE
// 	role := crypto.Keccak256([]byte(roleString))
// 	// needs to be an array not a slice
// 	var v [32]byte
// 	copy(v[:], role)

// 	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("grantRole", v, to)
// 	suite.Require().NoError(err)
// 	return suite.sendTx(contractAddr, from, transferData)
// }

func (suite *KeeperTestSuite) BalanceOf(contract, account common.Address) (interface{}, error) {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.factory.ExecuteContractCall(
		suite.keyring.GetPrivKey(0),
		evm.EvmTxArgs{
			To: &contract,
		},
		factory.CallArgs{
			ContractABI: erc20,
			MethodName:  "balanceOf",
			Args:        []interface{}{account},
		},
	)

	if err != nil {
		return nil, err
	}

	ethRes, err := evm.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, err
	}

	unpacked, err := erc20.Unpack("balanceOf", ethRes.Ret)
	if err != nil {
		return nil, err
	}
	if len(unpacked) == 0 {
		return nil, errors.New("nothing unpacked from response")
	}

	return unpacked[0], suite.network.NextBlock()
}

// func (suite *KeeperTestSuite) NameOf(contract common.Address) string {
// 	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

// 	res, err := suite.app.EvmKeeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, false, "name")
// 	suite.Require().NoError(err)
// 	suite.Require().NotNil(res)

// 	unpacked, err := erc20.Unpack("name", res.Ret)
// 	suite.Require().NoError(err)
// 	suite.Require().NotEmpty(unpacked)

// 	return fmt.Sprintf("%v", unpacked[0])
// }

// func (suite *KeeperTestSuite) TransferERC20Token(contractAddr, from, to common.Address, amount *big.Int) *evm.MsgEthereumTx {
// 	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", to, amount)
// 	suite.Require().NoError(err)
// 	return suite.sendTx(contractAddr, from, transferData)
// }
