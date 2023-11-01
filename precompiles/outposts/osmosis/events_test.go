package osmosis_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	"github.com/evmos/evmos/v15/utils"
)

const (
	receiver       = "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2"
	transferAmount = 10
)

func (s *PrecompileTestSuite) TestSwapEvent() {
	// Retrieve Evmos token information useful for the testing
	evmosDenomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, utils.BaseDenom)
	evmosTokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, evmosDenomID)
	s.Require().True(ok, "expected evmos token pair to be found")

	// Retrieve Osmo token information useful for the testing
	osmoIBCDenom := utils.ComputeIBCDenom(portID, channelID, osmosis.OsmosisDenom)
	osmoDenomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, osmoIBCDenom)
	osmoTokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, osmoDenomID)
	s.Require().True(ok, "expected osmo token pair to be found")

	testCases := []struct {
		name      string
		input     common.Address
		output    common.Address
		amount    *big.Int
		receiver  string
		postCheck func(input common.Address, output common.Address, amount *big.Int, receiver string)
	}{
		{
			"pass - correct event emitted",
			evmosTokenPair.GetERC20Contract(),
			osmoTokenPair.GetERC20Contract(),
			big.NewInt(transferAmount),
			receiver,
			func(input common.Address, output common.Address, amount *big.Int, receiver string) {
				swapLog := s.stateDB.Logs()[0]
				s.Require().Equal(
					swapLog.Address,
					s.precompile.Address(),
					"expected first log address equal to osmosis outpost precompile",
				)
				event := s.precompile.ABI.Events[osmosis.EventTypeSwap]
				s.Require().Equal(
					event.ID,
					common.HexToHash(swapLog.Topics[0].Hex()),
					"expected event signature equal to osmosis outpost event signature",
				)
				s.Require().Equal(
					swapLog.BlockNumber,
					uint64(s.ctx.BlockHeight()),
					"require event block height equal to context block height",
				)

				// Check for swap specific information in the event
				var swapEvent osmosis.EventSwap
				err := cmn.UnpackLog(s.precompile.ABI, &swapEvent, osmosis.EventTypeSwap, *swapLog)
				s.Require().NoError(err)
				s.Require().Equal(
					s.address,
					swapEvent.Sender,
					"expected a different sender in the event log",
				)
				s.Require().Equal(
					input,
					swapEvent.Input,
					"expected a different input value in the event",
				)
				s.Require().Equal(
					output,
					swapEvent.Output,
					"expected a different output value in the event",
				)
				s.Require().Equal(
					amount,
					swapEvent.Amount,
					"expected a different amount in the event log",
				)
				s.Require().Equal(
					receiver,
					swapEvent.Receiver,
					"expected a different receiver value in the event",
				)
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
			)
			s.Require().NoError(err)
			tc.postCheck(tc.input, tc.output, tc.amount, tc.receiver)
		})
	}
}
