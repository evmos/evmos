package erc20_test

import (
	"math/big"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	commonfactory "github.com/evmos/evmos/v15/testutil/integration/common/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Gomega
	. "github.com/onsi/gomega"
)

// setupSendAuthz is a helper function to set up a SendAuthorization for
// a given grantee and granter combination for a given amount.
//
// NOTE: A default expiration of 1 hour after the current block time is used.
func (s *PrecompileTestSuite) setupSendAuthz(
	grantee sdk.AccAddress, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	granter := sdk.AccAddress(granterPriv.PubKey().Address())
	expiration := s.network.GetContext().BlockHeader().Time.Add(time.Hour)
	sendAuthz := banktypes.NewSendAuthorization(
		amount,
		[]sdk.AccAddress{},
	)

	msgGrant, err := authz.NewMsgGrant(
		granter,
		grantee,
		sendAuthz,
		&expiration,
	)
	s.Require().NoError(err, "failed to create MsgGrant")

	// Create an authorization
	txArgs := commonfactory.CosmosTxArgs{Msgs: []sdk.Msg{msgGrant}}
	_, err = s.factory.ExecuteCosmosTx(granterPriv, txArgs)
	s.Require().NoError(err, "failed to execute MsgGrant")
}

// setupSendAuthzForContract is a helper function which executes an approval
// for the given contract data.
//
// If:
//   - the classic ERC20 contract is used, it calls the `approve` method on the contract.
//   - in other cases, it sends a `MsgGrant` to set up the authorization.
func (s *PrecompileTestSuite) setupSendAuthzForContract(
	callType int, contractData ContractData, grantee common.Address, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	Expect(amount).To(HaveLen(1), "expected only one coin")
	Expect(amount[0].Denom).To(Equal(s.tokenDenom),
		"this test utility only works with the token denom in the context of these integration tests",
	)

	switch callType {
	case directCall:
		s.setupSendAuthz(grantee.Bytes(), granterPriv, amount)
	case contractCall:
		s.setupSendAuthz(grantee.Bytes(), granterPriv, amount)
	case erc20Call:
		s.setupSendAuthzForERC20(callType, contractData, grantee, granterPriv, amount)
	case erc20V5Call:
		s.setupSendAuthzForERC20(callType, contractData, grantee, granterPriv, amount)
	case erc20V5CallerCall:
		s.setupSendAuthzForERC20(callType, contractData, grantee, granterPriv, amount)
	default:
		panic("unknown contract call type")
	}
}

// setupSendAuthzForERC20 is a helper function to set up a SendAuthorization for
// a given grantee and granter combination for a given amount.
func (s *PrecompileTestSuite) setupSendAuthzForERC20(
	callType int, contractData ContractData, grantee common.Address, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	txArgs, callArgs := s.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee, amount.AmountOf(s.tokenDenom).BigInt())

	// Check that an approval was made
	var abiEvents map[string]abi.Event
	switch callType {
	case erc20Call:
		abiEvents = contractData.erc20ABI.Events
	case erc20V5Call:
		abiEvents = contractData.erc20V5ABI.Events
	case erc20V5CallerCall:
		// NOTE: In order to set up an allowance from the granter to the grantee, the call needs
		// to be made to the actual ERC20 contract, not the ERC20Caller contract.
		abiEvents = contractData.erc20V5ABI.Events
		txArgs.To = &contractData.erc20V5Addr
		callArgs.ContractABI = contractData.erc20V5ABI
	default:
		panic("unknown contract call type")
	}

	approveCheck := testutil.LogCheckArgs{
		ABIEvents: abiEvents,
		ExpEvents: []string{auth.EventTypeApproval},
		ExpPass:   true,
	}

	_, _, err := s.factory.CallContractAndCheckLogs(granterPriv, txArgs, callArgs, approveCheck)
	Expect(err).ToNot(HaveOccurred(), "failed to execute approve")
}

// requireOut is a helper utility to reduce the amount of boilerplate code in the query tests.
//
// It requires the output bytes and error to match the expected values. Additionally, the method outputs
// are unpacked and the first value is compared to the expected value.
//
// NOTE: It's sufficient to only check the first value because all methods in the ERC20 precompile only
// return a single value.
func (s *PrecompileTestSuite) requireOut(
	bz []byte,
	err error,
	method abi.Method,
	expPass bool,
	errContains string,
	expValue interface{},
) {
	if expPass {
		s.Require().NoError(err, "expected no error")
		s.Require().NotEmpty(bz, "expected bytes not to be empty")

		// Unpack the name into a string
		out, err := method.Outputs.Unpack(bz)
		s.Require().NoError(err, "expected no error unpacking")

		// Check if expValue is a big.Int. Because of a difference in uninitialized/empty values for big.Ints,
		// this comparison is often not working as expected, so we convert to Int64 here and compare those values.
		bigExp, ok := expValue.(*big.Int)
		if ok {
			bigOut, ok := out[0].(*big.Int)
			s.Require().True(ok, "expected output to be a big.Int")
			s.Require().Equal(bigExp.Int64(), bigOut.Int64(), "expected different value")
		} else {
			s.Require().Equal(expValue, out[0], "expected different value")
		}
	} else {
		s.Require().Error(err, "expected error")
		s.Require().Contains(err.Error(), errContains, "expected different error")
	}
}

