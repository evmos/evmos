package osmosis_test

import (
	"fmt"
	"math/big"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	"github.com/evmos/evmos/v15/precompiles/outposts/stride"
	"github.com/evmos/evmos/v15/utils"
)

const receiver = "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0"

func (s *PrecompileTestSuite) TestSwapEvent() {
	// Retrieve Evmos token information useful for the testing
	evmosDenomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, utils.BaseDenom)
	evmosTokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, evmosDenomID)
	s.Require().True(ok, "expected evmos token pair to be found")

	// Retrieve Osmo token information useful for the testing
	osmoIBCDenom := utils.ComputeIBCDenom(s.precompile.portId, s.precompile.channelID, osmosis.OsmosisDenom)
	osmoDenomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, utils.BaseDenom)
	osmoTokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, osmoDenomID)
	s.Require().True(ok, "expected osmo token pair to be found")

	testCases := []struct {
		name      string
		input     common.Address
		output    common.Address
		amount    *big.Int
		receiver  string
		postCheck func()
	}{
		{
			"success",
            evmosTokenPair.GetERC20Contract(),
            osmoTokenPair.GetERC20Contract(),
            big.NewInt(10)
			func() {
				swapLog := s.stateDB.Logs()[0]
				s.Require().Equal(swapLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[osmosis.EventTypeSwap]
				s.Require().Equal(event.ID, common.HexToHash(swapLog.Topics[0].Hex()))
				s.Require().Equal(swapLog.BlockNumber, uint64(s.ctx.BlockHeight()))

				var swapEvent osmosis.EventSwap
				err := cmn.UnpackLog(s.precompile.ABI, &swapEvent, osmosis.EventTypeSwap, *swapLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), swapEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), swapEvent.Input)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), swapEvent.Output)
				s.Require().Equal(big.NewInt(0), swapEvent.Amount)
				s.Require().Equal("sender", swapEvent.Sender)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.precompile.EmitSwapEvent(
				s.ctx,
				s.stateDB,
				s.address,
				tc.input,
				tc.output,
				tc.amount,
				tc.receiver,
			)(s.ctx, s.stateDB, s.address, common.HexToAddress(tokenPair.Erc20Address), big.NewInt(1e18))
			s.Require().NoError(err)
			tc.postCheck()
		})
	}
}
