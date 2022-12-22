package ledger_test

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/evmos/v10/app"
	"github.com/evmos/evmos/v10/tests/integration/ledger/mocks"
	"github.com/evmos/evmos/v10/testutil"

	"github.com/spf13/cobra"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdktestutilcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	. "github.com/onsi/ginkgo/v2"
)

var (
	signOkMock = func(_ []uint32, msg []byte) ([]byte, error) {
		return s.privKey.Sign(msg)
	}

	signErrMock = func(_ []uint32, msg []byte) ([]byte, error) {
		return nil, mocks.ErrMockedSigning
	}
)

var _ = Describe("Ledger CLI and keyring functionality: ", func() {
	var (
		receiverAccAddr sdk.AccAddress
		encCfg          params.EncodingConfig
		kr              keyring.Keyring
		mockedIn        sdktestutil.BufferReader
		clientCtx       client.Context
		ctx             context.Context
		cmd             *cobra.Command
		krHome          string
		keyRecord       *keyring.Record
	)

	ledgerKey := "ledger_key"

	s.SetupTest()
	s.SetupEvmosApp()

	Describe("Perform key addition", func() {
		BeforeEach(func() {
			krHome = s.T().TempDir()
			encCfg = encoding.MakeConfig(app.ModuleBasics)
		})
		Context("add ledger key with different algorythms", func() {
			BeforeEach(func() {

				cmd = keys.AddKeyCommand()
				cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

				mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

				kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

				mocks.MClose(s.ledger)
				mocks.MGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)
			})
			It("should add the ledger key with eth_secp256k1", func() {
				out, err := sdktestutilcli.ExecTestCLICmd(clientCtx, cmd, []string{
					ledgerKey,
					fmt.Sprintf("--%s", flags.FlagUseLedger),
					s.FormatFlag(flags.FlagKeyAlgorithm),
					"eth_secp256k1",
				})

				s.Require().NoError(err)
				s.Require().Contains(out.String(), "name: ledger_key")

				_, err = kr.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")
			})
			It("should return error on ledger key addition with secp256k1", func() {

				_, err := sdktestutilcli.ExecTestCLICmd(clientCtx, cmd, []string{
					ledgerKey,
					fmt.Sprintf("--%s", flags.FlagUseLedger),
					s.FormatFlag(flags.FlagKeyAlgorithm),
					"secp256k1",
				})

				s.Require().Error(err, "false positive, error expected")
				s.Require().Contains(err.Error(), "")
			})
		})
	})
	Describe("Perform transaction signing", func() {
		BeforeEach(func() {
			krHome = s.T().TempDir()
			encCfg = encoding.MakeConfig(app.ModuleBasics)

			var err error
			// create add key command
			cmd = keys.AddKeyCommand()
			cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

			mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)
			mocks.MGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)

			kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			cmd.SetArgs([]string{ledgerKey, s.FormatFlag(flags.FlagUseLedger), s.FormatFlag(flags.FlagKeyAlgorithm), "eth_secp256k1"})
			// add ledger key for following tests
			s.Require().NoError(cmd.ExecuteContext(ctx))
			keyRecord, err = kr.Key(ledgerKey)
			s.Require().NoError(err, "can't find ledger key")
		})
		Context("tx bank send", func() {
			Context("keyring execution scope", func() {
				BeforeEach(func() {

					s.ledger = mocks.NewSECP256K1(s.T())

					mocks.MClose(s.ledger)
					mocks.MGetPublicKeySECP256K1(s.ledger, s.pubKey)

				})
				It("should return provided to sign message", func() {
					mocks.MSignSECP256K1(s.ledger, signOkMock, nil)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

					msg := []byte("test message")

					signed, _, err := kr.SignByAddress(ledgerAddr, msg)
					s.Require().NoError(err, "failed to sign messsage")

					valid := s.pubKey.VerifySignature(msg, signed)
					s.Require().True(valid, "invalid signature returned")
				})
				It("should raise error from ledger sign function to the top", func() {
					mocks.MSignSECP256K1(s.ledger, signErrMock, mocks.ErrMockedSigning)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

					msg := []byte("test message")

					_, _, err = kr.SignByAddress(ledgerAddr, msg)

					s.Require().Error(err, "false positive result, error expected")

					s.Require().Equal(mocks.ErrMockedSigning.Error(), err.Error(), "original and returned errors are not equal")
				})
			})
			Context("CLI execution scope", func() {
				BeforeEach(func() {
					s.ledger = mocks.NewSECP256K1(s.T())

					err := testutil.FundAccount(
						s.ctx,
						s.app.BankKeeper,
						s.accAddr,
						sdk.NewCoins(
							sdk.NewCoin("aevmos", sdk.NewInt(100000000000000)),
						),
					)
					s.Require().NoError(err)

					sk, err := ethsecp256k1.GenerateKey()
					s.Require().NoError(err)
					receiverAccAddr = sdk.AccAddress(tests.GenerateAddress().Bytes())

					cmd = bankcli.NewSendTxCmd()
					mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

					kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

					// register mocked funcs
					mocks.MClose(s.ledger)
					mocks.MGetPublicKeySECP256K1(s.ledger, s.pubKey)
					mocks.MEnsureExist(s.accRetriever, nil)
					mocks.MGetAccountNumberSequence(s.accRetriever, 0, 0, nil)
				})
				It("should execute bank tx", func() {
					mocks.MSignSECP256K1(s.ledger, signOkMock, nil)

					cmd.SetContext(ctx)
					cmd.SetArgs([]string{
						ledgerKey,
						receiverAccAddr.String(),
						sdk.NewCoin("aevmos", sdk.NewInt(1000)).String(),
						s.FormatFlag(flags.FlagUseLedger),
						s.FormatFlag(flags.FlagSkipConfirmation),
					})
					out := bytes.NewBufferString("")
					cmd.SetOutput(out)

					err := cmd.Execute()

					s.Require().NoError(err, "can't execute cli tx command")
				})
				It("should execute bank tx", func() {
					mocks.MSignSECP256K1(s.ledger, signErrMock, mocks.ErrMockedSigning)

					cmd.SetContext(ctx)
					cmd.SetArgs([]string{
						ledgerKey,
						receiverAccAddr.String(),
						sdk.NewCoin("aevmos", sdk.NewInt(1000)).String(),
						s.FormatFlag(flags.FlagUseLedger),
						s.FormatFlag(flags.FlagSkipConfirmation),
					})
					out := bytes.NewBufferString("")
					cmd.SetOutput(out)

					err := cmd.Execute()

					s.Require().Error(err, "false positive, error expected")
					s.Require().Equal(mocks.ErrMockedSigning.Error(), err.Error())
				})
			})
		})
	})
})
