package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	evmostypes "github.com/evmos/evmos/v20/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.False(t, cfg.JSONRPC.Enable)
	require.Equal(t, cfg.JSONRPC.Address, DefaultJSONRPCAddress)
	require.Equal(t, cfg.JSONRPC.WsAddress, DefaultJSONRPCWsAddress)
}

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    func() *viper.Viper
		want    func() Config
		wantErr bool
	}{
		{
			"test unmarshal embedded structs",
			func() *viper.Viper {
				v := viper.New()
				v.Set("minimum-gas-prices", fmt.Sprintf("100%s", evmostypes.AttoEvmos))
				return v
			},
			func() Config {
				cfg := DefaultConfig()
				cfg.MinGasPrices = fmt.Sprintf("100%s", evmostypes.AttoEvmos)
				return *cfg
			},
			false,
		},
		{
			"test unmarshal EVMConfig",
			func() *viper.Viper {
				v := viper.New()
				v.Set("evm.tracer", "struct")
				return v
			},
			func() Config {
				cfg := DefaultConfig()
				require.NotEqual(t, "struct", cfg.EVM.Tracer)
				cfg.EVM.Tracer = "struct"
				return *cfg
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig(tt.args())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want()) {
				t.Errorf("GetConfig() got = %v, want %v", got, tt.want())
			}
		})
	}
}
