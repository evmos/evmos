package types

//goland:noinspection SpellCheckingInspection
import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	feemarketkeeper "github.com/evmos/evmos/v16/x/feemarket/keeper"
	ibctransferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
)

type ChainApp interface {
	App() abci.Application
	BaseApp() *baseapp.BaseApp
	IbcTestingApp() ibctesting.TestingApp
	InterfaceRegistry() codectypes.InterfaceRegistry

	// Keepers

	AccountKeeper() *authkeeper.AccountKeeper
	BankKeeper() bankkeeper.Keeper
	DistributionKeeper() distkeeper.Keeper
	Erc20Keeper() *erc20keeper.Keeper
	EvmKeeper() *evmkeeper.Keeper
	FeeMarketKeeper() *feemarketkeeper.Keeper
	GovKeeper() *govkeeper.Keeper
	IbcTransferKeeper() *ibctransferkeeper.Keeper
	IbcKeeper() *ibckeeper.Keeper
	SlashingKeeper() *slashingkeeper.Keeper
	StakingKeeper() *stakingkeeper.Keeper

	// Tx

	FundAccount(ctx sdk.Context, account *TestAccount, amounts sdk.Coins) error
}
