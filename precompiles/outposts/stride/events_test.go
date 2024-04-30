package stride_test

import (
	"fmt"
	"math/big"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/outposts/stride"
	"github.com/evmos/evmos/v18/utils"
)

const receiver = "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0"

func (s *PrecompileTestSuite) TestLiquidStakeEvent() {
	ctx := s.network.GetContext()
	stateDB := s.network.GetStateDB()
	denomID := s.network.App.Erc20Keeper.GetDenomMap(ctx, utils.BaseDenom)
	tokenPair, ok := s.network.App.Erc20Keeper.GetTokenPair(ctx, denomID)
	s.Require().True(ok, "expected token pair to be found")

	testCases := []struct {
		name      string
		postCheck func()
	}{
		{
			"success",
			func() {
				liquidStakeLog := stateDB.Logs()[0]
				s.Require().Equal(liquidStakeLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[stride.EventTypeLiquidStake]
				s.Require().Equal(event.ID, common.HexToHash(liquidStakeLog.Topics[0].Hex()))
				s.Require().Equal(liquidStakeLog.BlockNumber, uint64(ctx.BlockHeight()))

				var liquidStakeEvent stride.EventLiquidStake
				err := cmn.UnpackLog(s.precompile.ABI, &liquidStakeEvent, stride.EventTypeLiquidStake, *liquidStakeLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.keyring.GetAccAddr(0)), liquidStakeEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), liquidStakeEvent.Token)
				s.Require().Equal(big.NewInt(1e18), liquidStakeEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.precompile.EmitLiquidStakeEvent(ctx, stateDB, s.keyring.GetAddr(0), common.HexToAddress(tokenPair.Erc20Address), big.NewInt(1e18))
			s.Require().NoError(err)
			tc.postCheck()
		})
	}
}

func (s *PrecompileTestSuite) TestRedeemEvent() {
	ctx := s.network.GetContext()
	stateDB := s.network.GetStateDB()
	bondDenom := s.network.App.StakingKeeper.BondDenom(ctx)
	denomTrace := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: "st" + bondDenom,
	}

	stEvmos := denomTrace.IBCDenom()

	denomID := s.network.App.Erc20Keeper.GetDenomMap(ctx, stEvmos)
	tokenPair, ok := s.network.App.Erc20Keeper.GetTokenPair(ctx, denomID)
	s.Require().True(ok, "expected token pair to be found")

	testCases := []struct {
		name      string
		postCheck func()
	}{
		{
			"success",
			func() {
				redeemLog := stateDB.Logs()[0]
				s.Require().Equal(redeemLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[stride.EventTypeRedeemStake]
				s.Require().Equal(event.ID, common.HexToHash(redeemLog.Topics[0].Hex()))
				s.Require().Equal(redeemLog.BlockNumber, uint64(ctx.BlockHeight()))

				var redeemEvent stride.EventRedeem
				err := cmn.UnpackLog(s.precompile.ABI, &redeemEvent, stride.EventTypeRedeemStake, *redeemLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.keyring.GetAccAddr(0)), redeemEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), redeemEvent.Token)
				s.Require().Equal(s.keyring.GetAddr(0), redeemEvent.Receiver)
				s.Require().Equal(receiver, redeemEvent.StrideForwarder)
				s.Require().Equal(big.NewInt(1e18), redeemEvent.Amount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.precompile.EmitRedeemStakeEvent(
				ctx,
				stateDB,
				s.keyring.GetAddr(0),
				common.HexToAddress(tokenPair.Erc20Address),
				s.keyring.GetAddr(0),
				receiver,
				big.NewInt(1e18),
			)
			s.Require().NoError(err)
			tc.postCheck()
		})
	}
}
