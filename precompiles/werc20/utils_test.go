package werc20_test

// // callType constants to differentiate between direct calls and calls through a contract.
// const (
// 	directCall = iota + 1
// 	contractCall
// 	erc20Call
// )

// TODO: remove all of this
//
// // ContractData is a helper struct to hold the addresses and ABIs for the
// // different contract instances that are subject to testing here.
// type ContractData struct {
// 	ownerPriv cryptotypes.PrivKey

// 	erc20Addr      common.Address
// 	erc20ABI       abi.ABI
// 	contractAddr   common.Address
// 	contractABI    abi.ABI
// 	precompileAddr common.Address
// 	precompileABI  abi.ABI
// }

// // ExpectedBalance is a helper struct to check the balances of accounts.
// type ExpectedBalance struct {
// 	address  sdk.AccAddress
// 	expCoins sdk.Coins
// }

// // ExpectBalances is a helper function to check if the balances of the given accounts are as expected.
// func (s *WERC20IntegrationTestSuite) ExpectBalances(expBalances []ExpectedBalance) {
// 	for _, expBalance := range expBalances {
// 		for _, expCoin := range expBalance.expCoins {
// 			coinBalance, err := s.grpcHandler.GetBalance(expBalance.address, expCoin.Denom)
// 			Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")
// 			Expect(coinBalance.Balance.Amount.Int64()).To(Equal(expCoin.Amount.Int64()), "expected different balance")
// 		}
// 	}
// }

// // checkBalances is a helper function to check the balances of the sender and receiver.
// func (s *WERC20IntegrationTestSuite) checkBalances(failCheck testutil.LogCheckArgs, sender keyring.Key, contractData ContractData) {
// 	balanceCheck := failCheck.WithExpPass(true)
// 	txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractData, erc20.BalanceOfMethod, sender.Addr)

// 	_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
// 	Expect(err).ToNot(HaveOccurred(), "failed to execute balanceOf")

// 	// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
// 	balanceAfter, err := s.grpcHandler.GetBalance(sender.AccAddr, s.network.GetDenom())
// 	Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")

// 	var erc20Balance *big.Int
// 	err = s.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
// 	Expect(err).ToNot(HaveOccurred(), "failed to unpack result")
// 	Expect(erc20Balance).To(Equal(balanceAfter.Balance.Amount.BigInt()), "expected different balance")
// }

// // setupSendAuthz is a helper function to set up a SendAuthorization for
// // a given grantee and granter combination for a given amount.
// //
// // NOTE: A default expiration of 1 hour after the current block time is used.
// func (s *WERC20IntegrationTestSuite) setupSendAuthz(
// 	grantee sdk.AccAddress, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
// ) {
// 	granter := sdk.AccAddress(granterPriv.PubKey().Address())
// 	expiration := s.network.GetContext().BlockHeader().Time.Add(time.Hour)
// 	sendAuthz := banktypes.NewSendAuthorization(
// 		amount,
// 		[]sdk.AccAddress{},
// 	)

// 	msgGrant, err := authz.NewMsgGrant(
// 		granter,
// 		grantee,
// 		sendAuthz,
// 		&expiration,
// 	)
// 	s.Require().NoError(err, "failed to create MsgGrant")

// 	// Create an authorization
// 	txArgs := commonfactory.CosmosTxArgs{Msgs: []sdk.Msg{msgGrant}}
// 	_, err = s.factory.ExecuteCosmosTx(granterPriv, txArgs)
// 	s.Require().NoError(err, "failed to execute MsgGrant")
// }

// // setupSendAuthzForContract is a helper function which executes an approval
// // for the given contract data.
// //
// // If:
// //   - the classic ERC20 contract is used, it calls the `approve` method on the contract.
// //   - in other cases, it sends a `MsgGrant` to set up the authorization.
// //
// // TODO: Should we add more cases for WERC20Caller
// func (s *WERC20IntegrationTestSuite) setupSendAuthzForContract(
// 	callType int, grantee common.Address, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
// ) {
// 	Expect(amount).To(HaveLen(1), "expected only one coin")
// 	Expect(amount[0].Denom).To(Equal(s.network.GetDenom()),
// 		"this test utility only works with the token denom in the context of these integration tests",
// 	)

// 	switch callType {
// 	case directCall:
// 		s.setupSendAuthz(grantee.Bytes(), granterPriv, amount)
// 	default:
// 		panic("unknown contract call type")
// 	}
// }
