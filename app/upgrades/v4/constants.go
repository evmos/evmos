package v4

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v4.0.0"
	// MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	MainnetUpgradeHeight = 581700 // TODO: Update
	// TestnetUpgradeHeight defines the Evmos testnet block height on which the upgrade will take place
	TestnetUpgradeHeight = 5817000 // TODO: Update
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/tharsis/evmos/releases/download/v4.0.0/evmos_4.0.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/tharsis/evmos/releases/download/v4.0.0/evmos_4.0.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/tharsis/evmos/releases/download/v4.0.0/evmos_4.0.0_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/tharsis/evmos/releases/download/v4.0.0/evmos_4.0.0_Linux_x86_64.tar.gz","windows/x86_64":"https://github.com/tharsis/evmos/releases/download/v4.0.0/evmos_4.0.0_Windows_x86_64.zip"}}'`
)
