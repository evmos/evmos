package gov_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/gov"
)

func (s *PrecompileTestSuite) TestVoteEvent() {
	method := s.precompile.Methods[gov.VoteMethod]
	testCases := []struct {
		name        string
		malleate    func(voter common.Address, proposalId uint64, option uint8, metadata string) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct event is emitted",
			func(voter common.Address, proposalId uint64, option uint8, metadata string) []interface{} {
				return []interface{}{
					voter,
					proposalId,
					option,
					metadata,
				}
			},
			func() {
				log := s.stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[gov.EventTypeVote]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var voteEvent gov.EventVote
				err := cmn.UnpackLog(s.precompile.ABI, &voteEvent, gov.EventTypeVote, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.address, voteEvent.Voter)
				s.Require().Equal(uint64(1), voteEvent.ProposalId)
				s.Require().Equal(uint8(1), voteEvent.Option)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.SetupTest()

		contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)
		s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		initialGas := s.ctx.GasMeter().GasConsumed()
		s.Require().Zero(initialGas)

		_, err := s.precompile.Vote(s.ctx, s.address, contract, s.stateDB, &method, tc.malleate(s.address, 1, 1, "metadata"))

		if tc.expError {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		} else {
			s.Require().NoError(err)
			tc.postCheck()
		}
	}
}
