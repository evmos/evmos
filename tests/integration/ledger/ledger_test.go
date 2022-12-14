package ledger_test

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/crypto/hd"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/evmos/v10/app"
	"github.com/evmos/evmos/v10/tests/integration/ledger/mocks"
	"github.com/evmos/evmos/v10/testutil"

	. "github.com/onsi/ginkgo/v2"
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

var _ = Describe("Sign transaction with ledger key", func() {
	var (
		receiverAccAddr sdk.AccAddress
		receiverEthAddr common.Address
		encCfg          params.EncodingConfig
		kb              keyring.Keyring
		mockedIn        sdktestutil.BufferReader
		clientCtx       client.Context
		cmd             *cobra.Command
		kbHome          string
		txProto         []byte
	)
	ledgerKey := "ledger_key"

	s.SetupTest()

	fmt.Println(receiverEthAddr, receiverAccAddr, txProto)

	Describe("Perform key addition", func() {
		BeforeEach(func() {
			// account key
			priv, err := ethsecp256k1.GenerateKey()
			s.Require().NoError(err, "can't generate private key")

			receiverEthAddr = common.BytesToAddress(priv.PubKey().Address().Bytes())
			receiverAccAddr = sdk.AccAddress(s.ethAddr.Bytes())
			testutil.FundAccount(
				s.ctx,
				s.app.BankKeeper,
				s.accAddr,
				sdk.NewCoins(
					sdk.NewCoin("aevmos", sdk.NewInt(100000)),
				),
			)
			kbHome = s.T().TempDir()
			encCfg = encoding.MakeConfig(app.ModuleBasics)
		})
		Context("with eth_secp256k1 and secp256k1 keyring algorythms", func() {
			BeforeEach(func() {
				var err error
				cmd = keys.AddKeyCommand()
				cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

				mockedIn = sdktestutil.ApplyMockIODiscardOutErr(cmd)

				kb, err = keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockedIn, encCfg.Codec, s.MockKeyringOption(), hd.EthSecp256k1Option())
				s.Require().NoError(err)

				clientCtx = client.Context{}.
					WithKeyringDir(kbHome).
					WithKeyring(kb).
					WithCodec(encCfg.Codec)
			})
			It("should add the ledger key with default algo", func() {
				mocks.RegisterClose(s.secp256k1)
				mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)

				ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)
				b := bytes.NewBufferString("")
				cmd.SetOut(b)

				cmd.SetArgs([]string{ledgerKey, s.FormatFlag(flags.FlagUseLedger)})
				s.Require().NoError(cmd.ExecuteContext(ctx))

				_, err := kb.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")

				out, err := io.ReadAll(b)
				s.Require().NoError(err)
				s.Require().Contains(string(out), "name: ledger_key")
			})
			It("should add the ledger key with secp256k1", func() {
				mocks.RegisterClose(s.secp256k1)
				mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)

				ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)
				b := bytes.NewBufferString("")
				cmd.SetOut(b)

				cmd.SetArgs([]string{ledgerKey, fmt.Sprintf("--%s", flags.FlagUseLedger), s.FormatFlag(flags.FlagKeyAlgorithm), "secp256k1"})
				s.Require().NoError(cmd.ExecuteContext(ctx))

				_, err := kb.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")

				out, err := io.ReadAll(b)
				s.Require().NoError(err)
				s.Require().Contains(string(out), "name: ledger_key")
			})
			It("should add the ledger key with eth_secp256k1", func() {
				mocks.RegisterClose(s.secp256k1)
				mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)

				ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)
				b := bytes.NewBufferString("")
				cmd.SetOut(b)

				cmd.SetArgs([]string{ledgerKey, fmt.Sprintf("--%s", flags.FlagUseLedger), s.FormatFlag(flags.FlagKeyAlgorithm), "eth_secp256k1"})
				s.Require().NoError(cmd.ExecuteContext(ctx))

				_, err := kb.Key(ledgerKey)
				s.Require().NoError(err, "can't find ledger key")

				out, err := io.ReadAll(b)
				s.Require().NoError(err)
				s.Require().Contains(string(out), "name: ledger_key")
			})
		})
	})

	Describe("Perform transaction signing", func() {
		// BeforeEach(func() {
		// 	// account key
		// 	priv, err := ethsecp256k1.GenerateKey()
		// 	s.Require().NoError(err, "can't generate private key")

		// 	receiverEthAddr = common.BytesToAddress(priv.PubKey().Address().Bytes())
		// 	receiverAccAddr = sdk.AccAddress(s.ethAddr.Bytes())
		// 	testutil.FundAccount(
		// 		s.ctx,
		// 		s.app.BankKeeper,
		// 		s.accAddr,
		// 		sdk.NewCoins(
		// 			sdk.NewCoin("aevmos", sdk.NewInt(100000)),
		// 		),
		// 	)
		// 	kbHome = s.T().TempDir()
		// 	encCfg = encoding.MakeConfig(app.ModuleBasics)

		// })

		// It("should call signing function on ledger side", func() {
		// 	ledgerAddr, err := rec.GetAddress()
		// 	s.Require().NoError(err, "can't retirieve ledger addr from a keyring")

		// 	msg := []byte("hello")

		// 	signed, pubKey, err := kb.SignByAddress(ledgerAddr, msg)
		// 	s.Require().Equal(mocks.ErrMockedSigning, err)
		// 	fmt.Println(string(signed), pubKey)

		// 	s.Require().Equal(string(msg), string(signed))
		// })
	})
})
