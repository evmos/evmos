// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package v82

const (
	// UpgradeName is the shared upgrade plan name for mainnet
	UpgradeName = "v8.2.0"
	// MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	MainnetUpgradeHeight = 4_888_000
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/amd64":"https://github.com/evmos/evmos/releases/download/v8.2.0/evmos_8.2.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evmos/evmos/releases/download/v8.2.0/evmos_8.2.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v8.2.0/evmos_8.2.0_Linux_arm64.tar.gz","linux/amd64":"https://github.com/evmos/evmos/releases/download/v8.2.0/evmos_8.2.0_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v8.2.0/evmos_8.2.0_Windows_x86_64.zip"}}'`
)
