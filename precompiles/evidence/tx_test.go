package evidence_test

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/evidence"
	"github.com/evmos/evmos/v20/precompiles/testutil"
)

func (s *PrecompileTestSuite) TestSubmitEvidence() {
	method := s.precompile.Methods[evidence.SubmitEvidenceMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"success - submit equivocation evidence",
			func() []interface{} {
				validators, err := s.network.App.StakingKeeper.GetAllValidators(s.network.GetContext())
				s.Require().NoError(err)
				s.Require().NotEmpty(validators)

				validator := validators[0]

				valConsAddr, err := validator.GetConsAddr()
				s.Require().NoError(err)

				evidenceData := evidence.EquivocationData{
					Height:           1,
					Time:             uint64(time.Now().UTC().Unix()), //nolint:gosec // G115
					Power:            1000,
					ConsensusAddress: types.ConsAddress(valConsAddr).String(),
				}
				return []interface{}{
					evidenceData,
				}
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			bytes, err := s.precompile.SubmitEvidence(ctx, s.keyring.GetAddr(0), contract, s.network.GetStateDB(), &method, tc.malleate())

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(cmn.TrueValue, bytes)
			}
		})
	}
}
