package v5

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v5.0.0"
	// MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	// TODO: define
	MainnetUpgradeHeight = 257_850
	// TestnetUpgradeHeight defines the Evmos testnet block height on which the upgrade will take place
	// TODO: define
	TestnetUpgradeHeight = 1_200_000
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Linux_x86_64.tar.gz","windows/x86_64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Windows_x86_64.zip"}}'`
	// ContributorAddrFrom is the lost address of an early contributor
	ContributorAddrFrom = "evmos13cf9npvns2vhh3097909mkhfxngmw6d6eppfm4"
	// ContributorAddrTo is the new address of an early contributor
	ContributorAddrTo = "evmos1hmntpkn623y3vl0nvzrazvq4rqzv3xa74l40gl"
)
