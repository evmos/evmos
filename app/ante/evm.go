// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmante "github.com/evmos/evmos/v18/app/ante/evm"
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
