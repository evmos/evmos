package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	erc20types "github.com/evmos/evmos/v12/x/erc20/types"
)

type ERC20BankContractRegistrationHook struct {
	erc20Keeper Keeper
}

// NewERC20ContractRegistrationHook returns the DenomMetadataHooks for ERC20 registration
func NewERC20ContractRegistrationHook(erc20Keeper Keeper) ERC20BankContractRegistrationHook {
	return ERC20BankContractRegistrationHook{
		erc20Keeper: erc20Keeper,
	}
}

func (e ERC20BankContractRegistrationHook) AfterDenomMetadataCreation(ctx sdk.Context, newDenomMetadata banktypes.Metadata) error {
	if e.erc20Keeper.IsERC20Enabled(ctx) && strings.HasPrefix(strings.ToLower(newDenomMetadata.Base), "ibc/") { // only deploy for IBC denom.
		// Mint the erc20 coin for the new IBC denom.
		// TODO: is this acceptable? otherwise coin cannot be registered in erc20 because of no supply
		if err := e.erc20Keeper.bankKeeper.MintCoins(ctx, erc20types.ModuleName, sdk.Coins{sdk.NewInt64Coin(newDenomMetadata.Base, 1)}); err != nil {
			return fmt.Errorf("failed to mint the erc20 coin: %s; error: %w", newDenomMetadata.Base, err)
		}
		// Deploy the erc20 contract for the new IBC denom.
		// Error, if any, no state transition will be made.
		_, err := e.erc20Keeper.RegisterCoin(ctx, newDenomMetadata)
		if err != nil {
			// TODO: what happens if this fails? We have the coin registered in bank, but not in erc20. Should we remove the coin from bank?
			return fmt.Errorf("failed to deploy the erc20 contract for the IBC coin: %s; error: %w", newDenomMetadata.Base, err)
		}
	}

	return nil
}

func (e ERC20BankContractRegistrationHook) AfterDenomMetadataUpdate(sdk.Context, banktypes.Metadata) error {
	// do nothing

	return nil
}
