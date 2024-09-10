// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package tx

import (
	"context"
	"encoding/json"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/server/config"
	"github.com/evmos/evmos/v19/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// PrepareEthTx creates an ethereum tx and signs it with the provided messages and private key.
// It returns the signed transaction and an error
func PrepareEthTx(
	txCfg client.TxConfig,
	appEvmos *app.Evmos,
	priv cryptotypes.PrivKey,
	msgs ...sdk.Msg,
) (authsigning.Tx, error) {
	txBuilder := txCfg.NewTxBuilder()

	signer := ethtypes.LatestSignerForChainID(appEvmos.EvmKeeper.ChainID())
	txFee := sdk.Coins{}
	txGasLimit := uint64(0)

	// Sign messages and compute gas/fees.
	for _, m := range msgs {
		msg, ok := m.(*evmtypes.MsgEthereumTx)
		if !ok {
			return nil, errorsmod.Wrapf(errorsmod.Error{}, "cannot mix Ethereum and Cosmos messages in one Tx")
		}

		if priv != nil {
			err := msg.Sign(signer, NewSigner(priv))
			if err != nil {
				return nil, err
			}
		}

		msg.From = ""

		txGasLimit += msg.GetGas()
		txFee = txFee.Add(sdk.Coin{Denom: utils.BaseDenom, Amount: sdkmath.NewIntFromBigInt(msg.GetFee())})
	}

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// Set the extension
	var option *codectypes.Any
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errorsmod.Wrapf(errorsmod.Error{}, "could not set extensions for Ethereum tx")
	}

	builder.SetExtensionOptions(option)

	txBuilder.SetGasLimit(txGasLimit)
	txBuilder.SetFeeAmount(txFee)

	return txBuilder.GetTx(), nil
}

// CreateEthTx is a helper function to create and sign an Ethereum transaction.
//
// If the given private key is not nil, it will be used to sign the transaction.
//
// It offers the ability to increment the nonce by a given amount in case one wants to set up
// multiple transactions that are supposed to be executed one after another.
// Should this not be the case, just pass in zero.
func CreateEthTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	privKey cryptotypes.PrivKey,
	from sdk.AccAddress,
	dest sdk.AccAddress,
	amount *big.Int,
	nonceIncrement int,
) (*evmtypes.MsgEthereumTx, error) {
	toAddr := common.BytesToAddress(dest.Bytes())
	fromAddr := common.BytesToAddress(from.Bytes())
	chainID := appEvmos.EvmKeeper.ChainID()

	// When we send multiple Ethereum Tx's in one Cosmos Tx, we need to increment the nonce for each one.
	nonce := appEvmos.EvmKeeper.GetNonce(ctx, fromAddr) + uint64(nonceIncrement) //nolint:gosec // G115
	evmTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &toAddr,
		Amount:    amount,
		GasLimit:  100000,
		GasFeeCap: appEvmos.FeeMarketKeeper.GetBaseFee(ctx),
		GasTipCap: big.NewInt(1),
		Accesses:  &ethtypes.AccessList{},
	}
	msgEthereumTx := evmtypes.NewTx(evmTxParams)
	msgEthereumTx.From = fromAddr.String()

	// If we are creating multiple eth Tx's with different senders, we need to sign here rather than later.
	if privKey != nil {
		signer := ethtypes.LatestSignerForChainID(appEvmos.EvmKeeper.ChainID())
		err := msgEthereumTx.Sign(signer, NewSigner(privKey))
		if err != nil {
			return nil, err
		}
	}

	return msgEthereumTx, nil
}

// GasLimit estimates the gas limit for the provided parameters. To achieve
// this, need to provide the corresponding QueryClient to call the
// `eth_estimateGas` rpc method. If not provided, returns a default value
func GasLimit(ctx context.Context, from common.Address, data evmtypes.HexString, queryClientEvm evmtypes.QueryClient) (uint64, error) {
	// default gas limit (used if no queryClientEvm is provided)
	gas := uint64(100000000000)

	if queryClientEvm != nil {
		args, err := json.Marshal(&evmtypes.TransactionArgs{
			From: &from,
			Data: (*hexutil.Bytes)(&data),
		})
		if err != nil {
			return gas, err
		}

		res, err := queryClientEvm.EstimateGas(ctx, &evmtypes.EthCallRequest{
			Args:   args,
			GasCap: config.DefaultGasCap,
		})
		if err != nil {
			return gas, err
		}
		gas = res.Gas
	}
	return gas, nil
}
