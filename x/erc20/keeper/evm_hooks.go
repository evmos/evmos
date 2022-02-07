package keeper

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/contracts"
	"github.com/tharsis/evmos/x/erc20/types"
)

// Hooks wrapper struct for erc20 keeper
type Hooks struct {
	k Keeper
}

var _ evmtypes.EvmHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// TODO: Make sure that if ConvertERC20 is called, that the Hook doesnt trigger
// if it does, delete minting from ConvertErc20

// PostTxProcessing implements EvmHooks.PostTxProcessing
func (h Hooks) PostTxProcessing(
	ctx sdk.Context,
	from common.Address,
	to *common.Address,
	receipt *ethtypes.Receipt,
) error {
	params := h.k.GetParams(ctx)
	if !params.EnableEVMHook {
		return sdkerrors.Wrap(types.ErrInternalTokenPair, "EVM Hook is currently disabled")
	}

	erc20 := contracts.ERC20BurnableContract.ABI

	for i, log := range receipt.Logs {
		if len(log.Topics) < 3 {
			continue
		}

		eventID := log.Topics[0] // event ID

		event, err := erc20.EventByID(eventID)
		if err != nil {
			// invalid event for ERC20
			continue
		}

		if event.Name != types.ERC20EventTransfer {
			h.k.Logger(ctx).Info("emitted event", "name", event.Name, "signature", event.Sig)
			continue
		}

		transferEvent, err := erc20.Unpack(event.Name, log.Data)
		if err != nil {
			h.k.Logger(ctx).Error("failed to unpack transfer event", "error", err.Error())
			continue
		}

		if len(transferEvent) == 0 {
			continue
		}

		tokens, ok := transferEvent[0].(*big.Int)
		// safety check and ignore if amount not positive
		if !ok || tokens == nil || tokens.Sign() != 1 {
			continue
		}

		// check that the contract is a registered token pair
		contractAddr := log.Address

		id := h.k.GetERC20Map(ctx, contractAddr)

		if len(id) == 0 {
			// no token is registered for the caller contract
			continue
		}

		pair, found := h.k.GetTokenPair(ctx, id)
		if !found {
			continue
		}

		// check that relaying for the pair is enabled
		if !pair.Enabled {
			return fmt.Errorf("internal relaying is disabled for pair %s, please create a governance proposal", contractAddr) // convert to SDK error
		}

		// ignore as the burning always transfers to the zero address
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
		case types.OWNER_MODULE:
			_, err = h.k.CallEVM(ctx, erc20, types.ModuleAddress, contractAddr, "burn", tokens)
		case types.OWNER_EXTERNAL:
			err = h.k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
		default:
			err = types.ErrUndefinedOwner
		}

		if err != nil {
			h.k.Logger(ctx).Debug(
				"failed to process EVM hook for ER20 -> coin conversion",
				"coin", pair.Denom, "contract", pair.Erc20Address, "error", err.Error(),
			)
			continue
		}

		// Only need last 20 bytes from log.topics
		from := common.BytesToAddress(log.Topics[1].Bytes())
		recipient := sdk.AccAddress(from.Bytes())

		// transfer the tokens from ModuleAccount to sender address
		if err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
			h.k.Logger(ctx).Debug(
				"failed to process EVM hook for ER20 -> coin conversion",
				"tx-hash", receipt.TxHash.Hex(), "log-idx", i,
				"coin", pair.Denom, "contract", pair.Erc20Address, "error", err.Error(),
			)
			continue
		}
	}

	return nil
}
