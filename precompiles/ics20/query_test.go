package ics20_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/core/vm"
	evmosibc "github.com/evmos/evmos/v18/ibc/testing"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/ics20"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/testutil/integration/ibc/coordinator"
	"github.com/evmos/evmos/v18/utils"
)

func (s *PrecompileTestSuite) TestDenomTrace() {
	var (
		ctx      sdk.Context
		nw       *network.UnitTestNetwork
		expTrace types.DenomTrace
	)
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
				nw.App.TransferKeeper.SetDenomTrace(ctx, expTrace)
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
			nw = s.network
			ctx = nw.GetContext()
			method := s.precompile.Methods[ics20.DenomTraceMethod]

			var contract *vm.Contract
			ctx, contract = s.NewPrecompileContract(tc.gas)
			args := tc.malleate()
			bz, err := s.precompile.DenomTrace(ctx, contract, &method, args)

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
	var (
		ctx       sdk.Context
		nw        *network.UnitTestNetwork
		expTraces = types.Traces(nil)
	)
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
					nw.App.TransferKeeper.SetDenomTrace(ctx, trace)
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
			nw = s.network
			ctx = nw.GetContext()

			method := s.precompile.Methods[ics20.DenomTracesMethod]

			var contract *vm.Contract
			ctx, contract = s.NewPrecompileContract(tc.gas)
			args := tc.malleate()
			bz, err := s.precompile.DenomTraces(ctx, contract, &method, args)

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
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	reqTrace := types.DenomTrace{
		Path:      "transfer/channelToA/transfer/channelToB",
		BaseDenom: utils.BaseDenom,
	}
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
				nw.App.TransferKeeper.SetDenomTrace(ctx, reqTrace)
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
			nw = s.network
			ctx = nw.GetContext()

			method := s.precompile.Methods[ics20.DenomHashMethod]

			var contract *vm.Contract
			ctx, contract = s.NewPrecompileContract(tc.gas)
			args := tc.malleate()

			bz, err := s.precompile.DenomHash(ctx, contract, &method, args)

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
		ctx    sdk.Context
		nw     *network.UnitTestNetwork
		coord  *coordinator.IntegrationCoordinator
		keys keyring.Keyring
		path   evmosibc.Path
		path2  evmosibc.Path
		paths  []evmosibc.Path
		chainB string
	)

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
					keys.GetAddr(0),
					keys.GetAddr(1),
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
					ctx,
					nw.App,
					keys.GetAddr(1),
					keys.GetAddr(0),
					&path,
					defaultCoins,
					[]string{coord.GetChainSenderAcc(chainB).GetAddress().String()},
				)
				s.Require().NoError(err)

				return []interface{}{
					keys.GetAddr(1),
					keys.GetAddr(0),
				}
			},
			func(bz []byte) {
				expAllocs := []cmn.ICS20Allocation{
					{
						SourcePort:    path.EndpointA.ChannelConfig.PortID,
						SourceChannel: path.EndpointA.ChannelID,
						SpendLimit:    defaultCmnCoins,
						AllowList:     []string{coord.GetChainSenderAcc(chainB).GetAddress().String()},
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
						AllowList:     []string{coord.GetChainSenderAcc(chainB).GetAddress().String()},
					}
				}

				err := s.NewTransferAuthorizationWithAllocations(
					ctx,
					nw.App,
					keys.GetAddr(1),
					keys.GetAddr(0),
					allocs,
				)
				s.Require().NoError(err)

				return []interface{}{
					keys.GetAddr(1),
					keys.GetAddr(0),
				}
			},
			func(bz []byte) {
				expAllocs := make([]cmn.ICS20Allocation, len(paths))
				for i, p := range paths {
					expAllocs[i] = cmn.ICS20Allocation{
						SourcePort:    p.EndpointA.ChannelConfig.PortID,
						SourceChannel: p.EndpointA.ChannelID,
						SpendLimit:    mutliCmnCoins,
						AllowList:     []string{coord.GetChainSenderAcc(chainB).GetAddress().String()},
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
			nw = s.network
			ctx = nw.GetContext()
			coord = s.coordinator
			chainB = s.chainB

			// set channel, otherwise is "" and throws error
			path = *s.transferPath
			path2 = *s.transferPath
			path.EndpointA.ChannelID = "channel-0"
			path2.EndpointA.ChannelID = "channel-1"
			paths = []evmosibc.Path{path, path2}

			method := s.precompile.Methods[authorization.AllowanceMethod]

			args := tc.malleate()
			bz, err := s.precompile.Allowance(nw.GetContext(), &method, args)

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
