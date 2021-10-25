package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

var _ evmtypes.EvmHooks = (*Keeper)(nil)

// PostTxProcessing implements EvmHooks.PostTxProcessing
func (k Keeper) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	erc20 := contracts.ERC20BurnableContract.ABI

	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}

		fmt.Println(log.Address.String())

		eventID := log.Topics[0] // event ID

		// check that the contract is a registered token pair
		contractAddr := log.Address
		fmt.Println(contractAddr.String())
		id := k.GetERC20Map(ctx, contractAddr)

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

		event, err := erc20.EventByID(eventID)
		if err != nil {
			// invalid event for ERC20
			continue
		}

		if event.Name != types.ERC20EventTransfer {
			k.Logger(ctx).Info("emitted event", "name", event.Name, "signature", event.Sig)
			continue
		}

		//FIX types.LogBurn should be equal to the unpacking event
		var transferEvent types.LogBurn
		err = erc20.UnpackIntoInterface(&transferEvent, event.Name, log.Data)
		if err != nil {
			k.Logger(ctx).Error("failed to unpack transfer event", "error", err.Error())
			continue
		}

		// safety check and ignore if amount not positive
		if transferEvent.Tokens == nil || transferEvent.Tokens.Sign() != 1 {
			continue
		}

		// // ignore as the burning always transfers to the zero address
		// if !bytes.Equal(transferEvent.To.Bytes(), common.Address{}.Bytes()) {
		// 	continue
		// }

		// check that the event is Burn from the ERC20Burnable interface
		// NOTE: assume that if they are burning the token that has been registered as a pair, they want to mint a Cosmos coin

		// create the corresponding sdk.Coin that is paired with ERC20
		coins := sdk.Coins{{Denom: pair.Denom, Amount: sdk.NewIntFromBigInt(transferEvent.Tokens)}}

		// Mint the coin
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			return err
		}

		//FIX if we use the burn method, we need to extract the sender from somewhere else
		// transfer to caller address
		// recipient := sdk.AccAddress(transferEvent.From.Bytes())
		// if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
		// 	return err
		// }
	}

	return nil
}
