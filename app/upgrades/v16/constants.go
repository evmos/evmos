// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

const (
	// UpgradeName is the shared upgrade plan name for mainnet
	UpgradeName = "v16.0.0"
	// UpgradeNameTestnetRC2 is the shared upgrade plan name for testnet rc2 upgrade
	UpgradeNameTestnetRC2 = "v16.0.0-rc2"
	// UpgradeNameTestnetRC3 is the shared upgrade plan name for testnet rc3 patch upgrade
	UpgradeNameTestnetRC3 = "v16.0.0-rc3"
	// UpgradeNameTestnetRC4 is the shared upgrade plan name for testnet rc4 hard-fork upgrade
	UpgradeNameTestnetRC4 = "v16.0.0-rc4"
	// TestnetUpgradeHeight defines the Evmos testnet block height on which the rc3 upgrade will take place
	TestnetUpgradeHeight = 19465900
	// UpgradeNameTestnetRC5 is the shared upgrade plan name for testnet rc5 upgrade
	UpgradeNameTestnetRC5 = "v16.0.0-rc5"
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/amd64":"https://github.com/evmos/evmos/releases/download/v16.0.0-rc5/evmos_16.0.0-rc5_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evmos/evmos/releases/download/v16.0.0-rc5/evmos_16.0.0-rc5_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v16.0.0-rc5/evmos_16.0.0-rc5_Linux_arm64.tar.gz","linux/amd64":"https://github.com/evmos/evmos/releases/download/v16.0.0-rc5/evmos_16.0.0-rc5_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v16.0.0-rc5/evmos_16.0.0-rc5_Windows_x86_64.zip"}}'`
)
