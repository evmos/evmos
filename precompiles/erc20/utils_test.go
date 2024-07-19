package erc20_test

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	auth "github.com/evmos/evmos/v19/precompiles/authorization"
	"github.com/evmos/evmos/v19/precompiles/erc20"
	"github.com/evmos/evmos/v19/precompiles/testutil"
	commonfactory "github.com/evmos/evmos/v19/testutil/integration/common/factory"
	commonnetwork "github.com/evmos/evmos/v19/testutil/integration/common/network"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	network "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"

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
	err := setupSendAuthz(
		s.network,
		s.factory,
		grantee,
		granterPriv,
		amount,
	)
	s.Require().NoError(err, "failed to set up send authorization")
}

func (is *IntegrationTestSuite) setupSendAuthz(
	grantee sdk.AccAddress, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	err := setupSendAuthz(
		is.network,
		is.factory,
		grantee,
		granterPriv,
		amount,
	)
	Expect(err).ToNot(HaveOccurred(), "failed to set up send authorization")
}

func setupSendAuthz(
	network commonnetwork.Network,
	factory commonfactory.TxFactory,
	grantee sdk.AccAddress,
	granterPriv cryptotypes.PrivKey,
	amount sdk.Coins,
) error {
	granter := sdk.AccAddress(granterPriv.PubKey().Address())
	expiration := network.GetContext().BlockHeader().Time.Add(time.Hour)
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
	if err != nil {
		return errorsmod.Wrap(err, "failed to create MsgGrant")
	}

	// Create an authorization
	txArgs := commonfactory.CosmosTxArgs{Msgs: []sdk.Msg{msgGrant}}
	_, err = factory.ExecuteCosmosTx(granterPriv, txArgs)
	if err != nil {
		return errorsmod.Wrap(err, "failed to execute MsgGrant")
	}

	return nil
}

// setupSendAuthzForContract is a helper function which executes an approval
// for the given contract data.
//
// If:
//   - the classic ERC20 contract is used, it calls the `approve` method on the contract.
//   - in other cases, it sends a `MsgGrant` to set up the authorization.
func (is *IntegrationTestSuite) setupSendAuthzForContract(
	callType CallType, contractData ContractsData, grantee common.Address, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	Expect(amount).To(HaveLen(1), "expected only one coin")
	Expect(amount[0].Denom).To(Equal(is.tokenDenom),
		"this test utility only works with the token denom in the context of these integration tests",
	)

	switch {
	case slices.Contains(nativeCallTypes, callType):
		is.setupSendAuthz(grantee.Bytes(), granterPriv, amount)
	case slices.Contains(erc20CallTypes, callType):
		is.setupSendAuthzForERC20(callType, contractData, grantee, granterPriv, amount)
	default:
		panic("unknown contract call type")
	}
}

// setupSendAuthzForERC20 is a helper function to set up a SendAuthorization for
// a given grantee and granter combination for a given amount.
func (is *IntegrationTestSuite) setupSendAuthzForERC20(
	callType CallType, contractData ContractsData, grantee common.Address, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	if callType == erc20V5CallerCall {
		// NOTE: When using the ERC20 caller contract, we must still approve from the actual ERC20 v5 contract.
		callType = erc20V5Call
	}

	abiEvents := contractData.GetContractData(callType).ABI.Events

	txArgs, callArgs := is.getTxAndCallArgs(callType, contractData, auth.ApproveMethod, grantee, amount.AmountOf(is.tokenDenom).BigInt())

	approveCheck := testutil.LogCheckArgs{
		ABIEvents: abiEvents,
		ExpEvents: []string{auth.EventTypeApproval},
		ExpPass:   true,
	}

	_, _, err := is.factory.CallContractAndCheckLogs(granterPriv, txArgs, callArgs, approveCheck)
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

	precompile, err := setupERC20PrecompileForTokenPair(*s.network, tokenPair)
	s.Require().NoError(err, "failed to set up %q erc20 precompile", tokenPair.Denom)

	return precompile
}

