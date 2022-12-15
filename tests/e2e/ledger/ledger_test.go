package ledger_test

// func (s *LedgerE2ESuite) TestAddLedgerKey() {
// 	mocks.RegisterClose(s.secp256k1)
// 	mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)
// 	clientCtx := s.network.Validators[0].ClientCtx

// 	cmd := keys.AddKeyCommand()
// 	clientCtx.OutputFormat = "text"
// 	out, err := testcli.ExecTestCLICmd(clientCtx, cmd, []string{"ledger_key", fmt.Sprintf("--%s", flags.FlagUseLedger)})
// 	s.Require().NoError(err)
// 	s.Require().NotEmpty(out.String(), "no output provided")
// 	s.T().Log(out.String())

// 	s.Require().NoError(s.network.WaitForNextBlock())

// 	//s.app.AccountKeeper.NewAccountWithAddress()
// }

// func (s *LedgerE2ESuite) TestSignMsg() {
// 	mocks.RegisterClose(s.secp256k1)
// 	mocks.RegisterGetAddressPubKeySECP256K1(s.secp256k1, s.accAddr, s.pubKey)
// 	mocks.RegisterSignSECP256K1(s.secp256k1)

// 	clientCtx := s.network.Validators[0].ClientCtx
// 	clientCtx.OutputFormat = "text"

// 	_, receiver, _, _ := s.CreateKeyPair()

// 	out, err := testcli.ExecTestCLICmd(
// 		clientCtx,
// 		bankcli.NewSendTxCmd(),
// 		[]string{
// 			"ledger_key",
// 			receiver.String(),
// 			sdk.NewCoin("aevmos", sdk.NewInt(100)).String(),
// 			s.FormatFlag(flags.FlagKeyringBackend),
// 			"test",
// 			s.FormatFlag(flags.FlagKeyringDir),
// 			"./build/node0/evmoscli/keyring-test",
// 		},
// 	)
// 	s.Require().NoError(err)

// 	s.Require().NotEmpty(out.String(), "no output provided")
// 	s.T().Log(out.String())

// }
