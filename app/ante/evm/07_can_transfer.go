package evm

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// CanTransferDecorator checks if the sender is allowed to transfer funds according to the EVM block
// context rules.
type CanTransferDecorator struct {
	evmKeeper EVMKeeper
}

// NewCanTransferDecorator creates a new CanTransferDecorator instance.
func NewCanTransferDecorator(evmKeeper EVMKeeper) CanTransferDecorator {
	return CanTransferDecorator{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle creates an EVM from the message and calls the BlockContext CanTransfer function to
// see if the address can execute the transaction.
func (ctd CanTransferDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	params := ctd.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(ctd.evmKeeper.ChainID())
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))
	baseFee := ctd.evmKeeper.GetBaseFee(ctx, ethCfg)
	isLondon := evmtypes.IsLondon(ethCfg, ctx.BlockHeight())

	if isLondon && baseFee == nil {
		return ctx, errorsmod.Wrap(
			evmtypes.ErrInvalidBaseFee,
			"base fee is supported but evm block context value is nil",
		)
	}

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		coreMsg, err := msgEthTx.AsMessage(signer, baseFee)
		if err != nil {
			return ctx, errorsmod.Wrapf(
				err,
				"failed to create an ethereum core.Message from signer %T", signer,
			)
		}

		if isLondon && coreMsg.GasFeeCap().Cmp(baseFee) < 0 {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrInsufficientFee,
				"max fee per gas less than block base fee (%s < %s)",
				coreMsg.GasFeeCap(), baseFee,
			)
		}

		// NOTE: pass in an empty coinbase address and nil tracer as we don't need them for the check below
		cfg := &statedb.EVMConfig{
			ChainConfig: ethCfg,
			Params:      params,
			CoinBase:    common.Address{},
			BaseFee:     baseFee,
		}

		stateDB := statedb.New(ctx, ctd.evmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash().Bytes())))
		evm := ctd.evmKeeper.NewEVM(ctx, coreMsg, cfg, evmtypes.NewNoOpTracer(), stateDB)

		// check that caller has enough balance to cover asset transfer for **topmost** call
		// NOTE: here the gas consumed is from the context with the infinite gas meter
		if coreMsg.Value().Sign() > 0 && !evm.Context.CanTransfer(stateDB, coreMsg.From(), coreMsg.Value()) {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrInsufficientFunds,
				"failed to transfer %s from address %s using the EVM block context transfer function",
				coreMsg.Value(),
				coreMsg.From(),
			)
		}
	}

	return next(ctx, tx, simulate)
}
