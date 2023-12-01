package bank_test

import (
	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/bank"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// setupBankPrecompile is a helper function to set up an instance of the Bank precompile for
// a given token denomination.
func (s *PrecompileTestSuite) setupBankPrecompile() *bank.Precompile {
	precompile, err := bank.NewPrecompile(
		s.network.App.BankKeeper,
		s.network.App.Erc20Keeper,
	)

	s.Require().NoError(err, "failed to create bank precompile")

	return precompile
}

// setupBankPrecompile is a helper function to set up an instance of the Bank precompile for
// a given token denomination.
func (is *IntegrationTestSuite) setupBankPrecompile() *bank.Precompile {
	precompile, err := bank.NewPrecompile(
		is.network.App.BankKeeper,
		is.network.App.Erc20Keeper,
	)
	Expect(err).ToNot(HaveOccurred(), "failed to create bank precompile")
	return precompile
}

// mintAndSendXMPLCoin is a helper function to mint and send a coin to a given address.
func (s *PrecompileTestSuite) mintAndSendXMPLCoin(addr sdk.AccAddress, amount math.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(s.tokenDenom, amount))
	err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), inflationtypes.ModuleName, addr, coins)
	s.Require().NoError(err)
}

// mintAndSendXMPLCoin is a helper function to mint and send a coin to a given address.
func (is *IntegrationTestSuite) mintAndSendXMPLCoin(addr sdk.AccAddress, amount math.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(is.tokenDenom, amount))
	err := is.network.App.BankKeeper.MintCoins(is.network.GetContext(), inflationtypes.ModuleName, coins)
	Expect(err).ToNot(HaveOccurred())
	err = is.network.App.BankKeeper.SendCoinsFromModuleToAccount(is.network.GetContext(), inflationtypes.ModuleName, addr, coins)
	Expect(err).ToNot(HaveOccurred())
}

// callType constants to differentiate between direct calls and calls through a contract.
const (
	directCall = iota + 1
	contractCall
)

// ContractData is a helper struct to hold the addresses and ABIs for the
// different contract instances that are subject to testing here.
type ContractData struct {
	ownerPriv cryptotypes.PrivKey

	contractAddr   common.Address
	contractABI    abi.ABI
	precompileAddr common.Address
	precompileABI  abi.ABI
}

// getCallArgs is a helper function to return the correct call arguments for a given call type.
// In case of a direct call to the precompile, the precompile's ABI is used. Otherwise a caller contract is used.
func getTxAndCallArgs(
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
	}

	callArgs.MethodName = methodName
	callArgs.Args = args

	return txArgs, callArgs
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