// requireSendAuthz is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount.
//
// NOTE: This helper expects only one authorization to exist.
func (s *PrecompileTestSuite) requireSendAuthz(grantee, granter sdk.AccAddress, amount sdk.Coins, allowList []string) {
	grants, err := s.grpcHandler.GetGrantsByGrantee(grantee.String())
	s.Require().NoError(err, "expected no error querying the grants")
	s.Require().Len(grants, 1, "expected one grant")
	s.Require().Equal(grantee.String(), grants[0].Grantee, "expected different grantee")
	s.Require().Equal(granter.String(), grants[0].Granter, "expected different granter")

	authzs, err := s.grpcHandler.GetAuthorizationsByGrantee(grantee.String())
	s.Require().NoError(err, "expected no error unpacking the authorization")
	s.Require().Len(authzs, 1, "expected one authorization")

	sendAuthz, ok := authzs[0].(*banktypes.SendAuthorization)
	s.Require().True(ok, "expected send authorization")

	s.Require().Equal(amount, sendAuthz.SpendLimit, "expected different spend limit amount")
	if len(allowList) == 0 {
		s.Require().Empty(sendAuthz.AllowList, "expected empty allow list")
	} else {
		s.Require().Equal(allowList, sendAuthz.AllowList, "expected different allow list")
	}
}

// setupERC20Precompile is a helper function to set up an instance of the ERC20 precompile for
// a given token denomination, set the token pair in the ERC20 keeper and adds the precompile
// to the available and active precompiles.
func (s *PrecompileTestSuite) setupERC20Precompile(denom string) *erc20.Precompile {
	tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), denom, erc20types.OWNER_MODULE)
	s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)

	precompile, err := erc20.NewPrecompile(
		tokenPair,
		s.network.App.BankKeeper,
		s.network.App.AuthzKeeper,
		s.network.App.TransferKeeper,
	)
	s.Require().NoError(err, "failed to create %q erc20 precompile", denom)

	err = s.network.App.EvmKeeper.AddEVMExtensions(s.network.GetContext(), precompile)
	s.Require().NoError(err, "failed to add %q erc20 precompile to EVM extensions", denom)

	return precompile
}

// callType constants to differentiate between direct calls and calls through a contract.
const (
	directCall = iota + 1
	contractCall
	erc20Call
	erc20V5Call
	erc20V5CallerCall
)

// getCallArgs is a helper function to return the correct call arguments for a given call type.
//
// In case of a direct call to the precompile, the precompile's ABI is used. Otherwise, the
// ERC20CallerContract's ABI is used and the given contract address.
func (s *PrecompileTestSuite) getTxAndCallArgs(
	callType int,
	contractData ContractData,
	methodName string,
	args ...interface{},
) (evmtypes.EvmTxArgs, factory.CallArgs) {
	txArgs := evmtypes.EvmTxArgs{}
	callArgs := factory.CallArgs{}

	switch callType {
	case directCall:
		txArgs.To = &contractData.precompileAddr
		callArgs.ContractABI = contractData.precompileABI
	case contractCall:
		txArgs.To = &contractData.contractAddr
		callArgs.ContractABI = contractData.contractABI
	case erc20Call:
		txArgs.To = &contractData.erc20Addr
		callArgs.ContractABI = contractData.erc20ABI
	case erc20V5Call:
		txArgs.To = &contractData.erc20V5Addr
		callArgs.ContractABI = contractData.erc20V5ABI
	case erc20V5CallerCall:
		txArgs.To = &contractData.erc20V5CallerAddr
		callArgs.ContractABI = contractData.erc20V5CallerABI
	default:
		panic("unknown contract call type")
	}

	callArgs.MethodName = methodName
	callArgs.Args = args

	return txArgs, callArgs
}

// ExpectedBalance is a helper struct to check the balances of accounts.
type ExpectedBalance struct {
	address  sdk.AccAddress
	expCoins sdk.Coins
}

// ExpectBalances is a helper function to check if the balances of the given accounts are as expected.
func (s *PrecompileTestSuite) ExpectBalances(expBalances []ExpectedBalance) {
	for _, expBalance := range expBalances {
		for _, expCoin := range expBalance.expCoins {
			coinBalance, err := s.grpcHandler.GetBalance(expBalance.address, expCoin.Denom)
			Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")
			Expect(coinBalance.Balance.Amount.Int64()).To(Equal(expCoin.Amount.Int64()), "expected different balance")
		}
	}
}

