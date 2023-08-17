package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/x/vesting/keeper"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
)

func (s *KeeperTestSuite) TestNewKeeper() {
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
			authority: sdk.AccAddress(s.address.Bytes()),
			expPass:   true,
		},
		{
			name:      "empty authority",
			authority: []byte{},
			expPass:   false,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			if tc.expPass {
				newKeeper := keeper.NewKeeper(
					storeKey,
					tc.authority,
					cdc,
					s.app.AccountKeeper,
					s.app.BankKeeper,
					s.app.DistrKeeper,
					s.app.StakingKeeper,
				)
				s.Require().NotNil(newKeeper)
			} else {
				s.Require().PanicsWithError("addresses cannot be empty: unknown address", func() {
					_ = keeper.NewKeeper(
						storeKey,
						tc.authority,
						cdc,
						s.app.AccountKeeper,
						s.app.BankKeeper,
						s.app.DistrKeeper,
						s.app.StakingKeeper,
					)
				})
			}
		})
	}
}
