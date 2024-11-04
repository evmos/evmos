package erc20_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v20/precompiles/authorization"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	erc20precompile "github.com/evmos/evmos/v20/precompiles/erc20"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
)

//nolint:dupl // this is not a duplicate of the approval events test
func (s *PrecompileTestSuite) TestEmitTransferEvent() {
	testcases := []struct {
		name   string
		from   common.Address
		to     common.Address
		amount *big.Int
	}{
		{
			name:   "pass",
			from:   utiltx.GenerateAddress(),
			to:     utiltx.GenerateAddress(),
			amount: big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB := s.network.GetStateDB()

			err := s.precompile.EmitTransferEvent(
				s.network.GetContext(), stateDB, tc.from, tc.to, tc.amount,
			)
			s.Require().NoError(err, "expected transfer event to be emitted successfully")

			log := stateDB.Logs()[0]
			s.Require().Equal(log.Address, s.precompile.Address())

			// Check event signature matches the one emitted
			event := s.precompile.ABI.Events[erc20precompile.EventTypeTransfer]
			s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
			s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight())) //nolint:gosec // G115

			// Check the fully unpacked event matches the one emitted
			var transferEvent erc20precompile.EventTransfer
			err = cmn.UnpackLog(s.precompile.ABI, &transferEvent, erc20precompile.EventTypeTransfer, *log)
			s.Require().NoError(err, "unable to unpack log into transfer event")

			s.Require().Equal(tc.from, transferEvent.From, "expected different from address")
			s.Require().Equal(tc.to, transferEvent.To, "expected different to address")
			s.Require().Equal(tc.amount, transferEvent.Value, "expected different amount")
		})
	}
}

//nolint:dupl // this is not a duplicate of the transfer events test
func (s *PrecompileTestSuite) TestEmitApprovalEvent() {
	testcases := []struct {
		name    string
		owner   common.Address
		spender common.Address
		amount  *big.Int
	}{
		{
			name:    "pass",
			owner:   utiltx.GenerateAddress(),
			spender: utiltx.GenerateAddress(),
			amount:  big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			err := s.precompile.EmitApprovalEvent(
				s.network.GetContext(), stateDB, tc.owner, tc.spender, tc.amount,
			)
			s.Require().NoError(err, "expected approval event to be emitted successfully")

			log := stateDB.Logs()[0]
			s.Require().Equal(log.Address, s.precompile.Address())

			// Check event signature matches the one emitted
			event := s.precompile.ABI.Events[authorization.EventTypeApproval]
			s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
			s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight())) //nolint:gosec // G115

			// Check the fully unpacked event matches the one emitted
			var approvalEvent erc20precompile.EventApproval
			err = cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeApproval, *log)
			s.Require().NoError(err, "unable to unpack log into approval event")

			s.Require().Equal(tc.owner, approvalEvent.Owner, "expected different owner address")
			s.Require().Equal(tc.spender, approvalEvent.Spender, "expected different spender address")
			s.Require().Equal(tc.amount, approvalEvent.Value, "expected different amount")
		})
	}
}