// ExpectBalancesForContract is a helper function to check expected balances for given accounts depending
// on the call type.
func (s *PrecompileTestSuite) ExpectBalancesForContract(callType int, contractData ContractData, expBalances []ExpectedBalance) {
	switch callType {
	case directCall:
		s.ExpectBalances(expBalances)
	case contractCall:
		s.ExpectBalances(expBalances)
	case erc20Call:
		s.ExpectBalancesForERC20(callType, contractData, expBalances)
	case erc20V5Call:
		s.ExpectBalancesForERC20(callType, contractData, expBalances)
	case erc20V5CallerCall:
		s.ExpectBalancesForERC20(callType, contractData, expBalances)
	default:
		panic("unknown contract call type")
	}
}

// ExpectBalancesForERC20 is a helper function to check expected balances for given accounts
// when using the ERC20 contract.
func (s *PrecompileTestSuite) ExpectBalancesForERC20(callType int, contractData ContractData, expBalances []ExpectedBalance) {
	for _, expBalance := range expBalances {
		for _, expCoin := range expBalance.expCoins {
			addr := common.BytesToAddress(expBalance.address.Bytes())

			txArgs, callArgs := s.getTxAndCallArgs(callType, contractData, "balanceOf", addr)

			passCheck := testutil.LogCheckArgs{ExpPass: true}

			_, ethRes, err := s.factory.CallContractAndCheckLogs(contractData.ownerPriv, txArgs, callArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")

			var balance *big.Int
			err = contractData.erc20ABI.UnpackIntoInterface(&balance, "balanceOf", ethRes.Ret)
			Expect(err).ToNot(HaveOccurred(), "expected no error unpacking balance")
			Expect(balance.Int64()).To(Equal(expCoin.Amount.Int64()), "expected different balance")
		}
	}
}

// expectSendAuthz is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount and optionally an access list.
//
// NOTE: This helper expects only one authorization to exist.
//
// NOTE 2: This mirrors the requireSendAuthz method but adapted to Ginkgo.
func (s *PrecompileTestSuite) expectSendAuthz(grantee, granter sdk.AccAddress, expAmount sdk.Coins) {
	authzs, err := s.grpcHandler.GetAuthorizations(grantee.String(), granter.String())
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking the authorization")
	Expect(authzs).To(HaveLen(1), "expected one authorization")

	sendAuthz, ok := authzs[0].(*banktypes.SendAuthorization)
	Expect(ok).To(BeTrue(), "expected send authorization")

	Expect(sendAuthz.SpendLimit).To(Equal(expAmount), "expected different spend limit amount")
}

// expectSendAuthzForERC20 is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount.
func (s *PrecompileTestSuite) expectSendAuthzForERC20(callType int, contractData ContractData, grantee, granter common.Address, expAmount sdk.Coins) {
	txArgs, callArgs := s.getTxAndCallArgs(callType, contractData, auth.AllowanceMethod, granter, grantee)

	passCheck := testutil.LogCheckArgs{ExpPass: true}

	_, ethRes, err := s.factory.CallContractAndCheckLogs(contractData.ownerPriv, txArgs, callArgs, passCheck)
	Expect(err).ToNot(HaveOccurred(), "expected no error getting allowance")

	var allowance *big.Int
	err = contractData.erc20ABI.UnpackIntoInterface(&allowance, "allowance", ethRes.Ret)
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking allowance")
	Expect(allowance.Int64()).To(Equal(expAmount.AmountOf(s.tokenDenom).Int64()), "expected different allowance")
}

// ExpectSendAuthzForContract is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount and optionally an access list.
//
// NOTE: This helper expects only one authorization to exist.
func (s *PrecompileTestSuite) ExpectSendAuthzForContract(
	callType int, contractData ContractData, grantee, granter common.Address, expAmount sdk.Coins,
) {
	switch callType {
	case directCall:
		s.expectSendAuthz(grantee.Bytes(), granter.Bytes(), expAmount)
	case contractCall:
		s.expectSendAuthz(grantee.Bytes(), granter.Bytes(), expAmount)
	case erc20Call:
		s.expectSendAuthzForERC20(callType, contractData, grantee, granter, expAmount)
	case erc20V5Call:
		s.expectSendAuthzForERC20(callType, contractData, grantee, granter, expAmount)
	case erc20V5CallerCall:
		s.expectSendAuthzForERC20(callType, contractData, grantee, granter, expAmount)
	default:
		panic("unknown contract call type")
	}
}

// expectNoSendAuthz is a helper function to check that no SendAuthorization
// exists for a given grantee and granter combination.
func (s *PrecompileTestSuite) expectNoSendAuthz(grantee, granter sdk.AccAddress) {
	authzs, err := s.grpcHandler.GetAuthorizations(grantee.String(), granter.String())
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking the authorizations")
	Expect(authzs).To(HaveLen(0), "expected no authorizations")
}

// expectNoSendAuthzForERC20 is a helper function to check that no SendAuthorization
// exists for a given grantee and granter combination.
func (s *PrecompileTestSuite) expectNoSendAuthzForERC20(callType int, contractData ContractData, grantee, granter common.Address) {
	s.expectSendAuthzForERC20(callType, contractData, grantee, granter, sdk.Coins{})
}

// ExpectNoSendAuthzForContract is a helper function to check that no SendAuthorization
// exists for a given grantee and granter combination.
func (s *PrecompileTestSuite) ExpectNoSendAuthzForContract(
	callType int, contractData ContractData, grantee, granter common.Address,
) {
	switch callType {
	case directCall:
		s.expectNoSendAuthz(grantee.Bytes(), granter.Bytes())
	case contractCall:
		s.expectNoSendAuthz(grantee.Bytes(), granter.Bytes())
	case erc20Call:
		s.expectNoSendAuthzForERC20(callType, contractData, grantee, granter)
	case erc20V5Call:
		s.expectNoSendAuthzForERC20(callType, contractData, grantee, granter)
	case erc20V5CallerCall:
		s.expectNoSendAuthzForERC20(callType, contractData, grantee, granter)
	default:
		panic("unknown contract call type")
	}
}

// ContractData is a helper struct to hold the addresses and ABIs for the
// different contract instances that are subject to testing here.
type ContractData struct {
	ownerPriv cryptotypes.PrivKey

	erc20Addr         common.Address
	erc20ABI          abi.ABI
	erc20V5Addr       common.Address
	erc20V5ABI        abi.ABI
	erc20V5CallerAddr common.Address
	erc20V5CallerABI  abi.ABI
	contractAddr      common.Address
	contractABI       abi.ABI
	precompileAddr    common.Address
	precompileABI     abi.ABI
}

// fundWithTokens is a helper function for the scope of the ERC20 integration tests.
// Depending on the passed call type, it funds the given address with tokens either
// using the Bank module or by minting straight on the ERC20 contract.
func (s *PrecompileTestSuite) fundWithTokens(
	callType int,
	contractData ContractData,
	receiver common.Address,
	fundCoins sdk.Coins,
) {
	Expect(fundCoins).To(HaveLen(1), "expected only one coin")
	Expect(fundCoins[0].Denom).To(Equal(s.tokenDenom),
		"this helper function only supports funding with the token denom in the context of these integration tests",
	)

	var err error

	switch callType {
	case directCall:
		err = s.network.FundAccount(receiver.Bytes(), fundCoins)
	case contractCall:
		err = s.network.FundAccount(receiver.Bytes(), fundCoins)
	case erc20Call:
		err = s.MintERC20(callType, contractData, receiver, fundCoins.AmountOf(s.tokenDenom).BigInt())
	case erc20V5Call:
		err = s.MintERC20(callType, contractData, receiver, fundCoins.AmountOf(s.tokenDenom).BigInt())
	case erc20V5CallerCall:
		err = s.MintERC20(callType, contractData, receiver, fundCoins.AmountOf(s.tokenDenom).BigInt())
	default:
		panic("unknown contract call type")
	}

	Expect(err).ToNot(HaveOccurred(), "failed to fund account")
}

// MintERC20 is a helper function to mint tokens on the ERC20 contract.
//
// NOTE: we are checking that there was a Transfer event emitted (which happens on minting).
func (s *PrecompileTestSuite) MintERC20(callType int, contractData ContractData, receiver common.Address, amount *big.Int) error {
	txArgs, callArgs := s.getTxAndCallArgs(callType, contractData, "mint", receiver, amount)

	var abiEvents map[string]abi.Event
	switch callType {
	case erc20Call:
		abiEvents = contractData.erc20ABI.Events
	case erc20V5Call:
		abiEvents = contractData.erc20V5ABI.Events
	case erc20V5CallerCall:
		// NOTE: When using the ERC20 caller contract, we must still mint from the actual ERC20 v5 contract.
		abiEvents = contractData.erc20V5ABI.Events
		txArgs.To = &contractData.erc20V5Addr
		callArgs.ContractABI = contractData.erc20V5ABI
	default:
		panic("unknown contract call type")
	}

	mintCheck := testutil.LogCheckArgs{
		ABIEvents: abiEvents,
		ExpEvents: []string{"Transfer"}, // NOTE: this event occurs when calling "mint" on ERC20s
		ExpPass:   true,
	}

	_, _, err := s.factory.CallContractAndCheckLogs(contractData.ownerPriv, txArgs, callArgs, mintCheck)

	return err
}
