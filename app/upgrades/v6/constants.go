package v6

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v6.0.1"
	// MainnetUpgradeHeight defines the Evoblock mainnet block height on which the upgrade will take place
	MainnetUpgradeHeight = 1_042_000
	// TestnetUpgradeHeight defines the Evoblock testnet block height on which the upgrade will take place
	TestnetUpgradeHeight = 2_176_500
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/evoblockchain/evoblock/releases/download/v6.0.1/evoblock_6.0.1_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evoblockchain/evoblock/releases/download/v6.0.1/evoblock_6.0.1_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evoblockchain/evoblock/releases/download/v6.0.1/evoblock_6.0.1_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/evoblockchain/evoblock/releases/download/v6.0.1/evoblock_6.0.1_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evoblockchain/evoblock/releases/download/v6.0.1/evoblock_6.0.1_Windows_x86_64.zip"}}'`
)
