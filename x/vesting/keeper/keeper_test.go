package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v17/app"
	"github.com/evmos/evmos/v17/encoding"
	"github.com/evmos/evmos/v17/x/vesting/keeper"
	vestingtypes "github.com/evmos/evmos/v17/x/vesting/types"
)

func (suite *KeeperTestSuite) TestNewKeeper() {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(vestingtypes.StoreKey)

	testcases := []struct {
		name      string
		authority sdk.AccAddress
		expPass   bool
	}{
		{
			name:      "valid authority format",
			authority: sdk.AccAddress(suite.address.Bytes()),
			expPass:   true,
		},
		{
			name:      "empty authority",
			authority: []byte{},
			expPass:   false,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			if tc.expPass {
				newKeeper := keeper.NewKeeper(
					storeKey,
					tc.authority,
					cdc,
					suite.app.AccountKeeper,
					suite.app.BankKeeper,
					suite.app.DistrKeeper,
					suite.app.StakingKeeper,
					suite.app.GovKeeper,
				)
				suite.Require().NotNil(newKeeper)
			} else {
				suite.Require().PanicsWithError("addresses cannot be empty: unknown address", func() {
					_ = keeper.NewKeeper(
						storeKey,
						tc.authority,
						cdc,
						suite.app.AccountKeeper,
						suite.app.BankKeeper,
						suite.app.DistrKeeper,
						suite.app.StakingKeeper,
						suite.app.GovKeeper,
					)
				})
			}
		})
	}
}
