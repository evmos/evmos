package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	// only deploy for IBC denom.
	if strings.HasPrefix(strings.ToLower(newDenomMetadata.Base), "ibc/") {
		return nil
	}

	if _, err := e.erc20Keeper.RegisterCoin(ctx, newDenomMetadata); err != nil {
		return fmt.Errorf("failed to deploy the erc20 contract for the IBC coin: %s; error: %w", newDenomMetadata.Base, err)
	}

	return nil
}

func (e ERC20BankContractRegistrationHook) AfterDenomMetadataUpdate(sdk.Context, banktypes.Metadata) error {
	// do nothing

	return nil
}
