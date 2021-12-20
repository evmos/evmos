package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var _ evmtypes.EvmHooks = (*Keeper)(nil)

// TODO: Make sure that if ConvertERC20 is called, that the Hook doesnt trigger
// if it does, delete minting from ConvertErc20

// PostTxProcessing implements EvmHooks.PostTxProcessing
func (k Keeper) PostTxProcessing(ctx sdk.Context, from common.Address, to *common.Address, receipt *ethtypes.Receipt) error {

	return nil
}
