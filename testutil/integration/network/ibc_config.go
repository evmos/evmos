package network

// IBCNetworkConfig is the configuration for the IBCNetwork
type IBCNetworkConfig struct {
	evmosConfig        Config
	numberOfChains     int
	validatorsPerChain int
}

// DefaultIBCNetworkConfig returns the default configuration for the IBCNetwork
func DefaultIBCNetworkConfig() IBCNetworkConfig {
	return IBCNetworkConfig{
		evmosConfig:        DefaultConfig(),
		numberOfChains:     2,
		validatorsPerChain: 2,
	}
}