// setupERC20Precompile is a helper function to set up an instance of the ERC20 precompile for
// a given token denomination, set the token pair in the ERC20 keeper and adds the precompile
// to the available and active precompiles.
//
// TODO: refactor
func (is *IntegrationTestSuite) setupERC20Precompile(denom string) *erc20.Precompile {
	tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), denom, erc20types.OWNER_MODULE)
	is.network.App.Erc20Keeper.SetToken(is.network.GetContext(), tokenPair)

	precompile, err := setupERC20PrecompileForTokenPair(*is.network, tokenPair)
	Expect(err).ToNot(HaveOccurred(), "failed to set up %q erc20 precompile", tokenPair.Denom)

	return precompile
}

// setupERC20PrecompileForTokenPair is a helper function to set up an instance of the ERC20 precompile for
// a given token pair and adds the precompile to the available and active precompiles.
func setupERC20PrecompileForTokenPair(
	unitNetwork network.UnitTestNetwork, tokenPair erc20types.TokenPair,
) (*erc20.Precompile, error) {
	precompile, err := erc20.NewPrecompile(
		tokenPair,
		unitNetwork.App.BankKeeper,
		unitNetwork.App.AuthzKeeper,
		unitNetwork.App.TransferKeeper,
	)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to create %q erc20 precompile", tokenPair.Denom)
	}

	err = unitNetwork.App.Erc20Keeper.EnableDynamicPrecompiles(
		unitNetwork.GetContext(),
		precompile.Address(),
	)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to add %q erc20 precompile to EVM extensions", tokenPair.Denom)
	}

	return precompile, nil
}

// CallType indicates which type of contract call is made during the integration tests.
type CallType int

// callType constants to differentiate between direct calls and calls through a contract.
const (
	directCall CallType = iota + 1
	contractCall
	erc20Call
	erc20CallerCall
	erc20V5Call
	erc20V5CallerCall
)

var (
	nativeCallTypes = []CallType{directCall, contractCall}
	erc20CallTypes  = []CallType{erc20Call, erc20CallerCall, erc20V5Call, erc20V5CallerCall}
)

// getCallArgs is a helper function to return the correct call arguments for a given call type.
//
// In case of a direct call to the precompile, the precompile's ABI is used. Otherwise, the
// ERC20CallerContract's ABI is used and the given contract address.
func (is *IntegrationTestSuite) getTxAndCallArgs(
	callType CallType,
	contractData ContractsData,
	methodName string,
	args ...interface{},
) (evmtypes.EvmTxArgs, factory.CallArgs) {
	cd := contractData.GetContractData(callType)

	txArgs := evmtypes.EvmTxArgs{
		To: &cd.Address,
	}

	callArgs := factory.CallArgs{
		ContractABI: cd.ABI,
		MethodName:  methodName,
		Args:        args,
	}

	return txArgs, callArgs
}

// ExpectedBalance is a helper struct to check the balances of accounts.
type ExpectedBalance struct {
	address  sdk.AccAddress
	expCoins sdk.Coins
}

// ExpectBalances is a helper function to check if the balances of the given accounts are as expected.
func (is *IntegrationTestSuite) ExpectBalances(expBalances []ExpectedBalance) {
	for _, expBalance := range expBalances {
		for _, expCoin := range expBalance.expCoins {
			coinBalance, err := is.handler.GetBalance(expBalance.address, expCoin.Denom)
			Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")
			Expect(coinBalance.Balance.Amount.Int64()).To(Equal(expCoin.Amount.Int64()), "expected different balance")
		}
	}
}

// ExpectBalancesForContract is a helper function to check expected balances for given accounts depending
// on the call type.
func (is *IntegrationTestSuite) ExpectBalancesForContract(callType CallType, contractData ContractsData, expBalances []ExpectedBalance) {
	switch {
	case slices.Contains(nativeCallTypes, callType):
		is.ExpectBalances(expBalances)
	case slices.Contains(erc20CallTypes, callType):
		is.ExpectBalancesForERC20(callType, contractData, expBalances)
	default:
		panic("unknown contract call type")
	}
}

