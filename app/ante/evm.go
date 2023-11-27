// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	cosmosante "github.com/evmos/evmos/v15/app/ante/cosmos"
	evmante "github.com/evmos/evmos/v15/app/ante/evm"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

func newMonoEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		evmante.NewMonoDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.DistributionKeeper,
			options.StakingKeeper,
			options.MaxTxGasWanted,
		),
	)
}

// newEVMAnteHandler creates the default ante handler for Ethereum transactions
func newEVMAnteHandler(options HandlerOptions) sdk.AnteHandler { //nolint: unused
	return sdk.ChainAnteDecorators(
		// outermost AnteDecorator. SetUpContext must be called first
		evmante.NewEthSetUpContextDecorator(options.EvmKeeper),
		// Check eth effective gas price against the node's minimal-gas-prices config
		evmante.NewEthMempoolFeeDecorator(options.EvmKeeper),
		// Check eth effective gas price against the global MinGasPrice
		evmante.NewEthMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		evmante.NewEthValidateBasicDecorator(options.EvmKeeper),
		evmante.NewEthSigVerificationDecorator(options.EvmKeeper),
		evmante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.EvmKeeper),
		evmante.NewCanTransferDecorator(options.EvmKeeper),
		evmante.NewEthVestingTransactionDecorator(options.AccountKeeper, options.BankKeeper, options.EvmKeeper),
		evmante.NewEthGasConsumeDecorator(options.BankKeeper, options.DistributionKeeper, options.EvmKeeper, options.StakingKeeper, options.MaxTxGasWanted),
		evmante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
		// emit eth tx hash and index at the very last ante handler.
		evmante.NewEthEmitEventDecorator(options.EvmKeeper),
	)
}

// newCosmosAnteHandlerEip712 creates the ante handler for transactions signed with EIP712
func newLegacyCosmosAnteHandlerEip712(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		cosmosante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		cosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		cosmosante.NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		cosmosante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.DistributionKeeper, options.FeegrantKeeper, options.StakingKeeper, options.TxFeeChecker),
		cosmosante.NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper, options.BankKeeper, options.Cdc),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		//nolint: staticcheck
		cosmosante.NewLegacyEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}
