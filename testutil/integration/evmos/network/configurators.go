// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"cosmossdk.io/math"
	"github.com/evmos/evmos/v20/app"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

func Test18DecimalsAppConfigurator(chainID string) error {
	feemarkettypes.DefaultBaseFee = math.LegacyNewDec(1_000_000_000)
	PrefundedAccountInitialBalance, _ = math.NewIntFromString("100_000_000_000_000_000_000_000")
	return app.AppConfigurator(chainID)
}

func Test6DecimalsAppConfigurator(chainID string) error {
	feemarkettypes.DefaultBaseFee = math.LegacyNewDec(1)
	PrefundedAccountInitialBalance, _ = math.NewIntFromString("100_000_000_000")
	return app.AppConfigurator(chainID)
}
