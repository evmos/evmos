package erc20_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	erc20precompile "github.com/evmos/evmos/v15/precompiles/erc20"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
)

func (s *PrecompileTestSuite) TestEmitTransferEvent() {
	s.SetupTest()

	from := utiltx.GenerateAddress()
	to := utiltx.GenerateAddress()
	amount := big.NewInt(100)

	err := s.precompile.EmitTransferEvent(
		s.network.GetContext(), s.stateDB, from, to, amount,
	)
	s.Require().NoError(err, "expected transfer event to be emitted successfully")

	log := s.stateDB.Logs()[0]
	s.Require().Equal(log.Address, s.precompile.Address())

	// Check event signature matches the one emitted
	event := s.precompile.ABI.Events[erc20precompile.EventTypeTransfer]
	s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
	s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

	// Check the fully unpacked event matches the one emitted
	var transferEvent erc20precompile.EventTransfer
	err = cmn.UnpackLog(s.precompile.ABI, &transferEvent, erc20precompile.EventTypeTransfer, *log)
	s.Require().NoError(err, "unable to unpack log into transfer event")

	s.Require().Equal(from, transferEvent.From, "expected different from address")
	s.Require().Equal(to, transferEvent.To, "expected different to address")
	s.Require().Equal(amount, transferEvent.Value, "expected different amount")
}
