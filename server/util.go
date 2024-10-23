// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package server

import (
	"net"
	"net/http"
	"time"

	// TODO update import to local pkg when rpc pkg is migrated
	"github.com/evmos/evmos/v20/server/config"
	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/cobra"
	"golang.org/x/net/netutil"

	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"

	"cosmossdk.io/log"
	tmcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	rpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
)

// AddCommands adds server commands
func AddCommands(
	rootCmd *cobra.Command,
	opts StartOptions,
	appExport types.AppExporter,
	addStartFlags types.ModuleInitFlags,
) {
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		sdkserver.ShowNodeIDCmd(),
		sdkserver.ShowValidatorCmd(),
		sdkserver.ShowAddressCmd(),
		sdkserver.VersionCmd(),
		tmcmd.ResetAllCmd,
		tmcmd.ResetStateCmd,
		sdkserver.BootstrapStateCmd(opts.AppCreator),
	)

	startCmd := StartCmd(opts)
	addStartFlags(startCmd)

	rootCmd.AddCommand(
		startCmd,
		tendermintCmd,
		sdkserver.ExportCmd(appExport, opts.DefaultNodeHome),
		version.NewVersionCommand(),
		sdkserver.NewRollbackCmd(opts.AppCreator, opts.DefaultNodeHome),

		// custom tx indexer command
		NewIndexTxCmd(),
	)
}

// ConnectTmWS connects to a Tendermint WebSocket (WS) server.
// Parameters:
// - tmRPCAddr: The RPC address of the Tendermint server.
// - tmEndpoint: The WebSocket endpoint on the Tendermint server.
// - logger: A logger instance used to log debug and error messages.
func ConnectTmWS(tmRPCAddr, tmEndpoint string, logger log.Logger) *rpcclient.WSClient {
	tmWsClient, err := rpcclient.NewWS(tmRPCAddr, tmEndpoint,
		rpcclient.MaxReconnectAttempts(256),
		rpcclient.ReadWait(120*time.Second),
		rpcclient.WriteWait(120*time.Second),
		rpcclient.PingPeriod(50*time.Second),
		rpcclient.OnReconnect(func() {
			logger.Debug("EVM RPC reconnects to Tendermint WS", "address", tmRPCAddr+tmEndpoint)
		}),
	)

	if err != nil {
		logger.Error(
			"Tendermint WS client could not be created",
			"address", tmRPCAddr+tmEndpoint,
			"error", err,
		)
	} else if err := tmWsClient.OnStart(); err != nil {
		logger.Error(
			"Tendermint WS client could not start",
			"address", tmRPCAddr+tmEndpoint,
			"error", err,
		)
	}

	return tmWsClient
}

// MountGRPCWebServices mounts gRPC-Web services on specific HTTP POST routes.
// Parameters:
// - router: The HTTP router instance to mount the routes on (using mux.Router).
// - grpcWeb: The wrapped gRPC-Web server that will handle incoming gRPC-Web and WebSocket requests.
// - grpcResources: A list of resource endpoints (URLs) that should be mounted for gRPC-Web POST requests.
// - logger: A logger instance used to log information about the mounted resources.
func MountGRPCWebServices(
	router *mux.Router,
	grpcWeb *grpcweb.WrappedGrpcServer,
	grpcResources []string,
	logger log.Logger,
) {
	for _, res := range grpcResources {
		logger.Info("[GRPC Web] HTTP POST mounted", "resource", res)

		s := router.Methods("POST").Subrouter()
		s.HandleFunc(res, func(resp http.ResponseWriter, req *http.Request) {
			if grpcWeb.IsGrpcWebSocketRequest(req) {
				grpcWeb.HandleGrpcWebsocketRequest(resp, req)
				return
			}

			if grpcWeb.IsGrpcWebRequest(req) {
				grpcWeb.HandleGrpcWebRequest(resp, req)
				return
			}
		})
	}
}

// Listen starts a net.Listener on the tcp network on the given address.
// If there is a specified MaxOpenConnections in the config, it will also set the limitListener.
func Listen(addr string, config *config.Config) (net.Listener, error) {
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if config.JSONRPC.MaxOpenConnections > 0 {
		ln = netutil.LimitListener(ln, config.JSONRPC.MaxOpenConnections)
	}
	return ln, err
}
