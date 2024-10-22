// Copyright Tharsis Labs Ltd.(Eidon-chain)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/Eidon-AI/eidon-chain/blob/main/LICENSE)

package tests

import (
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	eidon-chaintypes "github.com/Eidon-AI/eidon-chain/v20/types"
)

var (
	UosmoDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uosmo",
	}
	UosmoIbcdenom = UosmoDenomtrace.IBCDenom()

	UatomDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "uatom",
	}
	UatomIbcdenom = UatomDenomtrace.IBCDenom()

	Ueidon-chainDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: eidon-chaintypes.BaseDenom,
	}
	Ueidon-chainIbcdenom = Ueidon-chainDenomtrace.IBCDenom()

	UatomOsmoDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0/transfer/channel-1",
		BaseDenom: "uatom",
	}
	UatomOsmoIbcdenom = UatomOsmoDenomtrace.IBCDenom()

	Aeidon-chainDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: eidon-chaintypes.BaseDenom,
	}
	Aeidon-chainIbcdenom = Aeidon-chainDenomtrace.IBCDenom()
)
