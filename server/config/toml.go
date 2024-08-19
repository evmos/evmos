// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package config

// DefaultEVMConfigTemplate defines the configuration template for the EVM RPC configuration.
const DefaultEVMConfigTemplate = `
###############################################################################
###                             EVM Configuration                           ###
###############################################################################

[evm]

# Tracer defines the 'vm.Tracer' type that the EVM will use when the node is run in
# debug mode. To enable tracing use the '--evm.tracer' flag when starting your node.
# Valid types are: json|struct|access_list|markdown
tracer = "{{ .EVM.Tracer }}"

# MaxTxGasWanted defines the gas wanted for each eth tx returned in ante handler in check tx mode.
max-tx-gas-wanted = {{ .EVM.MaxTxGasWanted }}

###############################################################################
###                           JSON RPC Configuration                        ###
###############################################################################

[json-rpc]

# Enable defines if the JSONRPC server should be enabled.
enable = {{ .JSONRPC.Enable }}

# Address defines the EVM RPC HTTP server address to bind to.
address = "{{ .JSONRPC.Address }}"

# Address defines the EVM WebSocket server address to bind to.
ws-address = "{{ .JSONRPC.WsAddress }}"

# API defines a list of JSON-RPC namespaces that should be enabled
# Example: "eth,txpool,personal,net,debug,web3"
api = "{{range $index, $elmt := .JSONRPC.API}}{{if $index}},{{$elmt}}{{else}}{{$elmt}}{{end}}{{end}}"

# GasCap sets a cap on gas that can be used in eth_call/estimateGas (0=infinite). Default: 25,000,000.
gas-cap = {{ .JSONRPC.GasCap }}

# Allow insecure account unlocking when account-related RPCs are exposed by http
allow-insecure-unlock = {{ .JSONRPC.AllowInsecureUnlock }}

# EVMTimeout is the global timeout for eth_call. Default: 5s.
evm-timeout = "{{ .JSONRPC.EVMTimeout }}"

# TxFeeCap is the global tx-fee cap for send transaction. Default: 1eth.
txfee-cap = {{ .JSONRPC.TxFeeCap }}

# FilterCap sets the global cap for total number of filters that can be created
filter-cap = {{ .JSONRPC.FilterCap }}

# FeeHistoryCap sets the global cap for total number of blocks that can be fetched
feehistory-cap = {{ .JSONRPC.FeeHistoryCap }}

# LogsCap defines the max number of results can be returned from single 'eth_getLogs' query.
logs-cap = {{ .JSONRPC.LogsCap }}

# BlockRangeCap defines the max block range allowed for 'eth_getLogs' query.
block-range-cap = {{ .JSONRPC.BlockRangeCap }}

# HTTPTimeout is the read/write timeout of http json-rpc server.
http-timeout = "{{ .JSONRPC.HTTPTimeout }}"

# HTTPIdleTimeout is the idle timeout of http json-rpc server.
http-idle-timeout = "{{ .JSONRPC.HTTPIdleTimeout }}"

# AllowUnprotectedTxs restricts unprotected (non EIP155 signed) transactions to be submitted via
# the node's RPC when the global parameter is disabled.
allow-unprotected-txs = {{ .JSONRPC.AllowUnprotectedTxs }}

# MaxOpenConnections sets the maximum number of simultaneous connections
# for the server listener.
max-open-connections = {{ .JSONRPC.MaxOpenConnections }}

# EnableIndexer enables the custom transaction indexer for the EVM (ethereum transactions).
enable-indexer = {{ .JSONRPC.EnableIndexer }}

# MetricsAddress defines the EVM Metrics server address to bind to. Pass --metrics in CLI to enable
# Prometheus metrics path: /debug/metrics/prometheus
metrics-address = "{{ .JSONRPC.MetricsAddress }}"

# Upgrade height for fix of revert gas refund logic when transaction reverted.
fix-revert-gas-refund-height = {{ .JSONRPC.FixRevertGasRefundHeight }}

###############################################################################
###                             TLS Configuration                           ###
###############################################################################

[tls]

# Certificate path defines the cert.pem file path for the TLS configuration.
certificate-path = "{{ .TLS.CertificatePath }}"

# Key path defines the key.pem file path for the TLS configuration.
key-path = "{{ .TLS.KeyPath }}"
`

const DefaultRosettaConfigTemplate = `
###############################################################################
###                           Rosetta Configuration                         ###
###############################################################################

[rosetta]

# Enable defines if the Rosetta API server should be enabled.
enable = {{ .Rosetta.Enable }}

# Address defines the Rosetta API server to listen on.
address = "{{ .Rosetta.Config.Addr }}"

# Network defines the name of the blockchain that will be returned by Rosetta.
blockchain = "{{ .Rosetta.Config.Blockchain }}"

# Network defines the name of the network that will be returned by Rosetta.
network = "{{ .Rosetta.Config.Network }}"

# TendermintRPC defines the endpoint to connect to CometBFT RPC,
# specifying 'tcp://' before is not required, usually it's at port 26657
tendermint-rpc = "{{ .Rosetta.Config.TendermintRPC }}"

# GRPCEndpoint defines the cosmos application gRPC endpoint
# usually it is located at 9090 port
grpc-endpoint = "{{ .Rosetta.Config.GRPCEndpoint }}"

# Retries defines the number of retries when connecting to the node before failing.
retries = {{ .Rosetta.Config.Retries }}

# Offline defines if Rosetta server should run in offline mode.
offline = {{ .Rosetta.Config.Offline }}

# EnableFeeSuggestion indicates to use fee suggestion when 'construction/metadata' is called without gas limit and price.
enable-fee-suggestion = {{ .Rosetta.Config.EnableFeeSuggestion }}

# GasToSuggest defines gas limit when calculating the fee
gas-to-suggest = {{ .Rosetta.Config.GasToSuggest }}

# DenomToSuggest defines the defult denom for fee suggestion.
# Price must be in minimum-gas-prices.
denom-to-suggest = "{{ .Rosetta.Config.DenomToSuggest }}"

# GasPrices defines the gas prices for fee suggestion
gas-prices = "{{ .Rosetta.Config.GasPrices }}"
`

const DefaultVersionDBTemplate = `
###############################################################################
###                         VersionDB Configuration                         ###
###############################################################################

[versiondb]

# Enable defines if the versiondb should be enabled.
enable = {{ .VersionDB.Enable }}
`
