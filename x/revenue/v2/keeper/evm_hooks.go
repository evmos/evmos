// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

var outpostAddrThreshold = common.HexToAddress("0x0000000000000000000000000000000000000FFF")

var _ evmtypes.EvmHooks = Hooks{}

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer (or, if set,
// the withdraw address) receives a share from the transaction fees paid by the
// transaction sender.
func (k Keeper) PostTxProcessing(
	ctx sdk.Context,
	msg core.Message,
	receipt *ethtypes.Receipt,
) error {
	to := msg.To()
	// when baseFee and minGasPrice in freemarker module are both 0
	// the user may send a transaction with gasPrice of 0 to the precompiled contract
	if to == nil || msg.GasPrice().Sign() <= 0 {
		return nil
	}

	evmParams := k.EVMKeeper.GetParams(ctx)
	evmDenom := evmParams.EvmDenom
	chainCfg := evmParams.GetChainConfig()
	baseFeePerGas := k.EVMKeeper.GetBaseFee(ctx, &chainCfg)

	base := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(baseFeePerGas))
	baseFee := sdk.Coins{{Denom: evmDenom, Amount: base}}

	// token transfers don't have a defined data
	isTokenTransfer := len(msg.Data()) == 0
	// get active precompiles from EVM params and check if to (contract) is in the list
	isExtension := evmParams.IsActivePrecompile(to.String())
	// if the contract is a smart contract or an outpost, check if the contract is registered in the revenue module.
	// else, return and avoid performing unnecessary logic
	// NOTE: outposts have addresses addresses >= 0x…01000
	isOutpost := bytes.Compare(to.Bytes(), outpostAddrThreshold.Bytes()) > 0

	var err error
	switch {
	case isTokenTransfer:
		// Case 1: Token Transfer -> burn BaseFee
		err = k.burnBaseFee(ctx, baseFee)
	case isExtension && !isOutpost:
		// Case 2: EVM (Core) Extensions -> fund Community Pool
		err = k.fundCommunityPool(ctx, baseFee)
	case isExtension && isOutpost,
		!isExtension:
		// Case 3 and 4: Outposts and Smart Contracts-> allocate revenue, or
		// Case 5: burn if not registered
		internallCalls := []common.Address{*to}
		gasUsedByAddress := map[common.Address]uint64{
			*to: receipt.GasUsed,
		}
		err = k.allocateRevenue(ctx, internallCalls, gasUsedByAddress, evmDenom)
	}

	if err != nil {
		return err
	}

	return nil
}

// burnBaseFee burns the BaseFee amount of the transaction
func (k Keeper) burnBaseFee(ctx sdk.Context, baseFee sdk.Coins) error {
	if err := k.BankKeeper.BurnCoins(ctx, k.FeeCollectorName, baseFee); err != nil {
		return err
	}

	// TODO: emit Burn event from ERC-20 to the EVM

	return nil
}

func (k Keeper) fundCommunityPool(ctx sdk.Context, baseFee sdk.Coins) error {
	if err := k.DistributionKeeper.FundCommunityPool(ctx, baseFee, authtypes.NewModuleAddress(k.FeeCollectorName)); err != nil {
		return err
	}

	// TODO: emit Transfer event from ERC-20 to the EVM

	return nil
}

func (k Keeper) allocateRevenue(
	ctx sdk.Context,
	internalCalls []common.Address,
	gasUsedByAddress map[common.Address]uint64,
	denom string,
) error {
	if len(internalCalls) == 0 || len(internalCalls) != len(gasUsedByAddress) {
		return nil
	}
	// check if the fees are globally enabled
	params := k.GetParams(ctx)

	// check if the fees are globally enabled or if the
	// developer shares are set to zero
	if !params.EnableRevenue || params.DeveloperShares.IsZero() {
		return nil
	}

	burnedAmt := math.ZeroInt()

	for _, contract := range internalCalls {
		gasUsed := gasUsedByAddress[contract]
		if gasUsed == 0 {
			continue
		}

		gasUsedInt := math.NewIntFromUint64(gasUsed)

		revenue, found := k.GetRevenue(ctx, contract)
		if !found {
			burnedAmt = burnedAmt.Add(gasUsedInt)
			continue
		}

		// get withdrawer or deployer address
		withdrawer := revenue.GetWithdrawerAddr()
		if len(withdrawer) == 0 {
			withdrawer = revenue.GetDeployerAddr()
		}

		// allocate (baseFee * devShares) to developer and burn baseFee * (1 - devShares)
		devAllocation := params.DeveloperShares.MulInt(gasUsedInt).TruncateInt()
		burnedAllocation := gasUsedInt.Sub(devAllocation)
		developerFee := sdk.Coin{Amount: devAllocation, Denom: denom}
		burnedAmt = burnedAmt.Add(burnedAllocation)

		err := k.BankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			k.FeeCollectorName,
			withdrawer,
			sdk.Coins{developerFee},
		)
		if err != nil {
			return errorsmod.Wrapf(
				err,
				"fee collector account failed to distribute developer fees (%s) to withdraw address %s. contract %s",
				developerFee, withdrawer, contract,
			)

			// TODO: emit Transfer event from ERC-20 to the EVM
		}
	}

	// burn the the unregistered and leftover amount
	burnedCoins := sdk.Coins{{Amount: burnedAmt, Denom: denom}}

	if err := k.BankKeeper.BurnCoins(ctx, k.FeeCollectorName, burnedCoins); err != nil {
		return err
	}

	// TODO: emit Burn event from ERC-20 to the EVM

	return nil
}
