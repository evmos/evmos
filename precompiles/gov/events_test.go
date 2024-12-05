package gov_test

import (
	"math/big"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v20/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	"github.com/evmos/evmos/v20/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestVoteEvent() {
	var (
		stDB   *statedb.StateDB
		ctx    sdk.Context
		method = s.precompile.Methods[gov.VoteMethod]
	)

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
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[gov.EventTypeVote]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var voteEvent gov.EventVote
				err := cmn.UnpackLog(s.precompile.ABI, &voteEvent, gov.EventTypeVote, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), voteEvent.Voter)
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
		stDB = s.network.GetStateDB()
		ctx = s.network.GetContext()

		contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
		ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		initialGas := ctx.GasMeter().GasConsumed()
		s.Require().Zero(initialGas)

		_, err := s.precompile.Vote(ctx, s.keyring.GetAddr(0), contract, stDB, &method, tc.malleate(s.keyring.GetAddr(0), 1, 1, "metadata"))

		if tc.expError {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		} else {
			s.Require().NoError(err)
			tc.postCheck()
		}
	}
}

func (s *PrecompileTestSuite) TestVoteWeightedEvent() {
	var (
		stDB   *statedb.StateDB
		ctx    sdk.Context
		method = s.precompile.Methods[gov.VoteWeightedMethod]
	)

	testCases := []struct {
		name        string
		malleate    func(voter common.Address, proposalId uint64, options gov.WeightedVoteOptions) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct VoteWeighted event is emitted",
			func(voter common.Address, proposalId uint64, options gov.WeightedVoteOptions) []interface{} {
				return []interface{}{
					voter,
					proposalId,
					options,
					"",
				}
			},
			func() {
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[gov.EventTypeVoteWeighted]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var voteWeightedEvent gov.EventVoteWeighted
				err := cmn.UnpackLog(s.precompile.ABI, &voteWeightedEvent, gov.EventTypeVoteWeighted, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), voteWeightedEvent.Voter)
				s.Require().Equal(uint64(1), voteWeightedEvent.ProposalId)
				s.Require().Equal(2, len(voteWeightedEvent.Options))
				s.Require().Equal(uint8(1), voteWeightedEvent.Options[0].Option)
				s.Require().Equal("0.70", voteWeightedEvent.Options[0].Weight)
				s.Require().Equal(uint8(2), voteWeightedEvent.Options[1].Option)
				s.Require().Equal("0.30", voteWeightedEvent.Options[1].Weight)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stDB = s.network.GetStateDB()
			ctx = s.network.GetContext()

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
			ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
			initialGas := ctx.GasMeter().GasConsumed()
			s.Require().Zero(initialGas)

			options := gov.WeightedVoteOptions{
				gov.WeightedVoteOption{Option: 1, Weight: "0.70"},
				gov.WeightedVoteOption{Option: 2, Weight: "0.30"},
			}

			_, err := s.precompile.VoteWeighted(ctx, s.keyring.GetAddr(0), contract, stDB, &method, tc.malleate(s.keyring.GetAddr(0), 1, options))

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDepositEvent() {
	var (
		stDB   *statedb.StateDB
		ctx    sdk.Context
		method = s.precompile.Methods[gov.DepositMethod]
	)

	testCases := []struct {
		name        string
		malleate    func(depositor common.Address, proposalId uint64, amount sdk.Coins) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct Deposit event is emitted",
			func(depositor common.Address, proposalId uint64, amount sdk.Coins) []interface{} {
				return []interface{}{
					depositor,
					proposalId,
					cmn.NewCoinsResponse(amount),
				}
			},
			func() {
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[gov.EventTypeDeposit]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(uint64(ctx.BlockHeight()), log.BlockNumber) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var depositEvent gov.EventDeposit
				err := cmn.UnpackLog(s.precompile.ABI, &depositEvent, gov.EventTypeDeposit, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), depositEvent.Depositor)
				s.Require().Equal(uint64(1), depositEvent.ProposalId)
				s.Require().Equal(big.NewInt(1e18), depositEvent.Amount[0].Amount)
				s.Require().Equal(types.GetEVMCoinDenom(), depositEvent.Amount[0].Denom)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stDB = s.network.GetStateDB()
			ctx = s.network.GetContext()

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
			ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
			initialGas := ctx.GasMeter().GasConsumed()
			s.Require().Zero(initialGas)

			amount := sdk.NewCoins(sdk.NewCoin(types.GetEVMCoinDenom(), math.NewIntFromBigInt(big.NewInt(1e18))))

			_, err := s.precompile.Deposit(ctx, s.keyring.GetAddr(0), contract, stDB, &method, tc.malleate(s.keyring.GetAddr(0), 1, amount))

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCancelProposalEvent() {
	var (
		stDB   *statedb.StateDB
		ctx    sdk.Context
		method = s.precompile.Methods[gov.CancelProposalMethod]
	)

	testCases := []struct {
		name        string
		malleate    func(proposer common.Address, proposalId uint64) []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - the correct CancelProposal event is emitted",
			func(proposer common.Address, proposalId uint64) []interface{} {
				return []interface{}{
					proposer,
					proposalId,
				}
			},
			func() {
				log := stDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[gov.EventTypeCancelProposal]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Check the fully unpacked event matches the one emitted
				var cancelEvent gov.EventCancelProposal
				err := cmn.UnpackLog(s.precompile.ABI, &cancelEvent, gov.EventTypeCancelProposal, *log)
				s.Require().NoError(err)
				s.Require().Equal(s.keyring.GetAddr(0), cancelEvent.Proposer)
				s.Require().Equal(uint64(1), cancelEvent.ProposalId)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stDB = s.network.GetStateDB()
			ctx = s.network.GetContext()

			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
			ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
			initialGas := ctx.GasMeter().GasConsumed()
			s.Require().Zero(initialGas)

			_, err := s.precompile.CancelProposal(ctx, s.keyring.GetAddr(0), contract, stDB, &method, tc.malleate(s.keyring.GetAddr(0), 1))

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
