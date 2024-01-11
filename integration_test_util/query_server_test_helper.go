package integration_test_util

import (
	gocontext "context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"

	abci "github.com/cometbft/cometbft/abci/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryServiceTestHelper provides a helper for making grpc query service
// rpc calls in unit tests. It implements both the grpc Server and ClientConn
// interfaces needed to register a query service server and create a query
// service client.
type QueryServiceTestHelper struct {
	*baseapp.GRPCQueryRouter
	Ctx sdk.Context
	cdc encoding.Codec
}

var (
	_ gogogrpc.Server     = &QueryServiceTestHelper{}
	_ gogogrpc.ClientConn = &QueryServiceTestHelper{}
)

// NewQueryServerTestHelper creates a new QueryServiceTestHelper that wraps
// the provided sdk.Context.
//
// This one is copied from baseapp of cosmos-sdk to add ability to include x-cosmos-block-height header
func NewQueryServerTestHelper(ctx sdk.Context, interfaceRegistry types.InterfaceRegistry) *QueryServiceTestHelper {
	qrt := baseapp.NewGRPCQueryRouter()
	qrt.SetInterfaceRegistry(interfaceRegistry)
	return &QueryServiceTestHelper{GRPCQueryRouter: qrt, Ctx: ctx, cdc: codec.NewProtoCodec(interfaceRegistry).GRPCCodec()}
}

// Invoke implements the grpc ClientConn.Invoke method
func (q *QueryServiceTestHelper) Invoke(_ gocontext.Context, method string, args, reply interface{}, callOptions ...grpc.CallOption) error {
	querier := q.Route(method)
	if querier == nil {
		return fmt.Errorf("handler not found for %s", method)
	}
	reqBz, err := q.cdc.Marshal(args)
	if err != nil {
		return err
	}

	for _, option := range callOptions {
		if option == nil {
			continue
		}

		if header, ok := option.(grpc.HeaderCallOption); ok {
			if header.HeaderAddr != nil {
				var mdI = metadata.New(map[string]string{
					grpctypes.GRPCBlockHeightHeader: fmt.Sprintf("%d", q.Ctx.BlockHeight()),
				})
				*header.HeaderAddr = mdI
			}
			break
		}
	}

	res, err := querier(q.Ctx, abci.RequestQuery{Data: reqBz})
	if err != nil {
		return err
	}

	err = q.cdc.Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (q *QueryServiceTestHelper) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}
