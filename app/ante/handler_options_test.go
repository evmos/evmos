package ante_test

import (
	ethante "github.com/evmos/ethermint/app/ante"
	"github.com/evmos/ethermint/encoding"
	ethermint "github.com/evmos/ethermint/types"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/app/ante"
)

func (suite *AnteTestSuite) TestValidateHandlerOptions() {
	cases := []struct {
		name    string
		options ante.HandlerOptions
		expPass bool
	}{
		{
			"fail - empty options",
			ante.HandlerOptions{},
			false,
		},
		{
			"fail - empty account keeper",
			ante.HandlerOptions{
				Cdc:           suite.app.AppCodec(),
				AccountKeeper: nil,
			},
			false,
		},
		{
			"fail - empty bank keeper",
			ante.HandlerOptions{
				Cdc:           suite.app.AppCodec(),
				AccountKeeper: suite.app.AccountKeeper,
				BankKeeper:    nil,
			},
			false,
		},
		{
			"fail - empty IBC keeper",
			ante.HandlerOptions{
				Cdc:           suite.app.AppCodec(),
				AccountKeeper: suite.app.AccountKeeper,
				BankKeeper:    suite.app.BankKeeper,
				IBCKeeper:     nil,
			},
			false,
		},
		{
			"fail - empty staking keeper",
			ante.HandlerOptions{
				Cdc:           suite.app.AppCodec(),
				AccountKeeper: suite.app.AccountKeeper,
				BankKeeper:    suite.app.BankKeeper,
				IBCKeeper:     suite.app.IBCKeeper,
				StakingKeeper: nil,
			},
			false,
		},
		{
			"fail - empty fee market keeper",
			ante.HandlerOptions{
				Cdc:             suite.app.AppCodec(),
				AccountKeeper:   suite.app.AccountKeeper,
				BankKeeper:      suite.app.BankKeeper,
				IBCKeeper:       suite.app.IBCKeeper,
				StakingKeeper:   suite.app.StakingKeeper,
				FeeMarketKeeper: nil,
			},
			false,
		},
		{
			"fail - empty EVM keeper",
			ante.HandlerOptions{
				Cdc:             suite.app.AppCodec(),
				AccountKeeper:   suite.app.AccountKeeper,
				BankKeeper:      suite.app.BankKeeper,
				IBCKeeper:       suite.app.IBCKeeper,
				StakingKeeper:   suite.app.StakingKeeper,
				FeeMarketKeeper: suite.app.FeeMarketKeeper,
				EvmKeeper:       nil,
			},
			false,
		},
		{
			"fail - empty signature gas consumer",
			ante.HandlerOptions{
				Cdc:             suite.app.AppCodec(),
				AccountKeeper:   suite.app.AccountKeeper,
				BankKeeper:      suite.app.BankKeeper,
				IBCKeeper:       suite.app.IBCKeeper,
				StakingKeeper:   suite.app.StakingKeeper,
				FeeMarketKeeper: suite.app.FeeMarketKeeper,
				EvmKeeper:       suite.app.EvmKeeper,
				SigGasConsumer:  nil,
			},
			false,
		},
		{
			"fail - empty signature mode handler",
			ante.HandlerOptions{
				Cdc:             suite.app.AppCodec(),
				AccountKeeper:   suite.app.AccountKeeper,
				BankKeeper:      suite.app.BankKeeper,
				IBCKeeper:       suite.app.IBCKeeper,
				StakingKeeper:   suite.app.StakingKeeper,
				FeeMarketKeeper: suite.app.FeeMarketKeeper,
				EvmKeeper:       suite.app.EvmKeeper,
				SigGasConsumer:  app.SigVerificationGasConsumer,
				SignModeHandler: nil,
			},
			false,
		},
		{
			"success - default app options",
			ante.HandlerOptions{
				Cdc:                    suite.app.AppCodec(),
				AccountKeeper:          suite.app.AccountKeeper,
				BankKeeper:             suite.app.BankKeeper,
				ExtensionOptionChecker: ethermint.HasDynamicFeeExtensionOption,
				EvmKeeper:              suite.app.EvmKeeper,
				StakingKeeper:          suite.app.StakingKeeper,
				FeegrantKeeper:         suite.app.FeeGrantKeeper,
				IBCKeeper:              suite.app.IBCKeeper,
				FeeMarketKeeper:        suite.app.FeeMarketKeeper,
				SignModeHandler:        encoding.MakeConfig(app.ModuleBasics).TxConfig.SignModeHandler(),
				SigGasConsumer:         app.SigVerificationGasConsumer,
				MaxTxGasWanted:         40000000,
				TxFeeChecker:           ethante.NewDynamicFeeChecker(suite.app.EvmKeeper),
			},
			true,
		},
	}

	for _, tc := range cases {
		err := tc.options.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
