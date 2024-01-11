package types

//goland:noinspection SpellCheckingInspection
import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	chainapp "github.com/evmos/evmos/v16/app"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
)

var _ ChainApp = &chainAppImp{}

type chainAppImp struct {
	app *chainapp.Evmos
}

func (c chainAppImp) App() abci.Application {
	return c.app
}

func (c chainAppImp) BaseApp() *baseapp.BaseApp {
	return c.app.BaseApp
}

func (c chainAppImp) IbcTestingApp() ibctesting.TestingApp {
	return c.app
}

func (c chainAppImp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return c.app.InterfaceRegistry()
}

func (c chainAppImp) FundAccount(ctx sdk.Context, account *TestAccount, amounts sdk.Coins) error {
	if err := c.BankKeeper().MintCoins(ctx, inflationtypes.ModuleName, amounts); err != nil {
		return err
	}

	return c.BankKeeper().SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, account.GetCosmosAddress(), amounts)
}
