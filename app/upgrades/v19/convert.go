// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	fixes "github.com/evmos/evmos/v19/app/upgrades/v19/fixes"
	evmostypes "github.com/evmos/evmos/v19/types"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
)

// storeKey contains the slot in which the balance is stored in the evm.
var (
	storeKey       = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	storeKeyWevmos = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
)

// BalanceResult contains the data needed to perform the balance conversion
type BalanceResult struct {
	address      sdk.AccAddress
	balanceBytes []byte
	id           int
}

func GetMissingWalletsFromAuthModule(ctx sdk.Context,
	accountKeeper authkeeper.AccountKeeper,
) (addresses []sdk.AccAddress) {
	xenAccounts := 0
	missingAccounts := 0
	wallets := fixes.GetAllMissingWallets()

	filter := generateFilter(accountKeeper, ctx)

	for _, wallet := range wallets {
		ethAddr := common.HexToAddress(wallet)
		addr := sdk.AccAddress(ethAddr.Bytes())

		if accountKeeper.HasAccount(ctx, addr) {
			account := accountKeeper.GetAccount(ctx, addr)
			ethAccount, ok := account.(*evmostypes.EthAccount)
			if !ok {
				continue
			}

			if isAccountValid(ethAccount.EthAddress().Hex(), ethAccount.GetCodeHash(), filter) {
				continue
			}
			xenAccounts++
		} else {
			missingAccounts++
		}

		addresses = append(addresses, addr)
	}
	return addresses
}

// executeConversion receives the whole set of address with erc20 balances
// it sends the equivalent coin from the escrow address into the holder address
// it doesn't need to burn the erc20 balance, because the evm storage will be deleted later
func executeConversion(
	ctx sdk.Context,
	results []BalanceResult,
	bankKeeper bankkeeper.Keeper,
	wrappedEvmosAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	wevmosAccount := sdk.AccAddress(wrappedEvmosAddr.Bytes())
	// Go through every address with an erc20 balance
	for _, result := range results {
		tokenPair := nativeTokenPairs[result.id]

		// The conversion is different for Evmos/WEVMOS and IBC-coins
		// Convert balance Bytes into Big Int
		balance := new(big.Int).SetBytes(result.balanceBytes)
		if balance.Sign() <= 0 {
			continue
		}
		// Create the coin
		coins := sdk.Coins{sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(balance)}}

		// If its Wevmos
		if tokenPair.Erc20Address == wrappedEvmosAddr.Hex() {
			// Withdraw the balance from the contract
			// Unescrow coins and send to holder account
			err := bankKeeper.SendCoinsFromAccountToModule(ctx, wevmosAccount, erc20types.ModuleName, coins)
			if err != nil {
				return err
			}
		}

		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, erc20types.ModuleName, result.address, coins)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConvertERC20Coins iterates through all the authmodule accounts and all missing accounts from the auth module
// recovers the balance from erc20 contracts for the registered token pairs
// and for each entry it sends the balance from escrow into the account.
func ConvertERC20Coins(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	timeBegin := time.Now() // control the time of the execution
	var finalizedResults []BalanceResult

	missingAccounts := GetMissingWalletsFromAuthModule(ctx, accountKeeper)
	filter := generateFilter(accountKeeper, ctx)

	for _, account := range missingAccounts {
		addBalances(ctx, account, evmKeeper, wrappedAddr.Hex(), nativeTokenPairs, &finalizedResults)
	}

	i := 0
	// should ignore the xen token accounts
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		i++
		if i%100_000 == 0 {
			logger.Info(fmt.Sprintf("Processing account: %d", i))
		}

		ethAccount, ok := account.(*evmostypes.EthAccount)
		if ok {
			if !isAccountValid(ethAccount.EthAddress().Hex(), ethAccount.GetCodeHash(), filter) {
				return false
			}
		}

		addBalances(ctx, account.GetAddress(), evmKeeper, wrappedAddr.Hex(), nativeTokenPairs, &finalizedResults)
		return false
	})

	logger.Info(fmt.Sprint("Finalized results: ", len(finalizedResults)))

	// execute the actual conversion.
	err := executeConversion(ctx, finalizedResults, bankKeeper, wrappedAddr, nativeTokenPairs)
	if err != nil {
		// panic(err)
		return err
	}

	// NOTE: if there are tokens left in the ERC-20 module account
	// we return an error because this implies that the migration of native
	// coins to ERC-20 tokens was not fully completed.
	erc20ModuleAccountAddress := authtypes.NewModuleAddress(erc20types.ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, erc20ModuleAccountAddress)
	if !balances.IsZero() {
		logger.Info(fmt.Sprintf("there are still tokens in the erc-20 module account: %s", balances.String()))
		// we dont return an error here. Since we want the migration to pass
		// if any balance is left on escrow, we can recover it later.
	}
	duration := time.Since(timeBegin)
	logger.Info(fmt.Sprintf("Migration length %s", duration.String()))
	return nil
}

func addBalances(
	ctx sdk.Context,
	account sdk.AccAddress,
	evmKeeper evmkeeper.Keeper,
	wrappedAddr string,
	nativeTokenPairs []erc20types.TokenPair,
	balances *[]BalanceResult,
) {
	concatBytes := append(common.LeftPadBytes(account.Bytes(), 32), storeKey...)
	key := crypto.Keccak256Hash(concatBytes)

	concatBytesWevmos := append(common.LeftPadBytes(account.Bytes(), 32), storeKeyWevmos...)
	keyWevmos := crypto.Keccak256Hash(concatBytesWevmos)
	var value []byte
	for tokenID, tokenPair := range nativeTokenPairs {
		if tokenPair.Erc20Address == wrappedAddr {
			value = evmKeeper.GetFastState(ctx, tokenPair.GetERC20Contract(), keyWevmos)
		} else {
			value = evmKeeper.GetFastState(ctx, tokenPair.GetERC20Contract(), key)
		}
		if len(value) == 0 {
			continue
		}
		*balances = append(*balances, BalanceResult{address: account, balanceBytes: value, id: tokenID})
	}
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
