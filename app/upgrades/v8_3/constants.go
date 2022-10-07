package v83

const (
	// UpgradeName is the shared upgrade plan name for mainnet
	UpgradeName = "v8.3.0"
	// MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	MainnetUpgradeHeight = 5_888_000
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/amd64":"https://github.com/evmos/evmos/releases/download/v8.2.0/evmos_8.3.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evmos/evmos/releases/download/v8.3.0/evmos_8.3.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v8.3.0/evmos_8.3.0_Linux_arm64.tar.gz","linux/amd64":"https://github.com/evmos/evmos/releases/download/v8.3.0/evmos_8.3.0_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v8.3.0/evmos_8.3.0_Windows_x86_64.zip"}}'`
)
