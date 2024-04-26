// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	erc20keeper "github.com/evmos/evmos/v18/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// ConvertERC20Coins converts Native IBC coins from their ERC20 representation
// to the native representation. This also includes the withdrawal of WEVMOS tokens
// to EVMOS native tokens.
func ConvertERC20Coins(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	// iterate over all the accounts and convert the tokens to native coins
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		cosmosAddress := account.GetAddress()
		ethAddress := common.BytesToAddress(cosmosAddress.Bytes())
		ethHexAddr := ethAddress.String()

		balance, res, err := WithdrawWEVMOS(ctx, ethAddress, wrappedAddr, erc20Keeper)

		var bs string // NOTE: this is necessary so that there is no panic if balance is nil when logging
		if balance != nil {
			bs = balance.String()
		}

		if err != nil {
			logger.Error(
				"failed to withdraw WEVMOS",
				"account", ethHexAddr,
				"balance", bs,
				"error", err.Error(),
			)
		} else if res != nil && res.VmError != "" {
			logger.Error(
				"withdraw WEVMOS reverted",
				"account", ethHexAddr,
				"balance", bs,
				"vm-error", res.VmError,
			)
		}

		for _, tokenPair := range nativeTokenPairs {
			contract := tokenPair.GetERC20Contract()
			if err := ConvertERC20Token(ctx, ethAddress, contract, cosmosAddress, erc20Keeper); err != nil {
				logger.Error(
					"failed to convert ERC20 to native Coin",
					"account", ethHexAddr,
					"erc20", contract.String(),
					"balance", balance.String(),
					"error", err.Error(),
				)
			}
		}

		return false
	})

	// NOTE: if there are tokens left in the ERC-20 module account
	// we return an error because this implies that the migration of native
	// coins to ERC-20 tokens was not fully completed.
	erc20ModuleAccountAddress := authtypes.NewModuleAddress(erc20types.ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, erc20ModuleAccountAddress)
	if !balances.IsZero() {
		return fmt.Errorf("there are still tokens in the erc-20 module account: %s", balances.String())
	}

	return nil
}

// getNativeTokenPairs returns the token pairs that are registered for native Cosmos coins.
func getNativeTokenPairs(
	ctx sdk.Context,
	erc20Keeper erc20keeper.Keeper,
) []erc20types.TokenPair {
	var nativeTokenPairs []erc20types.TokenPair

	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
		// NOTE: here we check if the token pair contains an IBC coin. For now, we only want to convert those.
		if !tokenPair.IsNativeCoin() {
			return false
		}

		nativeTokenPairs = append(nativeTokenPairs, tokenPair)
		return false
	})

	return nativeTokenPairs
}

// ConvertERC20Token converts the given ERC20 token to the native representation.
func ConvertERC20Token(
	ctx sdk.Context,
	from, contract common.Address,
	receiver sdk.AccAddress,
	erc20Keeper erc20keeper.Keeper,
) error {
	balance := erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, contract, from)
	if balance == nil {
		return fmt.Errorf("failed to get ERC20 balance (contract %q) for %s", contract.String(), from.String())
	}

	if balance.Sign() <= 0 {
		return nil
	}

	msg := erc20types.NewMsgConvertERC20(sdk.NewIntFromBigInt(balance), receiver, contract, from)
	_, err := erc20Keeper.ConvertERC20(sdk.WrapSDKContext(ctx), msg)

	return err
}

// WithdrawWEVMOS withdraws all the WEVMOS tokens from the given account.
func WithdrawWEVMOS(
	ctx sdk.Context,
	from, wevmosContract common.Address,
	erc20Keeper erc20keeper.Keeper,
) (*big.Int, *evmtypes.MsgEthereumTxResponse, error) {
	balance := erc20Keeper.BalanceOf(ctx, contracts.WEVMOSContract.ABI, wevmosContract, from)
	if balance == nil {
		return common.Big0, nil, fmt.Errorf("failed to get WEVMOS balance for %s", from.String())
	}

	// only execute the withdrawal if balance is positive
	if balance.Sign() <= 0 {
		return common.Big0, nil, nil
	}

	// call withdraw method from the account
	data, err := contracts.WEVMOSContract.ABI.Pack("withdraw", balance)
	if err != nil {
		return balance, nil, errorsmod.Wrap(err, "failed to pack data for withdraw method")
	}

	res, err := erc20Keeper.CallEVMWithData(ctx, from, &wevmosContract, data, true)
	return balance, res, err
}
