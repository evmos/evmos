// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package config

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/spf13/viper"

	"github.com/cometbft/cometbft/libs/strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/rosetta"

	"github.com/crypto-org-chain/cronos/memiavl"
	memiavlcfg "github.com/crypto-org-chain/cronos/store/config"

	_ "github.com/evmos/evmos/v20/server/config/migration" // Add this import to set up the proper app.toml migration logic for sdk v0.50
)

const (
	// ServerStartTime defines the time duration that the server need to stay running after startup
	// for the startup be considered successful
	ServerStartTime = 5 * time.Second

	// DefaultAPIEnable is the default value for the parameter that defines if the cosmos REST API server is enabled
	DefaultAPIEnable = false

	// DefaultGRPCEnable is the default value for the parameter that defines if the gRPC server is enabled
	DefaultGRPCEnable = false

	// DefaultGRPCWebEnable is the default value for the parameter that defines if the gRPC web server is enabled
	DefaultGRPCWebEnable = false

	// DefaultJSONRPCEnable is the default value for the parameter that defines if the JSON-RPC server is enabled
	DefaultJSONRPCEnable = false

	// DefaultRosettaEnable is the default value for the parameter that defines if the Rosetta API server is enabled
	DefaultRosettaEnable = false

	// DefaultTelemetryEnable is the default value for the parameter that defines if the telemetry is enabled
	DefaultTelemetryEnable = false

	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// DefaultJSONRPCAddress is the default address the JSON-RPC server binds to.
	DefaultJSONRPCAddress = "127.0.0.1:8545"

	// DefaultJSONRPCWsAddress is the default address the JSON-RPC WebSocket server binds to.
	DefaultJSONRPCWsAddress = "127.0.0.1:8546"

	// DefaultJsonRPCMetricsAddress is the default address the JSON-RPC Metrics server binds to.
	DefaultJSONRPCMetricsAddress = "127.0.0.1:6065"

	// DefaultEVMTracer is the default vm.Tracer type
	DefaultEVMTracer = ""

	// DefaultFixRevertGasRefundHeight is the default height at which to overwrite gas refund
	DefaultFixRevertGasRefundHeight = 0

	// DefaultMaxTxGasWanted is the default gas wanted for each eth tx returned in ante handler in check tx mode
	DefaultMaxTxGasWanted = 0

	// DefaultGasCap is the default cap on gas that can be used in eth_call/estimateGas
	DefaultGasCap uint64 = 25000000

	// DefaultJSONRPCAllowInsecureUnlock is true
	DefaultJSONRPCAllowInsecureUnlock bool = true

	// DefaultFilterCap is the default cap for total number of filters that can be created
	DefaultFilterCap int32 = 200

	// DefaultFeeHistoryCap is the default cap for total number of blocks that can be fetched
	DefaultFeeHistoryCap int32 = 100

	// DefaultLogsCap is the default cap of results returned from single 'eth_getLogs' query
	DefaultLogsCap int32 = 10000

	// DefaultBlockRangeCap is the default cap of block range allowed for 'eth_getLogs' query
	DefaultBlockRangeCap int32 = 10000

	// DefaultEVMTimeout is the default timeout for eth_call
	DefaultEVMTimeout = 5 * time.Second

	// DefaultTxFeeCap is the default tx-fee cap for sending a transaction
	DefaultTxFeeCap float64 = 1.0

	// DefaultHTTPTimeout is the default read/write timeout of the http json-rpc server
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPIdleTimeout is the default idle timeout of the http json-rpc server
	DefaultHTTPIdleTimeout = 120 * time.Second

	// DefaultAllowUnprotectedTxs value is false
	DefaultAllowUnprotectedTxs = false

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0

	// DefaultGasAdjustment value to use as default in gas-adjustment flag
	DefaultGasAdjustment = 1.2

	// DefaultRosettaBlockchain defines the default blockchain name for the rosetta server
	DefaultRosettaBlockchain = "evmos"

	// DefaultRosettaNetwork defines the default network name for the rosetta server
	DefaultRosettaNetwork = "evmos"

	// DefaultRosettaGasToSuggest defines the default gas to suggest for the rosetta server
	DefaultRosettaGasToSuggest = 300_000

	// DefaultRosettaDenomToSuggest defines the default denom for fee suggestion
	DefaultRosettaDenomToSuggest = "aevmos"

	// ============================
	//           MemIAVL
	// ============================

	// DefaultMemIAVLEnable is the default value that defines if memIAVL is enabled
	DefaultMemIAVLEnable = false

	// DefaultZeroCopy is the default value that defines if
	// the zero-copied slices must be retained beyond current block's execution
	// the sdk address cache will be disabled if zero-copy is enabled
	DefaultZeroCopy = false

	// DefaultAsyncCommitBuffer value to use as default for the size of
	// asynchronous commit queue when using memIAVL
	DefaultAsyncCommitBuffer = 0

	// DefaultSnapshotKeepRecent default value for how many old snapshots
	// (excluding the latest one) should be kept after new snapshots
	// when using memIAVL
	DefaultSnapshotKeepRecent = 1

	// ============================
	//           VersionDB
	// ============================

	// DefaultVersionDBEnable is the default value that defines if versionDB is enabled
	DefaultVersionDBEnable = false
)

