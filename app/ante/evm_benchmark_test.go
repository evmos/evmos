// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ante_test

import (
	"fmt"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v20/app/ante"
	ethante "github.com/evmos/evmos/v20/app/ante/evm"
	cmmnfactory "github.com/evmos/evmos/v20/testutil/integration/common/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

type benchmarkSuite struct {
	network     *network.UnitTestNetwork
	grpcHandler grpc.Handler
	txFactory   factory.TxFactory
	keyring     testkeyring.Keyring
}

// Setup
var table = []struct {
	name     string
	txType   string
	simulate bool
}{
	{
		"evm_transfer_sim",
		"evm_transfer",
		true,
	},
	{
		"evm_transfer",
		"evm_transfer",
		false,
	},
	{
		"bank_msg_send_sim",
		"bank_msg_send",
		true,
	},
	{
		"bank_msg_send",
		"bank_msg_send",
		false,
	},
}

func BenchmarkAnteHandler(b *testing.B) {
	for _, v := range table {
		// Reset chain on every tx type to have a clean state
		// and a fair benchmark
		b.StopTimer()
		keyring := testkeyring.New(2)
		unitNetwork := network.NewUnitTestNetwork(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
		txFactory := factory.New(unitNetwork, grpcHandler)
		suite := benchmarkSuite{
			network:     unitNetwork,
			grpcHandler: grpcHandler,
			txFactory:   txFactory,
			keyring:     keyring,
		}

		handlerOptions := suite.generateHandlerOptions()
		ante := ante.NewAnteHandler(handlerOptions)
		b.StartTimer()

		b.Run(fmt.Sprintf("tx_type_%v", v.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Stop timer while building the tx setup
				b.StopTimer()
				// Start with a clean block
				if err := unitNetwork.NextBlock(); err != nil {
					fmt.Println(err)
					break
				}
				ctx := unitNetwork.GetContext()

				// Generate fresh tx type
				tx, err := suite.generateTxType(v.txType)
				if err != nil {
					fmt.Println(err)
					break
				}
				b.StartTimer()

				// Run benchmark
				_, err = ante(ctx, tx, v.simulate)
				if err != nil {
					fmt.Println(err)
					break
				}
			}
		})
	}
}

func (s *benchmarkSuite) generateTxType(txType string) (sdktypes.Tx, error) {
	switch txType {
	case "evm_transfer":
		senderPriv := s.keyring.GetPrivKey(0)
		receiver := s.keyring.GetKey(1)
		txArgs := evmtypes.EvmTxArgs{
			To:     &receiver.Addr,
			Amount: big.NewInt(1000),
		}
		return s.txFactory.GenerateSignedEthTx(senderPriv, txArgs)
	case "bank_msg_send":
		sender := s.keyring.GetKey(1)
		receiver := s.keyring.GetAccAddr(0)
		bankmsg := banktypes.NewMsgSend(
			sender.AccAddr,
			receiver,
			sdktypes.NewCoins(
				sdktypes.NewCoin(
					s.network.GetDenom(),
					math.NewInt(1000),
				),
			),
		)
		txArgs := cmmnfactory.CosmosTxArgs{Msgs: []sdktypes.Msg{bankmsg}}
		return s.txFactory.BuildCosmosTx(sender.Priv, txArgs)
	default:
		return nil, fmt.Errorf("invalid tx type")
	}
}

func (s *benchmarkSuite) generateHandlerOptions() ante.HandlerOptions {
	encCfg := s.network.GetEncodingConfig()
	return ante.HandlerOptions{
		Cdc:                    s.network.App.AppCodec(),
		AccountKeeper:          s.network.App.AccountKeeper,
		BankKeeper:             s.network.App.BankKeeper,
		ExtensionOptionChecker: evmostypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              s.network.App.EvmKeeper,
		StakingKeeper:          s.network.App.StakingKeeper,
		FeegrantKeeper:         s.network.App.FeeGrantKeeper,
		DistributionKeeper:     s.network.App.DistrKeeper,
		IBCKeeper:              s.network.App.IBCKeeper,
		FeeMarketKeeper:        s.network.App.FeeMarketKeeper,
		SignModeHandler:        encCfg.TxConfig.SignModeHandler(),
		SigGasConsumer:         ante.SigVerificationGasConsumer,
		MaxTxGasWanted:         1_000_000_000,
		TxFeeChecker:           ethante.NewDynamicFeeChecker(s.network.App.EvmKeeper),
	}
}
