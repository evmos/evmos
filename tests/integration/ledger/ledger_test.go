package main_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v10/app"
	evmosd "github.com/evmos/evmos/v10/cmd/evmosd"
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

var _ = Describe("Ledger", func() {
	var (
		receiverAccAddr sdk.AccAddress
		receiverEthAddr common.Address
		txProto         []byte
	)
	fmt.Println(receiverEthAddr)
	BeforeEach(
		func() {
			s.SetupEvmosApp()
			s.SetupLedger()
		},
	)

	Describe("Test evmosd ledger cli commands", func() {
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
			txProto = s.getMockTxProtobuf(receiverAccAddr, 1000)
			// s.RegisterMocks(txProto)
		})

		It("should add ledger key to keys list", func() {
			rootCmd, _ := evmosd.NewRootCmd()
			rootCmd.SetArgs([]string{
				"keys",
				"add",
				"ledger",
				fmt.Sprintf("--%s", flags.FlagUseLedger),
			})

			err := svrcmd.Execute(rootCmd, "EVMOSD", app.DefaultNodeHome)
			s.Require().NoError(err)

			ctx := s.validator.
				testcli.ExecTestCLICmd()

		})
		It("should sign valid tx with verifiable signature", func() {

			sign, err := s.SECP256K1.SignSECP256K1(ethaccounts.DefaultBaseDerivationPath, txProto)
			s.Require().NoError(err, "can't sign tx")

			valid := crypto.VerifySignature(crypto.FromECDSAPub(s.pubKey), crypto.Keccak256Hash(txProto).Bytes(), sign)
			s.Require().True(valid, "invalid signrature")
		})
	})

})
