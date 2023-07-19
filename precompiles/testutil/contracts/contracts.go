// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	evmosapp "github.com/evmos/evmos/v13/app"
	"github.com/evmos/evmos/v13/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v13/precompiles/testutil"
	evmosutil "github.com/evmos/evmos/v13/testutil"
	evmtypes "github.com/evmos/evmos/v13/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Call is a helper function to call any arbitrary smart contract.
func Call(ctx sdk.Context, app *evmosapp.Evmos, args CallArgs) (res abci.ResponseDeliverTx, ethRes *evmtypes.MsgEthereumTxResponse, err error) {
	var (
		nonce    uint64
		gasLimit = args.GasLimit
	)

	if args.PrivKey == nil {
		return abci.ResponseDeliverTx{}, nil, fmt.Errorf("private key is required; got: %v", args.PrivKey)
	}

	pk, ok := args.PrivKey.(*ethsecp256k1.PrivKey)
	if !ok {
		return abci.ResponseDeliverTx{}, nil, errors.New("error while casting type ethsecp256k1.PrivKey on provided private key")
	}

	key, err := pk.ToECDSA()
	if err != nil {
		return abci.ResponseDeliverTx{}, nil, fmt.Errorf("error while converting private key to ecdsa: %v", err)
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
		gasPrice = app.FeeMarketKeeper.GetBaseFee(ctx) // default gas price == block base fee
	} else {
		gasPrice = args.GasPrice
	}

	// Create MsgEthereumTx that calls the contract
	input, err := args.ContractABI.Pack(args.MethodName, args.Args...)
	if err != nil {
		return abci.ResponseDeliverTx{}, nil, fmt.Errorf("error while packing the input: %v", err)
	}

	msg := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		ChainID:   app.EvmKeeper.ChainID(),
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

	res, err = evmosutil.DeliverEthTxWithoutCheck(app, args.PrivKey, msg)
	if err != nil {
		return abci.ResponseDeliverTx{}, nil, fmt.Errorf("error during deliver tx: %s", err)
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
func CallContractAndCheckLogs(ctx sdk.Context, app *evmosapp.Evmos, cArgs CallArgs, logCheckArgs testutil.LogCheckArgs) (abci.ResponseDeliverTx, *evmtypes.MsgEthereumTxResponse, error) {
	res, ethRes, err := Call(ctx, app, cArgs)
	if err != nil {
		return abci.ResponseDeliverTx{}, nil, err
	}

	logCheckArgs.Res = res
	return res, ethRes, testutil.CheckLogs(logCheckArgs)
}
