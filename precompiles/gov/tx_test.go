package gov_test

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

func (s *PrecompileTestSuite) TestVote() {
	var ctx sdk.Context
	method := s.precompile.Methods[gov.VoteMethod]
	newVoterAddr := utiltx.GenerateAddress()
	const proposalID uint64 = 1
	const option uint8 = 1
	const metadata = "metadata"

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 4, 0),
		},
		{
			"fail - invalid voter address",
			func() []interface{} {
				return []interface{}{
					"",
					proposalID,
					option,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid voter address",
		},
		{
			"fail - invalid voter address",
			func() []interface{} {
				return []interface{}{
					common.Address{},
					proposalID,
					option,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid voter address",
		},
		{
			"fail - using a different voter address",
			func() []interface{} {
				return []interface{}{
					newVoterAddr,
					proposalID,
					option,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"does not match the voter address",
		},
		{
			"fail - invalid vote option",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					option + 10,
					metadata,
				}
			},
			func() {},
			200000,
			true,
			"invalid vote option",
		},
		{
			"success - vote proposal success",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
					proposalID,
					option,
					metadata,
				}
			},
			func() {
				proposal, _ := s.network.App.GovKeeper.Proposals.Get(ctx, proposalID)
				_, _, tallyResult, err := s.network.App.GovKeeper.Tally(ctx, proposal)
				s.Require().NoError(err)
				s.Require().Equal(math.NewInt(3e18).String(), tallyResult.YesCount)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, s.keyring.GetAddr(0), s.precompile, tc.gas)

			_, err := s.precompile.Vote(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
