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
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdktestutilcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/evmos/v10/app"
	"github.com/evmos/evmos/v10/tests/integration/ledger/mocks"
	"github.com/evmos/evmos/v10/testutil"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
)

var (
	signOkMock = func(_ []uint32, msg []byte) ([]byte, error) {
		b, err := s.privKey.Sign(msg)
		return b, err
	}

	signErrMock = func(_ []uint32, msg []byte) ([]byte, error) {
		return nil, mocks.ErrMockedSigning
	}
)

/*
1. Connect a Ledger device to your laptop using USB (Bluetooth is not supported)
2. Start a local Evmos node using ./init.sh
3. Unlock the Ledger and open the Ethereum Ledger app
4. Add the Ledger as an Evmos key using evmosd keys add myledger --ledger; it should display the Ledger device's default Ethereum address (copy this value)
5. Send funds to your Ledger account using evmosd tx bank send mykey [your Ledger address] 100000000000000000aevmos --fees 200aevmos
6. Check the balance of your Ledger account using evmosd query bank balances [your Ledger address]
7. Send funds from your Ledger account using evmosd tx bank send myledger evmos1e4etd2u9c2huyjacswsfukugztxvd9du52y49t 1000aevmos --fees 200aevmos
8. Check the balances of your Ledger account and the destination account using evmosd query bank balances [your Ledger address] and evmosd query bank balances evmos1e4etd2u9c2huyjacswsfukugztxvd9du52y49t
*/

var _ = Describe("ledger cli and keyring functionality", func() {
	var (
		receiverAccAddr sdk.AccAddress
		receiverEthAddr common.Address
		encCfg          params.EncodingConfig
		kr              keyring.Keyring
		mockedIn        sdktestutil.BufferReader
		clientCtx       client.Context
		ctx             context.Context
		cmd             *cobra.Command
		krHome          string
		txProto         []byte
		keyRecord       *keyring.Record
	)

	ledgerKey := "ledger_key"

	s.SetupTest()
	s.SetupEvmosApp()

	fmt.Println(receiverEthAddr, receiverAccAddr, txProto)

	Describe("Perform key addition", func() {
		BeforeEach(func() {
			// account key
			krHome = s.T().TempDir()
			encCfg = encoding.MakeConfig(app.ModuleBasics)
		})
		Context("add ledger key with different algorythms", func() {
			BeforeEach(func() {

				cmd = keys.AddKeyCommand()
				cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

				mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

				kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

				mocks.RegisterClose(s.ledger)
				mocks.RegisterGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)
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
			cmd = keys.AddKeyCommand()
			cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

			mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)
			mocks.RegisterGetAddressPubKeySECP256K1(s.ledger, s.accAddr, s.pubKey)

			kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			cmd.SetArgs([]string{ledgerKey, s.FormatFlag(flags.FlagUseLedger), s.FormatFlag(flags.FlagKeyAlgorithm), "eth_secp256k1"})
			s.Require().NoError(cmd.ExecuteContext(ctx))

			keyRecord, err = kr.Key(ledgerKey)
			s.Require().NoError(err, "can't find ledger key")
		})
		Context("tx bank send", func() {
			Context("keyring execution scope", func() {
				BeforeEach(func() {

					s.ledger = mocks.NewSECP256K1(s.T())

					mocks.RegisterClose(s.ledger)
					mocks.RegisterGetPublicKeySECP256K1(s.ledger, s.pubKey)

				})
				It("should return provided to sign message", func() {
					mocks.RegisterSignSECP256K1(s.ledger, signOkMock, nil)

					ledgerAddr, err := keyRecord.GetAddress()
					s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

					msg := []byte("test message")

					signed, _, err := kr.SignByAddress(ledgerAddr, msg)
					s.Require().NoError(err, "failed to sign messsage")
					_ = signed

					valid := s.pubKey.VerifySignature(msg, signed)
					s.Require().True(valid, "invalid sigrature returned")

				})

				It("should raise error from ledger sign function to the top", func() {
					mocks.RegisterSignSECP256K1Error(s.ledger)

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
					mocks.RegisterClose(s.ledger)
					mocks.RegisterGetPublicKeySECP256K1(s.ledger, s.pubKey)

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
					receiverAccAddr, err = sdk.AccAddressFromBech32(sdk.MustBech32ifyAddressBytes("evmos", sk.PubKey().Bytes()))

					cmd = bankcli.NewSendTxCmd()
					mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

					kr, clientCtx, ctx = s.NewKeyringAndCtxs(krHome, mockedIn, encCfg)
					s.Require().NoError(err)

				})
				It("should execute bank tx", func() {
					mocks.RegisterSignSECP256K1(s.ledger, signOkMock, nil)

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

					s.Require().NoError(err)
				})
				It("should execute bank tx", func() {
					mocks.RegisterSignSECP256K1(s.ledger, signErrMock, mocks.ErrMockedSigning)

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

					s.Require().Error(err)
					s.Require().Equal(mocks.ErrMockedSigning.Error(), err.Error())
				})
			})

		})

	})

})
