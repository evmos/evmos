package ante

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/params"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	feesplittypes "github.com/evmos/evmos/v7/x/feesplit/types"
)

// EvmKeeper defines the expected keeper interface used on the AnteHandler
type EvmKeeper interface {
	GetParams(ctx sdk.Context) (params evmtypes.Params)
	ChainID() *big.Int
	GetBaseFee(ctx sdk.Context, ethCfg *params.ChainConfig) *big.Int
}

// FeesplitKeeper defines the expected keeper interface used on the AnteHandler
type FeesplitKeeper interface {
	GetParams(ctx sdk.Context) (params feesplittypes.Params)
}
