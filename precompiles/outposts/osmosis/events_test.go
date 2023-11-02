package osmosis_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	evmosutiltx "github.com/evmos/evmos/v15/testutil/tx"
)

func (s *PrecompileTestSuite) TestSwapEvent() {
	s.SetupTest()

	// random common.Address that represents the evmos ERC20 token address and
	// the IBC OSMO ERC20 token address.
	evmosAddress := evmosutiltx.GenerateAddress()
	osmoAddress := evmosutiltx.GenerateAddress()

	sender := s.keyring.GetAddr(0)
	receiver := "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2"
	transferAmount := int64(10)

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
			evmosAddress,
			osmoAddress,
			big.NewInt(transferAmount),
			receiver,
			func(input common.Address, output common.Address, amount *big.Int, receiver string) {
				stateDB := s.network.GetStateDB()
				swapLog := stateDB.Logs()[0]
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
					uint64(s.network.GetContext().BlockHeight()),
					"require event block height equal to context block height",
				)

				// Check for swap specific information in the event
				var swapEvent osmosis.EventSwap
				err := cmn.UnpackLog(s.precompile.ABI, &swapEvent, osmosis.EventTypeSwap, *swapLog)
				s.Require().NoError(err)
				s.Require().Equal(
					sender,
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
			err := s.network.NextBlock()
			s.Require().NoError(err)

			err = s.precompile.EmitSwapEvent(
				s.network.GetContext(),
				s.network.GetStateDB(),
				sender,
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
