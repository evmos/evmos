// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	types "github.com/evmos/evmos/v20/x/erc20/types"
)

// MsgClient is an autogenerated mock type for the MsgClient type
type MsgClient struct {
	mock.Mock
}

// ConvertERC20 provides a mock function with given fields: ctx, in, opts
func (_m *MsgClient) ConvertERC20(ctx context.Context, in *types.MsgConvertERC20, opts ...grpc.CallOption) (*types.MsgConvertERC20Response, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ConvertERC20")
	}

	var r0 *types.MsgConvertERC20Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.MsgConvertERC20, ...grpc.CallOption) (*types.MsgConvertERC20Response, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.MsgConvertERC20, ...grpc.CallOption) *types.MsgConvertERC20Response); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgConvertERC20Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.MsgConvertERC20, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateParams provides a mock function with given fields: ctx, in, opts
func (_m *MsgClient) UpdateParams(ctx context.Context, in *types.MsgUpdateParams, opts ...grpc.CallOption) (*types.MsgUpdateParamsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateParams")
	}

	var r0 *types.MsgUpdateParamsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.MsgUpdateParams, ...grpc.CallOption) (*types.MsgUpdateParamsResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.MsgUpdateParams, ...grpc.CallOption) *types.MsgUpdateParamsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgUpdateParamsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.MsgUpdateParams, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMsgClient creates a new instance of MsgClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMsgClient(t interface {
	mock.TestingT
	Cleanup(func())
},
) *MsgClient {
	mock := &MsgClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
