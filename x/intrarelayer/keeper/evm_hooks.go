package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	// "github.com/tharsis/evmos/solidity/contracts"

	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var _ evmtypes.EvmHooks = (*Keeper)(nil)

// PostTxProcessing implements EvmHooks.PostTxProcessing
func (k Keeper) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}

		_ = log.Topics[0] // event ID

		// TODO: switch and handle events
		// switch event {
		// // case Burn
		// // case Mint
		// default:
		// 	continue
		// }

		// check that the contract is a registered token pair
		contractAddr := log.Address
		id := k.GetERC20Map(ctx, contractAddr) // TODO: rename

		if len(id) == 0 {
			// no token is registered for the caller contract
			continue
		}

		pair, found := k.GetTokenPair(ctx, id)
		if !found {
			continue
		}

		// check that relaying for the pair is enabled
		if !pair.Enabled {
			return fmt.Errorf("internal relaying is disabled for pair %s, please create a governance proposal", contractAddr) // convert to SDK error
		}

		// get the contract ABI
		// _, err := abi.JSON(strings.NewReader(contracts.ContractsABI))
		// if err != nil {
		// 	// the contract is not an ERC20Burnable
		// 	continue
		// }

		// contractABI.Events[event]

		// check that the event is Burn from the ERC20Burnable interface
		// NOTE: assume that if they are burning the token that has been registered as a pair, they want to mint a Cosmos coin

		// get the amount burned and the caller address
		// compare the caller address with the owner address and only mint if the burner is different from owner

		// create the corresponding sdk.Coin that is paired with ERC20
		coins := sdk.Coins{
			{
				Denom:  pair.Denom,
				Amount: sdk.ZeroInt(), // FIXME: get amount from event
			},
		}

		// Mint the coin
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			return err
		}

		// transfer to caller address
		recipient := sdk.AccAddress{} // FIXME: get caller from event
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
			return err
		}
	}

	return nil
}
