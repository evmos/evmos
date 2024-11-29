package evidence_test

import (
	"fmt"
	"time"

	evidencetypes "cosmossdk.io/x/evidence/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/evidence"
	"github.com/evmos/evmos/v20/precompiles/testutil"
)

func (s *PrecompileTestSuite) TestEvidence() {
	method := s.precompile.Methods[evidence.EvidenceMethod]

	testCases := []struct {
		name          string
		malleate      func(hash []byte) []interface{}
		setupEvidence func() []byte
		gas           uint64
		expError      bool
		errContains   string
		postCheck     func(evidence *evidence.EquivocationData)
	}{
		{
			"fail - empty input args",
			func(_ []byte) []interface{} {
				return []interface{}{}
			},
			func() []byte {
				return []byte{}
			},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
			nil,
		},
		{
			"fail - invalid evidence hash",
			func(_ []byte) []interface{} {
				return []interface{}{
					[]byte{},
				}
			},
			func() []byte {
				return []byte{}
			},
			200000,
			true,
			evidence.ErrInvalidEvidenceHash,
			nil,
		},
		{
			"success - evidence found",
			func(hash []byte) []interface{} {
				return []interface{}{
					hash,
				}
			},
			func() []byte {
				validators, err := s.network.App.StakingKeeper.GetAllValidators(s.network.GetContext())
				s.Require().NoError(err)
				s.Require().NotEmpty(validators)

				validator := validators[0]
				valConsAddr, err := validator.GetConsAddr()
				s.Require().NoError(err)

				evidenceData := &evidencetypes.Equivocation{
					Height:           1,
					Time:             time.Unix(1234567890, 0),
					Power:            1000,
					ConsensusAddress: sdk.ConsAddress(valConsAddr).String(),
				}

				err = s.network.App.EvidenceKeeper.SubmitEvidence(s.network.GetContext(), evidenceData)
				s.Require().NoError(err)

				return evidenceData.Hash()
			},
			200000,
			false,
			"",
			func(e *evidence.EquivocationData) {
				s.Require().Equal(int64(1), e.Height)
				s.Require().Equal(uint64(1234567890), e.Time)
				s.Require().Equal(int64(1000), e.Power)
				s.Require().NotEmpty(e.ConsensusAddress)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			evidenceHash := tc.setupEvidence()

			_, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.Evidence(ctx, &method, tc.malleate(evidenceHash))

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				var out evidence.SingleEvidenceOutput
				err = s.precompile.UnpackIntoInterface(&out, evidence.EvidenceMethod, bz)
				s.Require().NoError(err)
				if tc.postCheck != nil {
					tc.postCheck(&out.Evidence)
				}
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGetAllEvidence() {
	method := s.precompile.Methods[evidence.GetAllEvidenceMethod]

	testCases := []struct {
		name          string
		malleate      func() []interface{}
		setupEvidence func()
		gas           uint64
		expError      bool
		errContains   string
		postCheck     func(evidence []evidence.EquivocationData, pageResponse *query.PageResponse)
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func() {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
			nil,
		},
		{
			"success - empty evidence list",
			func() []interface{} {
				return []interface{}{
					&query.PageRequest{
						Limit:      10,
						CountTotal: true,
					},
				}
			},
			func() {},
			200000,
			false,
			"",
			nil,
		},
		{
			"success - with evidence",
			func() []interface{} {
				return []interface{}{
					&query.PageRequest{
						Limit:      1,
						CountTotal: true,
					},
				}
			},
			func() {
				validators, err := s.network.App.StakingKeeper.GetAllValidators(s.network.GetContext())
				s.Require().NoError(err)
				s.Require().NotEmpty(validators)

				validator := validators[0]
				valConsAddr, err := validator.GetConsAddr()
				s.Require().NoError(err)

				evidenceData := &evidencetypes.Equivocation{
					Height:           1,
					Time:             time.Unix(1234567890, 0),
					Power:            1000,
					ConsensusAddress: sdk.ConsAddress(valConsAddr).String(),
				}

				err = s.network.App.EvidenceKeeper.SubmitEvidence(s.network.GetContext(), evidenceData)
				s.Require().NoError(err)
			},
			200000,
			false,
			"",
			func(evidenceList []evidence.EquivocationData, _ *query.PageResponse) {
				s.Require().Len(evidenceList, 1)
				s.Require().Equal(int64(1), evidenceList[0].Height)
				s.Require().Equal(uint64(1234567890), evidenceList[0].Time)
				s.Require().Equal(int64(1000), evidenceList[0].Power)
				s.Require().NotEmpty(evidenceList[0].ConsensusAddress)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			tc.setupEvidence()

			_, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetAddr(0), s.precompile, tc.gas)

			bz, err := s.precompile.GetAllEvidence(ctx, &method, tc.malleate())

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				var out evidence.AllEvidenceOutput
				err = s.precompile.UnpackIntoInterface(&out, evidence.GetAllEvidenceMethod, bz)
				s.Require().NoError(err)
				if tc.postCheck != nil {
					tc.postCheck(out.Evidence, &out.PageResponse)
				}
			}
		})
	}
}
