// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// GetERC20Balance is a helper method to return the balance of the given ERC-20 contract for the given address.
func GetERC20Balance(txFactory testfactory.TxFactory, priv cryptotypes.PrivKey, erc20Addr common.Address) (*big.Int, error) {
	addr := common.BytesToAddress(priv.PubKey().Address().Bytes())

	return GetERC20BalanceForAddr(txFactory, priv, addr, erc20Addr)
}

// GetERC20BalanceForAddr is a helper method to return the balance of the given ERC-20 contract for the given address.
//
// NOTE: Under the hood this sends an actual EVM transaction instead of just querying the JSON-RPC.
// TODO: Use query instead of transaction in future.
func GetERC20BalanceForAddr(txFactory testfactory.TxFactory, priv cryptotypes.PrivKey, addr, erc20Addr common.Address) (*big.Int, error) {
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	txArgs := evmtypes.EvmTxArgs{
		To: &erc20Addr,
	}

	callArgs := testfactory.CallArgs{
		ContractABI: erc20ABI,
		MethodName:  "balanceOf",
		Args:        []interface{}{addr},
	}

	// TODO: should rather be done with EthCall
	res, err := txFactory.ExecuteContractCall(priv, txArgs, callArgs)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to execute contract call")
	}

	ethRes, err := evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to decode tx response")
	}
	if len(ethRes.Ret) == 0 {
		return nil, fmt.Errorf("got empty return value from contract call")
	}

	balanceI, err := erc20ABI.Unpack("balanceOf", ethRes.Ret)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to unpack balance")
	}

	balance, ok := balanceI[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert balance to big.Int; got %T", balanceI[0])
	}

	return balance, nil
}
