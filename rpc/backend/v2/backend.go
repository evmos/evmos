// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	"context"
	"math/big"

	"cosmossdk.io/log"
	cmtrpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	rpctypes "github.com/evmos/evmos/v20/rpc/types"
	"github.com/evmos/evmos/v20/server/config"
	evmostypes "github.com/evmos/evmos/v20/types"
)

var _ BackendI = (*Backend)(nil)

// Backend implements the BackendI interface
type Backend struct {
	ctx                 context.Context
	clientCtx           client.Context
	queryClient         *rpctypes.QueryClient // gRPC query client
	rpcClient           cmtrpcclient.Client
	logger              log.Logger
	chainID             *big.Int
	cfg                 config.Config
	allowUnprotectedTxs bool
}

// NewBackend creates a new Backend instance for cosmos and ethereum namespaces
func NewBackend(
	ctx *server.Context,
	logger log.Logger,
	clientCtx client.Context,
	allowUnprotectedTxs bool,
) *Backend {
	chainID, err := evmostypes.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	appConf, err := config.GetConfig(ctx.Viper)
	if err != nil {
		panic(err)
	}

	rpcClient := clientCtx.Client.(cmtrpcclient.Client)

	return &Backend{
		ctx:                 context.Background(),
		clientCtx:           clientCtx,
		rpcClient:           rpcClient,
		queryClient:         rpctypes.NewQueryClient(clientCtx),
		logger:              logger.With("module", "backend"),
		chainID:             chainID,
		cfg:                 appConf,
		allowUnprotectedTxs: allowUnprotectedTxs,
	}
}
