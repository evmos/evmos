package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v12/x/erc20/types"
)

type ERC20BankContractRegistrationHook struct {
	erc20Keeper keeper
}

type keeper interface {
	RegisterCoin(ctx sdk.Context, coinMetadata banktypes.Metadata) (*types.TokenPair, error)
}

const denomIBCPrefix = "ibc/"

// NewERC20ContractRegistrationHook returns the DenomMetadataHooks for ERC20 registration
func NewERC20ContractRegistrationHook(erc20Keeper keeper) ERC20BankContractRegistrationHook {
	return ERC20BankContractRegistrationHook{
		erc20Keeper: erc20Keeper,
	}
}

// AfterDenomMetadataCreation deploys the ERC20 contract for the IBC denom. It is called after the
// bank module creates the denom metadata. Without the ERC20 contract, the IBC denom cannot be
// converted to the ERC20 token, and the IBC transfer will fail.
func (e ERC20BankContractRegistrationHook) AfterDenomMetadataCreation(ctx sdk.Context, newDenomMetadata banktypes.Metadata) error {
	// only deploy for IBC denom.
	if !strings.HasPrefix(strings.ToLower(newDenomMetadata.Base), denomIBCPrefix) {
		return nil
	}

	if _, err := e.erc20Keeper.RegisterCoin(ctx, newDenomMetadata); err != nil {
		return fmt.Errorf("deploy the erc20 contract for the ibc coin: %s; error: %w", newDenomMetadata.Base, err)
	}

	return nil
}

func (e ERC20BankContractRegistrationHook) AfterDenomMetadataUpdate(sdk.Context, banktypes.Metadata) error {
	return fmt.Errorf("update the denom metadata while having an already existing contract")
}
