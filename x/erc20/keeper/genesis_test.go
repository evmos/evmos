package keeper_test

import (
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v16/x/erc20/types"
)

const (
	osmoIBCDenom          = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	osmoERC20ContractAddr = "0x5dCA2483280D9727c80b5518faC4556617fb19ZZ"
)

var osmoDenomTrace = transfertypes.DenomTrace{
	BaseDenom: "uosmo",
	Path:      "transfer/channel-0",
}

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
					Denom:         osmoIBCDenom,
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			nil,
			true,
			"denom trace not found",
		},
		{
			"success: custom genesis with denom trace in genesis",
			[]types.TokenPair{
				{
					Erc20Address:  osmoERC20ContractAddr,
					Denom:         osmoIBCDenom,
					Enabled:       true,
					ContractOwner: types.OWNER_MODULE,
				},
			},
			func() {
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, osmoDenomTrace)
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
			// check ERC20 contract was created successfully
			if len(tc.pairs) > 0 {
				acc := suite.app.EvmKeeper.GetAccount(suite.ctx, common.HexToAddress(osmoERC20ContractAddr))
				suite.Require().True(acc.IsContract())
			}
		} else {
			suite.Require().Error(err)
			suite.Require().Contains(err.Error(), tc.expErrMsg)
		}
	}
}