// DefaultRosettaGasPrices defines the default list of prices to suggest
var DefaultRosettaGasPrices = sdk.NewDecCoins(sdk.NewDecCoin(DefaultRosettaDenomToSuggest, math.NewInt(4_000_000)))

var evmTracers = []string{"json", "markdown", "struct", "access_list"}

// Config defines the server's top level configuration. It includes the default app config
// from the SDK as well as the EVM configuration to enable the JSON-RPC APIs.
type Config struct {
	config.Config `mapstructure:",squash"`

	EVM     EVMConfig     `mapstructure:"evm"`
	JSONRPC JSONRPCConfig `mapstructure:"json-rpc"`
	TLS     TLSConfig     `mapstructure:"tls"`
	Rosetta RosettaConfig `mapstructure:"rosetta"`

	MemIAVL   MemIAVLConfig   `mapstructure:"memiavl"`
	VersionDB VersionDBConfig `mapstructure:"versiondb"`
}

// EVMConfig defines the application configuration values for the EVM.
type EVMConfig struct {
	// Tracer defines vm.Tracer type that the EVM will use if the node is run in
	// trace mode. Default: 'json'.
	Tracer string `mapstructure:"tracer"`
	// MaxTxGasWanted defines the gas wanted for each eth tx returned in ante handler in check tx mode.
	MaxTxGasWanted uint64 `mapstructure:"max-tx-gas-wanted"`
}

