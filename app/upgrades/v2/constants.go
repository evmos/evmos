package v2

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v2.0.0"
	// MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	MainnetUpgradeHeight = 58700 // (24 * 60 * 60) / 6 + 44300
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/evmos/evmos/releases/download/v2.0.0/evmos_2.0.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evmos/evmos/releases/download/v2.0.0/evmos_2.0.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v2.0.0/evmos_2.0.0_Linux_arm64.tar.gz","linux/amd64":"https://github.com/evmos/evmos/releases/download/v2.0.0/evmos_2.0.0_Linux_x86_64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v2.0.0/evmos_2.0.0_Windows_x86_64.zip"}}'`
)
