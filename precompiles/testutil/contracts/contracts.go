// Copyright Tharsis Labs Ltd.(Eidon-chain)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/Eidon-AI/eidon-chain/blob/main/LICENSE)

package contracts

import (
	"errors"
	"fmt"
	"math/big"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	eidon-chainapp "github.com/Eidon-AI/eidon-chain/v20/app"
	"github.com/Eidon-AI/eidon-chain/v20/crypto/ethsecp256k1"
	"github.com/Eidon-AI/eidon-chain/v20/precompiles/testutil"
	eidon-chainutil "github.com/Eidon-AI/eidon-chain/v20/testutil"
	evmtypes "github.com/Eidon-AI/eidon-chain/v20/x/evm/types"
)

// Call is a helper function to call any arbitrary smart contract.
func Call(ctx sdk.Context, app *eidon-chainapp.Eidon-chain, args CallArgs) (res abci.ExecTxResult, ethRes *evmtypes.MsgEthereumTxResponse, err error) {
	var (
		nonce    uint64
		gasLimit = args.GasLimit
	)

	if args.PrivKey == nil {
		return abci.ExecTxResult{}, nil, fmt.Errorf("private key is required; got: %v", args.PrivKey)
	}

	pk, ok := args.PrivKey.(*ethsecp256k1.PrivKey)
	if !ok {
		return abci.ExecTxResult{}, nil, errors.New("error while casting type ethsecp256k1.PrivKey on provided private key")
	}

	key, err := pk.ToECDSA()
	if err != nil {
		return abci.ExecTxResult{}, nil, fmt.Errorf("error while converting private key to ecdsa: %v", err)
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)

	if args.Nonce == nil {
		nonce = app.EvmKeeper.GetNonce(ctx, addr)
	} else {
		nonce = args.Nonce.Uint64()
	}

	// if gas limit not provided
	// use default
	if args.GasLimit == 0 {
		gasLimit = 1000000
	}

	// if gas price not provided
	var gasPrice *big.Int
	if args.GasPrice == nil {
		baseFeeRes, err := app.EvmKeeper.BaseFee(ctx, &evmtypes.QueryBaseFeeRequest{})
		if err != nil {
			return abci.ExecTxResult{}, nil, err
		}
		gasPrice = baseFeeRes.BaseFee.BigInt() // default gas price == block base fee
	} else {
		gasPrice = args.GasPrice
	}
	// Create MsgEthereumTx that calls the contract
	input, err := args.ContractABI.Pack(args.MethodName, args.Args...)
	if err != nil {
		return abci.ExecTxResult{}, nil, fmt.Errorf("error while packing the input: %v", err)
	}

	// Create MsgEthereumTx that calls the contract
	msg := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		ChainID:   evmtypes.GetEthChainConfig().ChainID,
		Nonce:     nonce,
		To:        &args.ContractAddr,
		Amount:    args.Amount,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		GasFeeCap: args.GasFeeCap,
		GasTipCap: args.GasTipCap,
		Input:     input,
		Accesses:  args.AccessList,
	})
	msg.From = addr.Hex()

	res, err = eidon-chainutil.DeliverEthTx(app, args.PrivKey, msg)
	if err != nil {
		return res, nil, fmt.Errorf("error during deliver tx: %s", err)
	}
	if !res.IsOK() {
		return res, nil, fmt.Errorf("error during deliver tx: %v", res.Log)
	}

	ethRes, err = evmtypes.DecodeTxResponse(res.Data)
	if err != nil {
		return res, nil, fmt.Errorf("error while decoding tx response: %v", err)
	}

	return res, ethRes, nil
}

// CallContractAndCheckLogs is a helper function to call any arbitrary smart contract and check that the logs
// contain the expected events.
func CallContractAndCheckLogs(ctx sdk.Context, app *eidon-chainapp.Eidon-chain, cArgs CallArgs, logCheckArgs testutil.LogCheckArgs) (abci.ExecTxResult, *evmtypes.MsgEthereumTxResponse, error) {
	res, ethRes, err := Call(ctx, app, cArgs)
	if err != nil {
		return res, nil, err
	}

	logCheckArgs.Res = res
	return res, ethRes, testutil.CheckLogs(logCheckArgs)
}
