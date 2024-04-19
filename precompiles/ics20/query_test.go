package ics20_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/evmos/evmos/v17/precompiles/authorization"
	cmn "github.com/evmos/evmos/v17/precompiles/common"
	"github.com/evmos/evmos/v17/precompiles/ics20"
	"github.com/evmos/evmos/v17/utils"
)

func (s *PrecompileTestSuite) TestDenomTrace() {
	var expTrace types.DenomTrace
	method := s.precompile.Methods[ics20.DenomTraceMethod]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty args",
			func() []interface{} { return []interface{}{} },
			func([]byte, []interface{}) {},
			200000,
			true,
			"invalid input arguments",
		},
		{
			"fail - invalid denom trace",
			func() []interface{} {
				return []interface{}{"invalid denom trace"}
			},
			func([]byte, []interface{}) {},
			200000,
			true,
			"invalid denom trace",
		},
		{
			"success - denom trace not found, return empty struct",
			func() []interface{} {
				expTrace.Path = "transfer/channelToA/transfer/channelToB"
				expTrace.BaseDenom = utils.BaseDenom
				return []interface{}{
					expTrace.IBCDenom(),
				}
			},
			func(data []byte, _ []interface{}) {
				var out ics20.DenomTraceResponse
				err := s.precompile.UnpackIntoInterface(&out, ics20.DenomTraceMethod, data)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal("", out.DenomTrace.BaseDenom)
				s.Require().Equal("", out.DenomTrace.Path)
			},
			200000,
			false,
			"",
		},
		{
			"success - denom trace",
			func() []interface{} {
				expTrace.Path = "transfer/channelToA/transfer/channelToB"
				expTrace.BaseDenom = utils.BaseDenom
				s.app.TransferKeeper.SetDenomTrace(s.ctx, expTrace)
				return []interface{}{
					expTrace.IBCDenom(),
				}
			},
			func(data []byte, _ []interface{}) {
				var out ics20.DenomTraceResponse
				err := s.precompile.UnpackIntoInterface(&out, ics20.DenomTraceMethod, data)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(expTrace.Path, out.DenomTrace.Path)
				s.Require().Equal(expTrace.BaseDenom, out.DenomTrace.BaseDenom)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			contract := s.NewPrecompileContract(tc.gas)
			args := tc.malleate()
			bz, err := s.precompile.DenomTrace(s.ctx, contract, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDenomTraces() {
	expTraces := types.Traces(nil)
	method := s.precompile.Methods[ics20.DenomTracesMethod]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty args",
			func() []interface{} { return []interface{}{} },
			func([]byte, []interface{}) {},
			200000,
			true,
			"invalid number of arguments",
		},
		{
			"success - gets denom traces",
			func() []interface{} {
				expTraces = append(expTraces, types.DenomTrace{Path: "", BaseDenom: utils.BaseDenom})
				expTraces = append(expTraces, types.DenomTrace{Path: "transfer/channelToA/transfer/channelToB", BaseDenom: utils.BaseDenom})
				expTraces = append(expTraces, types.DenomTrace{Path: "transfer/channelToB", BaseDenom: utils.BaseDenom})

				for _, trace := range expTraces {
					s.app.TransferKeeper.SetDenomTrace(s.ctx, trace)
				}
				return []interface{}{
					query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			func(data []byte, _ []interface{}) {
				var denomTraces ics20.DenomTracesResponse
				err := s.precompile.UnpackIntoInterface(&denomTraces, ics20.DenomTracesMethod, data)
				s.Require().Equal(denomTraces.PageResponse.Total, uint64(3))
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(3, len(denomTraces.DenomTraces))
				for i, trace := range denomTraces.DenomTraces {
					s.Require().Equal(expTraces[i].Path, trace.Path)
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
			contract := s.NewPrecompileContract(tc.gas)
			args := tc.malleate()
			bz, err := s.precompile.DenomTraces(s.ctx, contract, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDenomHash() {
	reqTrace := types.DenomTrace{
		Path:      "transfer/channelToA/transfer/channelToB",
		BaseDenom: utils.BaseDenom,
	}
	method := s.precompile.Methods[ics20.DenomHashMethod]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success - trace not found, returns empty string",
			func() []interface{} { return []interface{}{"transfer/channelToB/transfer/channelToA"} },
			func(data []byte, _ []interface{}) {
				var hash string
				err := s.precompile.UnpackIntoInterface(&hash, ics20.DenomHashMethod, data)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal("", hash)
			},
			200000,
			false,
			"",
		},
		{
			"success - get the hash of a denom trace",
			func() []interface{} {
				s.app.TransferKeeper.SetDenomTrace(s.ctx, reqTrace)
				return []interface{}{
					reqTrace.GetFullDenomPath(),
				}
			},
			func(data []byte, _ []interface{}) {
				var hash string
				err := s.precompile.UnpackIntoInterface(&hash, ics20.DenomHashMethod, data)
				s.Require().NoError(err, "failed to unpack output", err)
				s.Require().Equal(reqTrace.Hash().String(), hash)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			contract := s.NewPrecompileContract(tc.gas)
			args := tc.malleate()

			bz, err := s.precompile.DenomHash(s.ctx, contract, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				tc.postCheck(bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestAllowance() {
	var (
		path   = NewTransferPath(s.chainA, s.chainB)
		path2  = NewTransferPath(s.chainA, s.chainB)
		paths  = []*ibctesting.Path{path, path2}
		method = s.precompile.Methods[authorization.AllowanceMethod]
	)
	// set channel, otherwise is "" and throws error
	path.EndpointA.ChannelID = "channel-0"
	path2.EndpointA.ChannelID = "channel-1"

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(bz []byte)
		gas         uint64
		expErr      bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			func([]byte) {},
			100000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 1),
		},
		{
			"success - no allowance == empty array",
			func() []interface{} {
				return []interface{}{
					s.address,
					s.differentAddr,
				}
			},
			func(bz []byte) {
				var allocations []cmn.ICS20Allocation
				err := s.precompile.UnpackIntoInterface(&allocations, authorization.AllowanceMethod, bz)
				s.Require().NoError(err, "failed to unpack output")
				s.Require().Len(allocations, 0)
			},
			100000,
			false,
			"",
		},
		{
			"success - auth with one allocation",
			func() []interface{} {
				err := s.NewTransferAuthorization(
					s.ctx,
					s.app,
					s.differentAddr,
					s.address,
					path,
					defaultCoins,
					[]string{s.chainB.SenderAccount.GetAddress().String()},
				)
				s.Require().NoError(err)

				return []interface{}{
					s.differentAddr,
					s.address,
				}
			},
			func(bz []byte) {
				expAllocs := []cmn.ICS20Allocation{
					{
						SourcePort:    path.EndpointA.ChannelConfig.PortID,
						SourceChannel: path.EndpointA.ChannelID,
						SpendLimit:    defaultCmnCoins,
						AllowList:     []string{s.chainB.SenderAccount.GetAddress().String()},
					},
				}

				var allocations []cmn.ICS20Allocation
				err := s.precompile.UnpackIntoInterface(&allocations, authorization.AllowanceMethod, bz)
				s.Require().NoError(err, "failed to unpack output")

				s.Require().Equal(expAllocs, allocations)
			},
			100000,
			false,
			"",
		},
		{
			"success - auth with multiple allocations",
			func() []interface{} {
				allocs := make([]types.Allocation, len(paths))
				for i, p := range paths {
					allocs[i] = types.Allocation{
						SourcePort:    p.EndpointA.ChannelConfig.PortID,
						SourceChannel: p.EndpointA.ChannelID,
						SpendLimit:    mutliSpendLimit,
						AllowList:     []string{s.chainB.SenderAccount.GetAddress().String()},
					}
				}

				err := s.NewTransferAuthorizationWithAllocations(
					s.ctx,
					s.app,
					s.differentAddr,
					s.address,
					allocs,
				)
				s.Require().NoError(err)

				return []interface{}{
					s.differentAddr,
					s.address,
				}
			},
			func(bz []byte) {
				expAllocs := make([]cmn.ICS20Allocation, len(paths))
				for i, p := range paths {
					expAllocs[i] = cmn.ICS20Allocation{
						SourcePort:    p.EndpointA.ChannelConfig.PortID,
						SourceChannel: p.EndpointA.ChannelID,
						SpendLimit:    mutliCmnCoins,
						AllowList:     []string{s.chainB.SenderAccount.GetAddress().String()},
					}
				}

				var allocations []cmn.ICS20Allocation
				err := s.precompile.UnpackIntoInterface(&allocations, authorization.AllowanceMethod, bz)
				s.Require().NoError(err, "failed to unpack output")

				s.Require().Equal(expAllocs, allocations)
			},
			100000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			args := tc.malleate()
			bz, err := s.precompile.Allowance(s.ctx, &method, args)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(bz)
				tc.postCheck(bz)
			}
		})
	}
}