// ExpectBalancesForERC20 is a helper function to check expected balances for given accounts
// when using the ERC20 contract.
func (is *IntegrationTestSuite) ExpectBalancesForERC20(callType CallType, contractData ContractsData, expBalances []ExpectedBalance) {
	contractABI := contractData.GetContractData(callType).ABI

	for _, expBalance := range expBalances {
		addr := common.BytesToAddress(expBalance.address.Bytes())
		txArgs, callArgs := is.getTxAndCallArgs(callType, contractData, "balanceOf", addr)

		passCheck := testutil.LogCheckArgs{ExpPass: true}

		_, ethRes, err := is.factory.CallContractAndCheckLogs(contractData.ownerPriv, txArgs, callArgs, passCheck)
		Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")

		var balance *big.Int
		err = contractABI.UnpackIntoInterface(&balance, "balanceOf", ethRes.Ret)
		Expect(err).ToNot(HaveOccurred(), "expected no error unpacking balance")
		Expect(balance.Int64()).To(Equal(expBalance.expCoins.AmountOf(is.tokenDenom).Int64()), "expected different balance")
	}
}

// expectSendAuthz is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount and optionally an access list.
//
// NOTE: This helper expects only one authorization to exist.
//
// NOTE 2: This mirrors the requireSendAuthz method but adapted to Ginkgo.
func (is *IntegrationTestSuite) expectSendAuthz(grantee, granter sdk.AccAddress, expAmount sdk.Coins) {
	authzs, err := is.handler.GetAuthorizations(grantee.String(), granter.String())
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking the authorization")
	Expect(authzs).To(HaveLen(1), "expected one authorization")

	sendAuthz, ok := authzs[0].(*banktypes.SendAuthorization)
	Expect(ok).To(BeTrue(), "expected send authorization")

	Expect(sendAuthz.SpendLimit).To(Equal(expAmount), "expected different spend limit amount")
}

// expectSendAuthzForERC20 is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount.
func (is *IntegrationTestSuite) expectSendAuthzForERC20(callType CallType, contractData ContractsData, grantee, granter common.Address, expAmount sdk.Coins) {
	contractABI := contractData.GetContractData(callType).ABI

	txArgs, callArgs := is.getTxAndCallArgs(callType, contractData, auth.AllowanceMethod, granter, grantee)

	passCheck := testutil.LogCheckArgs{ExpPass: true}

	_, ethRes, err := is.factory.CallContractAndCheckLogs(contractData.ownerPriv, txArgs, callArgs, passCheck)
	Expect(err).ToNot(HaveOccurred(), "expected no error getting allowance")

	var allowance *big.Int
	err = contractABI.UnpackIntoInterface(&allowance, "allowance", ethRes.Ret)
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking allowance")
	Expect(allowance.Int64()).To(Equal(expAmount.AmountOf(is.tokenDenom).Int64()), "expected different allowance")
}

// ExpectSendAuthzForContract is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount and optionally an access list.
//
// NOTE: This helper expects only one authorization to exist.
func (is *IntegrationTestSuite) ExpectSendAuthzForContract(
	callType CallType, contractData ContractsData, grantee, granter common.Address, expAmount sdk.Coins,
) {
	switch {
	case slices.Contains(nativeCallTypes, callType):
		is.expectSendAuthz(grantee.Bytes(), granter.Bytes(), expAmount)
	case slices.Contains(erc20CallTypes, callType):
		is.expectSendAuthzForERC20(callType, contractData, grantee, granter, expAmount)
	default:
		panic("unknown contract call type")
	}
}

// expectNoSendAuthz is a helper function to check that no SendAuthorization
// exists for a given grantee and granter combination.
func (is *IntegrationTestSuite) expectNoSendAuthz(grantee, granter sdk.AccAddress) {
	authzs, err := is.handler.GetAuthorizations(grantee.String(), granter.String())
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking the authorizations")
	Expect(authzs).To(HaveLen(0), "expected no authorizations")
}

// expectNoSendAuthzForERC20 is a helper function to check that no SendAuthorization
// exists for a given grantee and granter combination.
func (is *IntegrationTestSuite) expectNoSendAuthzForERC20(callType CallType, contractData ContractsData, grantee, granter common.Address) {
	is.expectSendAuthzForERC20(callType, contractData, grantee, granter, sdk.Coins{})
}

