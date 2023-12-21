// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"fmt"
	"math/big"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/precompiles/werc20/testdata"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
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

		erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
			if !tokenPair.IsNativeCoin() {
				return false
			}

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

			return false
		})

		return false
	})

	erc20ModuleAccountAddress := authtypes.NewModuleAddress(erc20types.ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, erc20ModuleAccountAddress)
	if balances.IsZero() {
		return nil
	}

	// burn all the coins left in the module account
	return bankKeeper.BurnCoins(ctx, erc20types.ModuleName, balances)
}

// WithdrawWEVMOS withdraws all the WEVMOS tokens from the given account.
func WithdrawWEVMOS(
	ctx sdk.Context,
	from, wevmosContract common.Address,
	erc20Keeper erc20keeper.Keeper,
) (*big.Int, *evmtypes.MsgEthereumTxResponse, error) {
	balance := erc20Keeper.BalanceOf(ctx, testdata.WEVMOSContract.ABI, wevmosContract, from)
	if balance == nil {
		return common.Big0, nil, fmt.Errorf("failed to get WEVMOS balance for %s", from.String())
	}

	// only execute the withdrawal if balance is positive
	if balance.Sign() <= 0 {
		return common.Big0, nil, nil
	}

	// call withdraw method from the account
	//
	// TODO: implement call to the WEVMOS withdraw method (also the balance amount has to be passed)
	data, err := testdata.WEVMOSContract.ABI.Pack("withdraw", balance)
	if err != nil {
		fmt.Println("error packing data for withdraw method", err.Error())
		return balance, nil, err
	}

	res, err := erc20Keeper.CallEVMWithData(ctx, from, &wevmosContract, data, true)
	return balance, res, err
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
	if err != nil {
		return err
	}

	return nil
}
