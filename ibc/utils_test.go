package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("acre", "acrepub")
}