// ExpectNoSendAuthzForContract is a helper function to check that no SendAuthorization
// exists for a given grantee and granter combination.
func (is *IntegrationTestSuite) ExpectNoSendAuthzForContract(
	callType CallType, contractData ContractsData, grantee, granter common.Address,
) {
	switch {
	case slices.Contains(nativeCallTypes, callType):
		is.expectNoSendAuthz(grantee.Bytes(), granter.Bytes())
	case slices.Contains(erc20CallTypes, callType):
		is.expectNoSendAuthzForERC20(callType, contractData, grantee, granter)
	default:
		panic("unknown contract call type")
	}
}

// ExpectTrueToBeReturned is a helper function to check that the precompile returns true
// in the ethereum transaction response.
func (is *IntegrationTestSuite) ExpectTrueToBeReturned(res *evmtypes.MsgEthereumTxResponse, methodName string) {
	var ret bool
	err := is.precompile.UnpackIntoInterface(&ret, methodName, res.Ret)
	Expect(err).ToNot(HaveOccurred(), "expected no error unpacking")
	Expect(ret).To(BeTrue(), "expected true to be returned")
}

// ContractsData is a helper struct to hold the addresses and ABIs for the
// different contract instances that are subject to testing here.
type ContractsData struct {
	contractData map[CallType]ContractData
	ownerPriv    cryptotypes.PrivKey
}

// ContractData is a helper struct to hold the address and ABI for a given contract.
type ContractData struct {
	Address common.Address
	ABI     abi.ABI
}

// GetContractData is a helper function to return the contract data for a given call type.
func (cd ContractsData) GetContractData(callType CallType) ContractData {
	data, found := cd.contractData[callType]
	if !found {
		panic(fmt.Sprintf("no contract data found for call type: %d", callType))
	}
	return data
}

// fundWithTokens is a helper function for the scope of the ERC20 integration tests.
// Depending on the passed call type, it funds the given address with tokens either
// using the Bank module or by minting straight on the ERC20 contract.
func (is *IntegrationTestSuite) fundWithTokens(
	callType CallType,
	contractData ContractsData,
	receiver common.Address,
	fundCoins sdk.Coins,
) {
	Expect(fundCoins).To(HaveLen(1), "expected only one coin")
	Expect(fundCoins[0].Denom).To(Equal(is.tokenDenom),
		"this helper function only supports funding with the token denom in the context of these integration tests",
	)

	var err error

	switch {
	case slices.Contains(nativeCallTypes, callType):
		err = is.network.FundAccount(receiver.Bytes(), fundCoins)
	case slices.Contains(erc20CallTypes, callType):
		err = is.MintERC20(callType, contractData, receiver, fundCoins.AmountOf(is.tokenDenom).BigInt())
	default:
		panic("unknown contract call type")
	}

	Expect(err).ToNot(HaveOccurred(), "failed to fund account")
}

// MintERC20 is a helper function to mint tokens on the ERC20 contract.
//
// NOTE: we are checking that there was a Transfer event emitted (which happens on minting).
func (is *IntegrationTestSuite) MintERC20(callType CallType, contractData ContractsData, receiver common.Address, amount *big.Int) error {
	if callType == erc20V5CallerCall {
		// NOTE: When using the ERC20 caller contract, we must still mint from the actual ERC20 v5 contract.
		callType = erc20V5Call
	}
	abiEvents := contractData.GetContractData(callType).ABI.Events

	txArgs, callArgs := is.getTxAndCallArgs(callType, contractData, "mint", receiver, amount)

	mintCheck := testutil.LogCheckArgs{
		ABIEvents: abiEvents,
		ExpEvents: []string{erc20.EventTypeTransfer}, // NOTE: this event occurs when calling "mint" on ERC20s
		ExpPass:   true,
	}

	_, _, err := is.factory.CallContractAndCheckLogs(contractData.ownerPriv, txArgs, callArgs, mintCheck)

	return err
}
