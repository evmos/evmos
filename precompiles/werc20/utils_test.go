package werc20_test

import (
	"math/big"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// callType constants to differentiate between direct calls and calls through a contract.
const (
	directCall = iota + 1
	contractCall
	erc20Call
)

// ContractData is a helper struct to hold the addresses and ABIs for the
// different contract instances that are subject to testing here.
type ContractData struct {
	ownerPriv cryptotypes.PrivKey

	erc20Addr      common.Address
	erc20ABI       abi.ABI
	contractAddr   common.Address
	contractABI    abi.ABI
	precompileAddr common.Address
	precompileABI  abi.ABI
}

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
	}

	callArgs.MethodName = methodName
	callArgs.Args = args

	return txArgs, callArgs
}

func (s *PrecompileTestSuite) checkBalances(failCheck testutil.LogCheckArgs, sender keyring.Key, contractData ContractData) {
	balanceCheck := failCheck.WithExpPass(true)
	txArgs, balancesArgs := s.getTxAndCallArgs(erc20Call, contractData, erc20.BalanceOfMethod, sender.Addr)

	_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, balancesArgs, balanceCheck)
	s.Require().NoError(err, "failed to call contract")

	// Check the balance in the bank module is the same as calling `balanceOf` on the precompile
	balanceAfter := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sender.AccAddr, s.bondDenom)

	var erc20Balance *big.Int
	err = s.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
	s.Require().NoError(err, "failed to unpack result")
	s.Require().Equal(erc20Balance, balanceAfter.Amount.BigInt(), "expected different balance")
}
