package stride_test

import (
	"fmt"
	"math/big"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/stride"
	"github.com/evmos/evmos/v15/utils"
)

const receiver = "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0"

func (s *PrecompileTestSuite) TestLiquidStakeEvent() {
	denomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, utils.BaseDenom)
	tokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, denomID)
	s.Require().True(ok, "expected token pair to be found")

	//nolint:dupl
	testCases := []struct {
		name      string
		postCheck func()
	}{
		{
			"success",
			func() {
				liquidStakeLog := s.stateDB.Logs()[0]
				s.Require().Equal(liquidStakeLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[stride.EventTypeLiquidStake]
				s.Require().Equal(event.ID, common.HexToHash(liquidStakeLog.Topics[0].Hex()))
				s.Require().Equal(liquidStakeLog.BlockNumber, uint64(s.ctx.BlockHeight()))

				var liquidStakeEvent stride.EventLiquidStake
				err := cmn.UnpackLog(s.precompile.ABI, &liquidStakeEvent, stride.EventTypeLiquidStake, *liquidStakeLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), liquidStakeEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), liquidStakeEvent.Token)
				s.Require().Equal(big.NewInt(1e18), liquidStakeEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.precompile.EmitLiquidStakeEvent(s.ctx, s.stateDB, s.address, common.HexToAddress(tokenPair.Erc20Address), big.NewInt(1e18))
			s.Require().NoError(err)
			tc.postCheck()
		})
	}
}

func (s *PrecompileTestSuite) TestRedeemEvent() {
	bondDenom := s.app.StakingKeeper.BondDenom(s.ctx)
	denomTrace := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: "st" + bondDenom,
	}

	stEvmos := denomTrace.IBCDenom()

	denomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, stEvmos)
	tokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, denomID)
	s.Require().True(ok, "expected token pair to be found")

	//nolint:dupl
	testCases := []struct {
		name      string
		postCheck func()
	}{
		{
			"success",
			func() {
				redeemLog := s.stateDB.Logs()[0]
				s.Require().Equal(redeemLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[stride.EventTypeRedeem]
				s.Require().Equal(event.ID, common.HexToHash(redeemLog.Topics[0].Hex()))
				s.Require().Equal(redeemLog.BlockNumber, uint64(s.ctx.BlockHeight()))

				var redeemEvent stride.EventRedeem
				err := cmn.UnpackLog(s.precompile.ABI, &redeemEvent, stride.EventTypeRedeem, *redeemLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), redeemEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), redeemEvent.Token)
				s.Require().Equal(big.NewInt(1e18), redeemEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.precompile.EmitRedeemEvent(s.ctx, s.stateDB, s.address, common.HexToAddress(tokenPair.Erc20Address), receiver, big.NewInt(1e18))
			s.Require().NoError(err)
			tc.postCheck()
		})
	}
}
