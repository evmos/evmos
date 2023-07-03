package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.False(t, cfg.JSONRPC.Enable)
	require.Equal(t, cfg.JSONRPC.Address, DefaultJSONRPCAddress)
	require.Equal(t, cfg.JSONRPC.WsAddress, DefaultJSONRPCWsAddress)
}
