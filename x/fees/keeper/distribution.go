package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
)

// DistributeFees allocates the tx fees to the block proposer and the
// address registered in the fee distribution module.
func (k Keeper) DistributeFees(ctx sdk.Context, feesPayed sdk.Coins, contract common.Address) error {
	if len(feesPayed) == 0 {
		return nil
	}

	params := k.GetParams(ctx)

	if params.FeesEnabled || params.DeveloperDistribution.IsZero() {
		// don't distribute fees to registered contract withdraw address
		return nil
	}

	withdrawAddr, found := k.GetContractWithdrawAddress(ctx, contract)
	if !found {
		// return nil if the contract withdraw address is not registered for
		// fee distribution
		return nil
	}

	logger := k.Logger(ctx)

	payableDev := make(sdk.Coins, len(feesPayed))
	payableBlockProposer := make(sdk.Coins, len(feesPayed))

	for i := range feesPayed {
		payableDev[i] = sdk.Coin{
			Denom:  feesPayed[i].Denom,
			Amount: sdk.NewDecFromInt(feesPayed[i].Amount).Mul(params.DeveloperDistribution).TruncateInt(),
		}

		payableBlockProposer[i] = sdk.Coin{
			Denom:  feesPayed[i].Denom,
			Amount: feesPayed[i].Amount.Sub(payableDev[i].Amount), // 100% - dev amount
		}
	}

	// TODO: this needs to be modified
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, withdrawAddr.Bytes(), payableDev)
	if err != nil {
		logger.Debug(
			"failed to transfer from module account to contract developer",
			"fee-collector", k.feeCollectorName,
			"fees", payableDev.String(),
			"error", err.Error(),
		)
		return err
	}

	logger.Info(
		"distributed tx fees",
		"contract-address", contract.Hex(),
		"withdraw-address", withdrawAddr.Hex(),
		"fees", payableDev.String(),
	)

	return nil
}
