package keeper

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

var _ evmtypes.EvmHooks = (*Keeper)(nil)

// TODO: Make sure that if ConvertERC20 is called, that the Hook doesnt trigger
// if it does, delete minting from ConvertErc20

// PostTxProcessing implements EvmHooks.PostTxProcessing
func (k Keeper) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	erc20 := contracts.ERC20BurnableContract.ABI

	for _, log := range logs {
		if len(log.Topics) < 3 {
			continue
		}

		eventID := log.Topics[0] // event ID

		// check that the contract is a registered token pair
		contractAddr := log.Address

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

		burnEvent, err := erc20.Unpack(event.Name, log.Data)
		if err != nil {
			k.Logger(ctx).Error("failed to unpack transfer event", "error", err.Error())
			continue
		}

		if len(burnEvent) == 0 {
			continue
		}

		tokens, ok := burnEvent[0].(*big.Int)
		// safety check and ignore if amount not positive
		if !ok || tokens == nil || tokens.Sign() != 1 {
			continue
		}

		// ignore as the burning always transfers to the moduleAddress
		to := common.BytesToAddress(log.Topics[2].Bytes())
		if !bytes.Equal(to.Bytes(), types.ModuleAddress.Bytes()) {
			continue
		}

		// check that the event is Burn from the ERC20Burnable interface
		// NOTE: assume that if they are burning the token that has been registered as a pair, they want to mint a Cosmos coin

		// create the corresponding sdk.Coin that is paired with ERC20
		coins := sdk.Coins{{Denom: pair.Denom, Amount: sdk.NewIntFromBigInt(tokens)}}

		// Mint the coin only if ERC20 is external
		switch pair.ContractOwner {
		case types.MODULE_OWNER:
			// BURN ERC20 FROM MODULE
			_, err := k.CallEVM(ctx, erc20, contractAddr, "burn", tokens)
			if err != nil {
				continue
			}
		case types.EXTERNAL_OWNER:
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
				continue
			}
		}

		// transfer to caller address
		from := common.BytesToAddress(log.Topics[1].Bytes())
		recipient := sdk.AccAddress(from.Bytes())
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
			continue
		}
	}

	return nil
}
