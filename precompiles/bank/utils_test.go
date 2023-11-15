package bank_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v15/precompiles/bank"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
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

// mintAndSendCoin is a helper function to mint and send a coin to a given address.
func (s *PrecompileTestSuite) mintAndSendCoin(denom string, addr sdk.AccAddress, amount math.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(s.network.GetContext(), inflationtypes.ModuleName, addr, coins)
	s.Require().NoError(err)
}
