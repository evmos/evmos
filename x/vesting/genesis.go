package vesting

import (
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/tharsis/evmos/x/vesting/keeper"
	"github.com/tharsis/evmos/x/vesting/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ac authkeeper.AccountKeeper,
	bk types.BankKeeper,
	_ types.GenesisState,
) {

	vestungAccs := []struct {
		fromAddress    string
		toAddress      string
		startTime      time.Time
		lockupPeriods  []sdkvesting.Period
		vestingPeriods []sdkvesting.Period
	}{
		{
			"evmos1wmzx0dj2emwxh5q3a9znwjtgrne0aadh5c7lxl",
			"evmos160ka7ccq7tn2355lppetq262sgh6r0u379pr4g",
			ctx.BlockTime(),
			[]sdkvesting.Period{{Length: 120, Amount: sdk.NewCoins(sdk.NewInt64Coin("aevmos", 1000))}},
			[]sdkvesting.Period{
				{Length: 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("aevmos", 250))},
				{Length: 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("aevmos", 250))},
				{Length: 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("aevmos", 250))},
				{Length: 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("aevmos", 250))},
			},
		},
	}
	for _, va := range vestungAccs {
		from, err := sdk.AccAddressFromBech32(va.fromAddress)
		if err != nil {
			panic(err)
		}

		to, err := sdk.AccAddressFromBech32(va.toAddress)
		if err != nil {
			panic(err)
		}
		vestingCoins := sdk.NewCoins()
		for _, period := range va.vestingPeriods {
			vestingCoins = vestingCoins.Add(period.Amount...)
		}

		lockupCoins := sdk.NewCoins()
		for _, period := range va.lockupPeriods {
			lockupCoins = lockupCoins.Add(period.Amount...)
		}

		// If lockup absent, default to an instant unlock schedule
		if !vestingCoins.IsZero() && len(va.lockupPeriods) == 0 {
			va.lockupPeriods = []sdkvesting.Period{
				{Length: 0, Amount: vestingCoins},
			}
			lockupCoins = vestingCoins
		}

		// If vesting absent, default to an instant vesting schedule
		if !lockupCoins.IsZero() && len(va.vestingPeriods) == 0 {
			va.vestingPeriods = []sdkvesting.Period{
				{Length: 0, Amount: lockupCoins},
			}
			vestingCoins = lockupCoins
		}

		// The vesting and lockup schedules must describe the same total amount.
		// IsEqual can panic, so use (a == b) <=> (a <= b && b <= a).
		if !(vestingCoins.IsAllLTE(lockupCoins) && lockupCoins.IsAllLTE(vestingCoins)) {
			panic("different schedule amounts")
		}

		// Add Grant if vesting account exists, "merge" is true and funder is correct.
		// Otherwise create a new Clawback Vesting Account
		madeNewAcc := false
		acc := ac.GetAccount(ctx, to)
		var vestingAcc *types.ClawbackVestingAccount

		if acc != nil {
			panic("account already exists")
		}
		baseAcc := authtypes.NewBaseAccountWithAddress(to)
		vestingAcc = types.NewClawbackVestingAccount(
			baseAcc,
			from,
			vestingCoins,
			va.startTime,
			va.lockupPeriods,
			va.vestingPeriods,
		)
		acc = ac.NewAccount(ctx, vestingAcc)
		ac.SetAccount(ctx, acc)
		madeNewAcc = true

		// TODO check what this is for
		if madeNewAcc {
			defer func() {
				telemetry.IncrCounter(1, "new", "account")

				for _, a := range vestingCoins {
					if a.Amount.IsInt64() {
						telemetry.SetGaugeWithLabels(
							[]string{"tx", "msg", "create_clawback_vesting_account"},
							float32(a.Amount.Int64()),
							[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
						)
					}
				}
			}()
		}

		// Send coins from the funder to vesting account
		if err := bk.SendCoins(ctx, from, to, vestingCoins); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{}
}
