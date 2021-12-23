package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
)

// DistributeFees allocates the tx fees to the block proposer and the
// address registered in the fee distribution module.
func (k Keeper) DistributeFees(ctx sdk.Context, feesPayed sdk.Coins, contract common.Address) error {
	params := k.GetParams(ctx)

	if params.ContractDistribution.IsZero() {
		// don't distribute fees to registered contract withdraw address
		return nil
	}

	// use the mint denomination since the fees are distributed to the fee collector
	mintDenom := k.mintKeeper.GetParams(ctx).MintDenom

	feeAmt := feesPayed.AmountOf(mintDenom)
	if !feeAmt.IsPositive() {
		return nil
	}

	feeDec := sdk.NewDecFromInt(feeAmt)

	withdrawAddr, found := k.GetContractWithdrawAddress(ctx, contract)
	if !found {
		// return nil if the contract withdraw address is not registered for
		// fee distribution
		return nil
	}

	logger := k.Logger(ctx)

	payableAmt := feeDec.Mul(params.ContractDistribution).TruncateInt()
	payable := sdk.Coins{{Denom: mintDenom, Amount: payableAmt}}

	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, withdrawAddr.Bytes(), payable)
	if err != nil {
		logger.Debug(
			"failed to transfer from module account to withdraw address",
			"fee-collector", k.feeCollectorName,
			"fees", payable.String(),
			"error", err.Error(),
		)
		return err
	}

	logger.Info(
		"distributed tx fees",
		"contract-address", contract.Hex(),
		"withdraw-address", withdrawAddr.Hex(),
		"fees", payable.String(),
	)

	return nil
}
