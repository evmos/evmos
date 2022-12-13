package ledger_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v10/tests/integration/ledger/mocks"
	"github.com/evmos/evmos/v10/testutil"
	testcli "github.com/evmos/evmos/v10/testutil/cli"

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

	s.SetupEvmosApp()
	s.SetupLedger()
	//s.RegisterMocks(txProto)
	s.SetupNetwork()

	fmt.Println(receiverEthAddr, txProto)

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

		})

		It("should add ledger key to keys list", func() {
			mocks.RegisterClose(s.secp256k1)
			mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)
			clientCtx := s.network.Validators[0].ClientCtx

			cmd := keys.AddKeyCommand()
			clientCtx.OutputFormat = "text"
			out, err := testcli.ExecTestCLICmd(clientCtx, cmd, []string{"ledger_key", fmt.Sprintf("--%s", flags.FlagUseLedger)})
			s.Require().NoError(err)
			s.Require().NotEmpty(out.String(), "no output provided")
			s.T().Log(out.String())

			s.Require().NoError(s.network.WaitForNextBlock())

			//s.app.AccountKeeper.NewAccountWithAddress()
		})
		It("should sign valid tx with verifiable signature", func() {
			mocks.RegisterClose(s.secp256k1)
			mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)
			mocks.RegisterSignSECP256K1(s.secp256k1)

			clientCtx := s.network.Validators[0].ClientCtx
			clientCtx.OutputFormat = "text"

			out, err := testcli.ExecTestCLICmd(
				clientCtx,
				bankcli.NewSendTxCmd(),
				[]string{
					"ledger_key",
					receiverAccAddr.String(),
					sdk.NewCoin("aevmos", sdk.NewInt(100)).String(),
					s.FormatFlag(flags.FlagKeyringBackend),
					"test",
					s.FormatFlag(flags.FlagKeyringDir),
					"./build/node0/evmoscli/keyring-test",
				},
			)
			s.Require().NoError(err)

			s.Require().NotEmpty(out.String(), "no output provided")
			s.T().Log(out.String())
		})

	})
	s.TearDownSuite()

})
