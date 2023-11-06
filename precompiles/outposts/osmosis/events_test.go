package osmosis_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	evmosutiltx "github.com/evmos/evmos/v15/testutil/tx"
	"github.com/evmos/evmos/v15/x/evm/statedb"

	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
)

func (s *PrecompileTestSuite) TestSwapEvent() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	precompile, err := osmosis.NewPrecompile(
		portID,
		channelID,
		osmosis.XCSContract,
		unitNetwork.App.BankKeeper,
		unitNetwork.App.TransferKeeper,
		unitNetwork.App.StakingKeeper,
		unitNetwork.App.Erc20Keeper,
	)
	s.Require().NoError(err)
	// random common.Address that represents the evmos ERC20 token address and
	// the IBC OSMO ERC20 token address.
	evmosAddress := evmosutiltx.GenerateAddress()
	osmoAddress := evmosutiltx.GenerateAddress()

	sender := keyring.GetAddr(0)
	receiver := "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2"
	transferAmount := int64(10)

	testCases := []struct {
		name      string
		input     common.Address
		output    common.Address
		amount    *big.Int
		receiver  string
		postCheck func(input common.Address, output common.Address, amount *big.Int, receiver string, stateDB *statedb.StateDB)
	}{
		{
			"pass - correct event emitted",
			evmosAddress,
			osmoAddress,
			big.NewInt(transferAmount),
			receiver,
			func(input common.Address, output common.Address, amount *big.Int, receiver string, stateDB *statedb.StateDB) {
				s.Require().Len(stateDB.Logs(), 1, "expected one log in the stateDB")

				swapLog := stateDB.Logs()[0]
				s.Require().Equal(
					swapLog.Address,
					precompile.Address(),
					"expected first log address equal to osmosis outpost precompile",
				)
				event := precompile.ABI.Events[osmosis.EventTypeSwap]
				s.Require().Equal(
					event.ID,
					common.HexToHash(swapLog.Topics[0].Hex()),
					"expected event signature equal to osmosis outpost event signature",
				)
				s.Require().Equal(
					swapLog.BlockNumber,
					uint64(unitNetwork.GetContext().BlockHeight()),
					"require event block height equal to context block height",
				)

				// Check for swap specific information in the event
				var swapEvent osmosis.EventSwap
				err := cmn.UnpackLog(precompile.ABI, &swapEvent, osmosis.EventTypeSwap, *swapLog)
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
			err := unitNetwork.NextBlock()
			s.Require().NoError(err)

			stateDB := unitNetwork.GetStateDB()

			err = precompile.EmitSwapEvent(
				unitNetwork.GetContext(),
				stateDB,
				sender,
				tc.input,
				tc.output,
				tc.amount,
				tc.receiver,
			)
			s.Require().NoError(err)
			tc.postCheck(tc.input, tc.output, tc.amount, tc.receiver, stateDB)
		})
	}
}
