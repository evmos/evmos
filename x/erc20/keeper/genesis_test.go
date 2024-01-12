package keeper_test

import (
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v16/x/erc20/types"
)

const (
	osmoERC20ContractAddr = "0x5dCA2483280D9727c80b5518faC4556617fb19ZZ"
	junoERC20ContractAddr = "0x5db67696C3c088DfBf588d3dd849f44266ff0ffa"
)

var (
	osmoDenomTrace = transfertypes.DenomTrace{
		BaseDenom: "uosmo",
		Path:      "transfer/channel-0",
	}
	junoDenomTrace = transfertypes.DenomTrace{
		BaseDenom: "ujuno",
		Path:      "transfer/channel-1",
	}
	invalidDenomTrace = transfertypes.DenomTrace{
		BaseDenom: "osmo", // missing the 'u' prefix
		Path:      "transfer/channel-0",
	}
	invalidDenomTrace2 = transfertypes.DenomTrace{
		BaseDenom: "u", // denom trace is too short
		Path:      "transfer/channel-0",
	}
)

func (suite *KeeperTestSuite) TestSetGenesisTokenPairs() {
	testCases := []struct {
		name      string
		pairs     []types.TokenPair
		malleate  func()
		expFail   bool
		expErrMsg string
	}{
		{
			"no-op: no token pairs",
			[]types.TokenPair{},
			nil,
			false,
			"",
		},
		{
			"fail: invalid denom",
			[]types.TokenPair{
				{
					Erc20Address:  osmoERC20ContractAddr,
					Denom:         "uosmo",
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			nil,
			true,
			"denom is not an IBC voucher",
		},
		{
			"fail: custom genesis - denom trace not in genesis",
			[]types.TokenPair{
				{
					Erc20Address:  osmoERC20ContractAddr,
					Denom:         osmoDenomTrace.IBCDenom(),
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			nil,
			true,
			"denom trace not found",
		},
		{
			"fail: custom genesis with invalid denom traces in genesis",
			[]types.TokenPair{
				{
					Erc20Address:  osmoERC20ContractAddr,
					Denom:         invalidDenomTrace.IBCDenom(),
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			func() {
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, invalidDenomTrace)
			},
			true,
			"invalid base denomination",
		},
		{
			"fail: custom genesis with invalid denom traces in genesis - base denom too short",
			[]types.TokenPair{
				{
					Erc20Address:  osmoERC20ContractAddr,
					Denom:         invalidDenomTrace2.IBCDenom(),
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			func() {
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, invalidDenomTrace2)
			},
			true,
			"denom trace base denom is too short",
		},
		{
			"success: custom genesis with denom traces in genesis",
			[]types.TokenPair{
				{
					Erc20Address:  junoERC20ContractAddr,
					Denom:         junoDenomTrace.IBCDenom(),
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
				{
					Erc20Address:  osmoERC20ContractAddr,
					Denom:         osmoDenomTrace.IBCDenom(),
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			func() {
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, osmoDenomTrace)
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, junoDenomTrace)
			},
			false,
			"",
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset
		if tc.malleate != nil {
			tc.malleate()
		}
		err := suite.app.Erc20Keeper.SetGenesisTokenPairs(suite.ctx, tc.pairs)

		if !tc.expFail {
			suite.Require().NoError(err)
			tokenPairs := suite.app.Erc20Keeper.GetTokenPairs(suite.ctx)
			suite.Require().Equal(tc.pairs, tokenPairs)
			// check ERC20 contracts were created successfully
			for _, p := range tc.pairs {
				acc := suite.app.EvmKeeper.GetAccount(suite.ctx, common.HexToAddress(p.Erc20Address))
				suite.Require().True(acc.IsContract())
				// check bank metadata was stored properly
				meta, found := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, p.Denom)
				suite.Require().True(found)
				suite.Require().Len(meta.DenomUnits, 2)

				switch meta.Base {
				case osmoDenomTrace.IBCDenom():
					suite.Require().Equal(meta.DenomUnits[0].Aliases, []string{"uosmo"})
					suite.Require().Equal(meta.DenomUnits[1].Denom, "osmo")
					suite.Require().Equal(meta.DenomUnits[1].Exponent, uint32(6))
					suite.Require().Equal(meta.Display, "osmo")
					suite.Require().Equal(meta.Name, "Osmo")
					suite.Require().Equal(meta.Symbol, "OSMO")
				case junoDenomTrace.IBCDenom():
					suite.Require().Equal(meta.DenomUnits[0].Aliases, []string{"ujuno"})
					suite.Require().Equal(meta.DenomUnits[1].Denom, "juno")
					suite.Require().Equal(meta.DenomUnits[1].Exponent, uint32(6))
					suite.Require().Equal(meta.Display, "juno")
					suite.Require().Equal(meta.Name, "Juno")
					suite.Require().Equal(meta.Symbol, "JUNO")
				}
			}
		} else {
			suite.Require().Error(err)
			suite.Require().Contains(err.Error(), tc.expErrMsg)
		}
	}
}
