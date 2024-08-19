package keeper_test

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/evmos/evmos/v19/contracts"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	evm "github.com/evmos/evmos/v19/x/evm/types"
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
