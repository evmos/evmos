package gov_test

import (
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

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
				{Option: 1, Weight: "0.70"},
				{Option: 2, Weight: "0.30"},
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

func (s *PrecompileTestSuite) TestEmitUpdateParamsEvent() {
	var (
		stateDB *statedb.StateDB
		ctx     sdk.Context
		method  = s.precompile.Methods[gov.UpdateParamsMethod]
	)
	params := v1.DefaultParams()
	votingPeriod := time.Hour * 24 * 7
	adjustedParams := v1.Params{
		MinDeposit:                 params.MinDeposit,
		MaxDepositPeriod:           params.MaxDepositPeriod,
		VotingPeriod:               &votingPeriod,
		Quorum:                     params.Quorum,
		Threshold:                  params.Threshold,
		VetoThreshold:              params.VetoThreshold,
		MinInitialDepositRatio:     params.MinInitialDepositRatio,
		ProposalCancelRatio:        params.ProposalCancelRatio,
		ProposalCancelDest:         params.ProposalCancelDest,
		ExpeditedVotingPeriod:      params.ExpeditedVotingPeriod,
		ExpeditedThreshold:         params.ExpeditedThreshold,
		ExpeditedMinDeposit:        params.ExpeditedMinDeposit,
		BurnVoteQuorum:             params.BurnVoteQuorum,
		BurnProposalDepositPrevote: params.BurnProposalDepositPrevote,
		BurnVoteVeto:               params.BurnVoteVeto,
		MinDepositRatio:            params.MinDepositRatio,
	}

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expErr      bool
		errContains string
		gas         uint64
		postCheck   func()
	}{
		{
			"success - emit update params event",
			func() []interface{} {
				return []interface{}{
					adjustedParams,
				}
			},
			false,
			"",
			20000,
			func() {
				log := stateDB.Logs()[0]
				s.Require().Equal(log.Address, s.precompile.Address())

				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[gov.EventTypeUpdateParams]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight())) //nolint:gosec // G115

				// Create a struct that matches the event structure
				var eventData struct {
					Params gov.ParamsOutput `json:"params"`
				}
				err := cmn.UnpackLog(s.precompile.ABI, &eventData, gov.EventTypeUpdateParams, *log)
				s.Require().NoError(err)

				// Now check the values from eventData.Params
				s.Require().Equal(adjustedParams.VotingPeriod.Nanoseconds(), eventData.Params.VotingPeriod)
				s.Require().Equal(adjustedParams.MaxDepositPeriod.Nanoseconds(), eventData.Params.MaxDepositPeriod)
				s.Require().Equal(adjustedParams.ExpeditedVotingPeriod.Nanoseconds(), eventData.Params.ExpeditedVotingPeriod)
				s.Require().Equal(cmn.NewCoinsResponse(adjustedParams.MinDeposit), eventData.Params.MinDeposit)
				s.Require().Equal(adjustedParams.MaxDepositPeriod.Nanoseconds(), eventData.Params.MaxDepositPeriod)
				s.Require().Equal(adjustedParams.Quorum, eventData.Params.Quorum)
				s.Require().Equal(adjustedParams.Threshold, eventData.Params.Threshold)
				s.Require().Equal(adjustedParams.VetoThreshold, eventData.Params.VetoThreshold)
				s.Require().Equal(adjustedParams.MinInitialDepositRatio, eventData.Params.MinInitialDepositRatio)
				s.Require().Equal(adjustedParams.ProposalCancelRatio, eventData.Params.ProposalCancelRatio)
				s.Require().Equal(adjustedParams.ProposalCancelDest, eventData.Params.ProposalCancelDest)
				s.Require().Equal(adjustedParams.ExpeditedThreshold, eventData.Params.ExpeditedThreshold)
				s.Require().Equal(cmn.NewCoinsResponse(adjustedParams.ExpeditedMinDeposit), eventData.Params.ExpeditedMinDeposit)
				s.Require().Equal(adjustedParams.BurnVoteQuorum, eventData.Params.BurnVoteQuorum)
				s.Require().Equal(adjustedParams.BurnProposalDepositPrevote, eventData.Params.BurnProposalDepositPrevote)
				s.Require().Equal(adjustedParams.BurnVoteVeto, eventData.Params.BurnVoteVeto)
				s.Require().Equal(adjustedParams.MinDepositRatio, eventData.Params.MinDepositRatio)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			stateDB = s.network.GetStateDB()
			ctx = s.network.GetContext()
			authority := s.network.App.GovKeeper.GetAuthority()

			bech32Prefix := strings.SplitN(authority, "1", 2)[0]

			addressBz, err := sdk.GetFromBech32(authority, bech32Prefix)
			s.Require().NoError(err)

			authorityAddress := common.BytesToAddress(addressBz)
			contract := vm.NewContract(vm.AccountRef(s.keyring.GetAddr(0)), s.precompile, big.NewInt(0), tc.gas)
			_, err = s.precompile.UpdateParams(ctx, authorityAddress, contract, stateDB, &method, tc.malleate())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				tc.postCheck()
			}
		})
	}
}
