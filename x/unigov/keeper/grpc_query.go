package keeper

import (
	"github.com/Canto-Network/canto/v3/x/unigov/types"
)

var _ types.QueryServer = Keeper{}
