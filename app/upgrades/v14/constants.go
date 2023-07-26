// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14

// !! ATTENTION !!
// created this upgrade folder to include the upgrade handler needed
// when upgrading to cosmos-sdk v0.47
// If v14 is not including this upgrade,
// make sure to move the store upgrades needed (in app.go) and
// the upgrade handler (in upgrades.go) to the corresponding upgrade
// source: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xconsensus
// !! ATTENTION !!

const (
	// UpgradeName is the shared upgrade plan name for mainnet
	UpgradeName = "v14.0.0"
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/evmos/evmos/releases/download/v14.0.0/evmos_14.0.0_Darwin_arm64.tar.gz","darwin/amd64":"https://github.com/evmos/evmos/releases/download/v14.0.0/evmos_14.0.0_Darwin_amd64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v14.0.0/evmos_14.0.0_Linux_arm64.tar.gz","linux/amd64":"https://github.com/evmos/evmos/releases/download/v14.0.0/evmos_14.0.0_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v14.0.0/evmos_14.0.0_Windows_x86_64.zip"}}'`
)
