package types

//goland:noinspection SpellCheckingInspection
import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	feemarketkeeper "github.com/evmos/evmos/v16/x/feemarket/keeper"
	ibctransferkeeper "github.com/evmos/evmos/v16/x/ibc/transfer/keeper"
)

func (c chainAppImp) AccountKeeper() *authkeeper.AccountKeeper {
	return &c.app.AccountKeeper
}

func (c chainAppImp) BankKeeper() bankkeeper.Keeper {
	return c.app.BankKeeper
}

func (c chainAppImp) DistributionKeeper() distkeeper.Keeper {
	return c.app.DistrKeeper
}

func (c chainAppImp) Erc20Keeper() *erc20keeper.Keeper {
	return &c.app.Erc20Keeper
}

func (c chainAppImp) EvmKeeper() *evmkeeper.Keeper {
	return c.app.EvmKeeper
}

func (c chainAppImp) FeeMarketKeeper() *feemarketkeeper.Keeper {
	return &c.app.FeeMarketKeeper
}

func (c chainAppImp) GovKeeper() *govkeeper.Keeper {
	return &c.app.GovKeeper
}

func (c chainAppImp) IbcTransferKeeper() *ibctransferkeeper.Keeper {
	return &c.app.TransferKeeper
}

func (c chainAppImp) IbcKeeper() *ibckeeper.Keeper {
	return c.app.IBCKeeper
}

func (c chainAppImp) SlashingKeeper() *slashingkeeper.Keeper {
	return &c.app.SlashingKeeper
}

func (c chainAppImp) StakingKeeper() *stakingkeeper.Keeper {
	return &c.app.StakingKeeper
}