// JSONRPCConfig defines configuration for the EVM RPC server.
type JSONRPCConfig struct {
	// API defines a list of JSON-RPC namespaces that should be enabled
	API []string `mapstructure:"api"`
	// Address defines the HTTP server to listen on
	Address string `mapstructure:"address"`
	// WsAddress defines the WebSocket server to listen on
	WsAddress string `mapstructure:"ws-address"`
	// GasCap is the global gas cap for eth-call variants.
	GasCap uint64 `mapstructure:"gas-cap"`
	// AllowInsecureUnlock toggles if account unlocking is enabled when account-related RPCs are exposed by http.
	AllowInsecureUnlock bool `mapstructure:"allow-insecure-unlock"`
	// EVMTimeout is the global timeout for eth-call.
	EVMTimeout time.Duration `mapstructure:"evm-timeout"`
	// TxFeeCap is the global tx-fee cap for send transaction
	TxFeeCap float64 `mapstructure:"txfee-cap"`
	// FilterCap is the global cap for total number of filters that can be created.
	FilterCap int32 `mapstructure:"filter-cap"`
	// FeeHistoryCap is the global cap for total number of blocks that can be fetched
	FeeHistoryCap int32 `mapstructure:"feehistory-cap"`
	// Enable defines if the EVM RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// LogsCap defines the max number of results can be returned from single `eth_getLogs` query.
	LogsCap int32 `mapstructure:"logs-cap"`
	// BlockRangeCap defines the max block range allowed for `eth_getLogs` query.
	BlockRangeCap int32 `mapstructure:"block-range-cap"`
	// HTTPTimeout is the read/write timeout of http json-rpc server.
	HTTPTimeout time.Duration `mapstructure:"http-timeout"`
	// HTTPIdleTimeout is the idle timeout of http json-rpc server.
	HTTPIdleTimeout time.Duration `mapstructure:"http-idle-timeout"`
	// AllowUnprotectedTxs restricts unprotected (non EIP155 signed) transactions to be submitted via
	// the node's RPC when global parameter is disabled.
	AllowUnprotectedTxs bool `mapstructure:"allow-unprotected-txs"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
	// EnableIndexer defines if enable the custom indexer service.
	EnableIndexer bool `mapstructure:"enable-indexer"`
	// MetricsAddress defines the metrics server to listen on
	MetricsAddress string `mapstructure:"metrics-address"`
	// FixRevertGasRefundHeight defines the upgrade height for fix of revert gas refund logic when transaction reverted
	FixRevertGasRefundHeight int64 `mapstructure:"fix-revert-gas-refund-height"`
}

// TLSConfig defines the certificate and matching private key for the server.
type TLSConfig struct {
	// CertificatePath the file path for the certificate .pem file
	CertificatePath string `mapstructure:"certificate-path"`
	// KeyPath the file path for the key .pem file
	KeyPath string `mapstructure:"key-path"`
}

// RosettaConfig defines configuration for the Rosetta server.
type RosettaConfig struct {
	rosetta.Config
	// Enable defines if the Rosetta server should be enabled.
	Enable bool `mapstructure:"enable"`
}

// MemIAVLConfig defines the configuration for memIAVL.
type MemIAVLConfig struct {
	memiavlcfg.MemIAVLConfig
}

// VersionDBConfig defines the configuration for versionDB.
type VersionDBConfig struct {
	// Enable defines if the versiondb should be enabled.
	Enable bool `mapstructure:"enable"`
}

// AppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func AppConfig(denom string) (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	customAppConfig := DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In evmos, we set the min gas prices to 0.
	if denom != "" {
		customAppConfig.Config.MinGasPrices = "0" + denom
	}

	customAppTemplate := config.DefaultConfigTemplate +
		DefaultEVMConfigTemplate +
		DefaultRosettaConfigTemplate +
		DefaultVersionDBTemplate +
		memiavlcfg.DefaultConfigTemplate

	return customAppTemplate, *customAppConfig
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	defaultSDKConfig := config.DefaultConfig()
	defaultSDKConfig.API.Enable = DefaultAPIEnable
	defaultSDKConfig.GRPC.Enable = DefaultGRPCEnable
	defaultSDKConfig.GRPCWeb.Enable = DefaultGRPCWebEnable
	defaultSDKConfig.Telemetry.Enabled = DefaultTelemetryEnable

	return &Config{
		Config:    *defaultSDKConfig,
		EVM:       *DefaultEVMConfig(),
		JSONRPC:   *DefaultJSONRPCConfig(),
		TLS:       *DefaultTLSConfig(),
		Rosetta:   *DefaultRosettaConfig(),
		MemIAVL:   *DefaultMemIAVLConfig(),
		VersionDB: *DefaultVersionDBConfig(),
	}
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultEVMConfig() *EVMConfig {
	return &EVMConfig{
		Tracer:         DefaultEVMTracer,
		MaxTxGasWanted: DefaultMaxTxGasWanted,
	}
}

// Validate returns an error if the tracer type is invalid.
func (c EVMConfig) Validate() error {
	if c.Tracer != "" && !strings.StringInSlice(c.Tracer, evmTracers) {
		return fmt.Errorf("invalid tracer type %s, available types: %v", c.Tracer, evmTracers)
	}

	return nil
}

// GetDefaultAPINamespaces returns the default list of JSON-RPC namespaces that should be enabled
func GetDefaultAPINamespaces() []string {
	return []string{"eth", "net", "web3"}
}

// GetAPINamespaces returns the all the available JSON-RPC API namespaces.
func GetAPINamespaces() []string {
	return []string{"web3", "eth", "personal", "net", "txpool", "debug", "miner"}
}

// DefaultJSONRPCConfig returns an EVM config with the JSON-RPC API enabled by default
func DefaultJSONRPCConfig() *JSONRPCConfig {
	return &JSONRPCConfig{
		Enable:                   false,
		API:                      GetDefaultAPINamespaces(),
		Address:                  DefaultJSONRPCAddress,
		WsAddress:                DefaultJSONRPCWsAddress,
		GasCap:                   DefaultGasCap,
		AllowInsecureUnlock:      DefaultJSONRPCAllowInsecureUnlock,
		EVMTimeout:               DefaultEVMTimeout,
		TxFeeCap:                 DefaultTxFeeCap,
		FilterCap:                DefaultFilterCap,
		FeeHistoryCap:            DefaultFeeHistoryCap,
		BlockRangeCap:            DefaultBlockRangeCap,
		LogsCap:                  DefaultLogsCap,
		HTTPTimeout:              DefaultHTTPTimeout,
		HTTPIdleTimeout:          DefaultHTTPIdleTimeout,
		AllowUnprotectedTxs:      DefaultAllowUnprotectedTxs,
		MaxOpenConnections:       DefaultMaxOpenConnections,
		EnableIndexer:            false,
		MetricsAddress:           DefaultJSONRPCMetricsAddress,
		FixRevertGasRefundHeight: DefaultFixRevertGasRefundHeight,
	}
}

// Validate returns an error if the JSON-RPC configuration fields are invalid.
func (c JSONRPCConfig) Validate() error {
	if c.Enable && len(c.API) == 0 {
		return errors.New("cannot enable JSON-RPC without defining any API namespace")
	}

	if c.FilterCap < 0 {
		return errors.New("JSON-RPC filter-cap cannot be negative")
	}

	if c.FeeHistoryCap <= 0 {
		return errors.New("JSON-RPC feehistory-cap cannot be negative or 0")
	}

	if c.TxFeeCap < 0 {
		return errors.New("JSON-RPC tx fee cap cannot be negative")
	}

	if c.EVMTimeout < 0 {
		return errors.New("JSON-RPC EVM timeout duration cannot be negative")
	}

	if c.LogsCap < 0 {
		return errors.New("JSON-RPC logs cap cannot be negative")
	}

	if c.BlockRangeCap < 0 {
		return errors.New("JSON-RPC block range cap cannot be negative")
	}

	if c.HTTPTimeout < 0 {
		return errors.New("JSON-RPC HTTP timeout duration cannot be negative")
	}

	if c.HTTPIdleTimeout < 0 {
		return errors.New("JSON-RPC HTTP idle timeout duration cannot be negative")
	}

	// check for duplicates
	seenAPIs := make(map[string]bool)
	for _, api := range c.API {
		if seenAPIs[api] {
			return fmt.Errorf("repeated API namespace '%s'", api)
		}

		seenAPIs[api] = true
	}

	return nil
}

// DefaultTLSConfig returns the default TLS configuration
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		CertificatePath: "",
		KeyPath:         "",
	}
}

// Validate returns an error if the TLS certificate and key file extensions are invalid.
func (c TLSConfig) Validate() error {
	certExt := path.Ext(c.CertificatePath)

	if c.CertificatePath != "" && certExt != ".pem" {
		return fmt.Errorf("invalid extension %s for certificate path %s, expected '.pem'", certExt, c.CertificatePath)
	}

	keyExt := path.Ext(c.KeyPath)

	if c.KeyPath != "" && keyExt != ".pem" {
		return fmt.Errorf("invalid extension %s for key path %s, expected '.pem'", keyExt, c.KeyPath)
	}

	return nil
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultRosettaConfig() *RosettaConfig {
	return &RosettaConfig{
		Config: rosetta.Config{
			Blockchain:          DefaultRosettaBlockchain,
			Network:             DefaultRosettaNetwork,
			TendermintRPC:       rosetta.DefaultCometEndpoint,
			GRPCEndpoint:        rosetta.DefaultGRPCEndpoint,
			Addr:                rosetta.DefaultAddr,
			Retries:             rosetta.DefaultRetries,
			Offline:             rosetta.DefaultOffline,
			EnableFeeSuggestion: rosetta.DefaultEnableFeeSuggestion,
			GasToSuggest:        DefaultRosettaGasToSuggest,
			DenomToSuggest:      DefaultRosettaDenomToSuggest,
			GasPrices:           DefaultRosettaGasPrices,
		},
		Enable: DefaultRosettaEnable,
	}
}

// DefaultVersionDBConfig returns the default versionDB configuration
func DefaultVersionDBConfig() *VersionDBConfig {
	return &VersionDBConfig{
		Enable: DefaultVersionDBEnable,
	}
}

// DefaultMemIAVLConfig returns the default MemIAVL configuration
func DefaultMemIAVLConfig() *MemIAVLConfig {
	return &MemIAVLConfig{memiavlcfg.MemIAVLConfig{
		Enable:             DefaultMemIAVLEnable,
		ZeroCopy:           DefaultZeroCopy,
		AsyncCommitBuffer:  DefaultAsyncCommitBuffer,
		SnapshotKeepRecent: DefaultSnapshotKeepRecent,
		SnapshotInterval:   memiavl.DefaultSnapshotInterval,
		CacheSize:          memiavlcfg.DefaultCacheSize,
	}}
}

// Validate returns an error if the MemIAVL configuration fields are invalid.
func (c MemIAVLConfig) Validate() error {
	// AsyncCommitBuffer can be -1, which means synchronous commit
	if c.AsyncCommitBuffer < -1 {
		return errors.New("AsyncCommitBuffer cannot be negative")
	}

	if c.CacheSize < 0 {
		return errors.New("CacheSize cannot be negative")
	}

	return nil
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) (Config, error) {
	conf := DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return Config{}, fmt.Errorf("error extracting app config: %w", err)
	}
	return *conf, nil
}

// ValidateBasic returns an error any of the application configuration fields are invalid
func (c Config) ValidateBasic() error {
	if err := c.EVM.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid evm config value: %s", err.Error())
	}

	if err := c.JSONRPC.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid json-rpc config value: %s", err.Error())
	}

	if err := c.TLS.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid tls config value: %s", err.Error())
	}

	if err := c.MemIAVL.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid memIAVL config value: %s", err.Error())
	}

	return c.Config.ValidateBasic()
}
