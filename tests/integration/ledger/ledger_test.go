package ledger_test

import (
	"bytes"
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/crypto/hd"
	"github.com/evmos/evmos/v18/encoding"
	"github.com/evmos/evmos/v18/tests/integration/ledger/mocks"

	"github.com/spf13/cobra"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdktestutilcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdktestutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
)

var (
	signOkMock = func(_ []uint32, msg []byte) ([]byte, error) {
		return s.privKey.Sign(msg)
	}

	signErrMock = func([]uint32, []byte) ([]byte, error) {
		return nil, mocks.ErrMockedSigning
	}
)

var _ = Describe("Ledger CLI and keyring functionality: ", func() {
	var (
		encCfg    sdktestutilmod.TestEncodingConfig
		kr        keyring.Keyring
		mockedIn  sdktestutil.BufferReader
		clientCtx client.Context
		ctx       context.Context
		cmd       *cobra.Command
		krHome    string
		keyRecord *keyring.Record
	)

	ledgerKey := "ledger_key"

	s.SetupTest()
	s.SetupEvmosApp()

	Describe("Adding a key from ledger using the CLI", func() {
		BeforeEach(func() {
			krHome = s.T().TempDir()
			encCfg = encoding.MakeConfig(app.ModuleBasics)

			cmd = s.evmosAddKeyCmd()

			mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

			kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

			mocks.MClose(s.ledger)
			mocks.MGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)
		})
		Context("with default algo", func() {
			It("should use eth_secp256k1 by default and pass", func() {
				out, err := sdktestutilcli.ExecTestCLICmd(clientCtx, cmd, []string{
					ledgerKey,
					s.FormatFlag(flags.FlagUseLedger),
				})

				s.Require().NoError(err)
				s.Require().Contains(out.String(), "name: ledger_key")

				_, err = kr.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")
			})
		})
		Context("with eth_secp256k1 algo", func() {
			It("should add the ledger key ", func() {
				out, err := sdktestutilcli.ExecTestCLICmd(clientCtx, cmd, []string{
					ledgerKey,
					s.FormatFlag(flags.FlagUseLedger),
					s.FormatFlag(flags.FlagKeyType),
					string(hd.EthSecp256k1Type),
				})

				s.Require().NoError(err)
				s.Require().Contains(out.String(), "name: ledger_key")

				_, err = kr.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")
			})
		})
	})
	Describe("Singing a transactions", func() {
		BeforeEach(func() {
			krHome = s.T().TempDir()
			encCfg = encoding.MakeConfig(app.ModuleBasics)

			var err error

			// create add key command
			cmd = s.evmosAddKeyCmd()

			mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)
			mocks.MGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)

			kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			cmd.SetArgs([]string{
				ledgerKey,
				s.FormatFlag(flags.FlagUseLedger),
				s.FormatFlag(flags.FlagKeyType),
				"eth_secp256k1",
			})
			// add ledger key for following tests
			s.Require().NoError(cmd.ExecuteContext(ctx))
			keyRecord, err = kr.Key(ledgerKey)
			s.Require().NoError(err, "can't find ledger key")
		})
		Context("perform bank send", func() {
			Context("with keyring functions calling", func() {
				BeforeEach(func() {
					s.ledger = mocks.NewSECP256K1(s.T())

					mocks.MClose(s.ledger)
					mocks.MGetPublicKeySECP256K1(s.ledger, s.pubKey)
				})
				It("should return valid signature", func() {
					mocks.MSignSECP256K1(s.ledger, signOkMock, nil)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

					msg := []byte("test message")

					signed, _, err := kr.SignByAddress(ledgerAddr, msg, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
					s.Require().NoError(err, "failed to sign messsage")

					valid := s.pubKey.VerifySignature(msg, signed)
					s.Require().True(valid, "invalid signature returned")
				})
				It("should raise error from ledger sign function to the top", func() {
					mocks.MSignSECP256K1(s.ledger, signErrMock, mocks.ErrMockedSigning)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

					msg := []byte("test message")

					_, _, err = kr.SignByAddress(ledgerAddr, msg, signingtypes.SignMode_SIGN_MODE_TEXTUAL)

					s.Require().Error(err, "false positive result, error expected")

					s.Require().Equal(mocks.ErrMockedSigning.Error(), err.Error(), "original and returned errors are not equal")
				})
			})
		})
	})
})
