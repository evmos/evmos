// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v20

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	fixes "github.com/evmos/evmos/v18/app/upgrades/v20/fixes"
	erc20keeper "github.com/evmos/evmos/v18/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v18/x/evm/keeper"
)

// storeKey contains the slot in which the balance is stored in the evm.
var storeKey []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
var storeKeyWevmos []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}

type parseTokenPairs = []common.Address

// BalanceResult contains the data needed to perform the balance conversion
type BalanceResult struct {
	address      sdk.AccAddress
	balanceBytes []byte
	id           int
}

// ExportResult holds the data
// to be exported to a json file
type ExportResult struct {
	Address string
	Balance string
	Erc20   string
}

// executeConversion receives the whole set of adress with erc20 balances
// it sends the equivalent coin from the escrow address into the holder address
// it doesnt need to burn the erc20 balance, because the evm storage will be deleted later
func executeConversion(
	ctx sdk.Context,
	results []BalanceResult,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedEvmosAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
	fromSnapshot bool,
) error {
	wevmosAccount := sdk.AccAddress(wrappedEvmosAddr.Bytes())
	// Go trough every address with an erc20 balance
	for _, result := range results {
		if fromSnapshot && erc20Keeper.HasSTRv2Address(ctx, result.address) {
			continue
		}

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

// ConvertERC20Coins generates the list of address-erc20 balance that need to be migrated
// It takes some steps to generate this list (parallel)
//   - Divide all the accounts into smaller batches
//   - Have parallel workers query the db for erc20 balance
//   - Consolidate all the balances on the same array
//
// Once the list is generated, it does three things  (serialized)
//   - Save the result into a file
//   - Actually move all the balances from erc20 to bank
//   - Check that all the balances has been moved.
func ConvertERC20Coins(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	timeBegin := time.Now() // control the time of the execution

	// simplify the list of erc20 token pairs to handle less data
	tokenPairs := make(parseTokenPairs, len(nativeTokenPairs))
	for i := range nativeTokenPairs {
		tokenPairs[i] = nativeTokenPairs[i].GetERC20Contract()
	}

	modifiedBalancesAccounts := erc20Keeper.GetAllSTRV2Address(ctx)
	var modifiedBalancesWallets = make([]string, len(modifiedBalancesAccounts))
	for i, addr := range modifiedBalancesAccounts {
		modifiedBalancesWallets[i] = addr.String()
	}

	// Need to filter only uniques accounts
	missingWallets := fixes.GetMissingWalletsFromAuthModule(ctx, accountKeeper)
	combinedWallets := append(missingWallets, modifiedBalancesWallets...)
	slices.Sort(combinedWallets)
	combinedWallets = slices.Compact(combinedWallets)

	// Convert to accounts
	allAccounts := make([]sdk.AccAddress, len(combinedWallets))
	for i, wallet := range combinedWallets {
		allAccounts[i] = sdk.MustAccAddressFromBech32(wallet)
	}

	wevmosId := 0
	tokenPairStores := make([]sdk.KVStore, len(tokenPairs))
	for i, pair := range tokenPairs {
		tokenPairStores[i] = evmKeeper.GetStoreDummy(ctx, pair)
		if wrappedAddr.Hex() == pair.Hex() {
			wevmosId = i
		}
	}

	resultsCol := []BalanceResult{}
	for _, account := range allAccounts {
		concatBytes := append(common.LeftPadBytes(account.Bytes(), 32), storeKey...)
		key := crypto.Keccak256Hash(concatBytes)

		concatBytesWevmos := append(common.LeftPadBytes(account.Bytes(), 32), storeKeyWevmos...)
		keyWevmos := crypto.Keccak256Hash(concatBytesWevmos)
		var value []byte
		for tokenId, store := range tokenPairStores {
			if tokenId == wevmosId {
				value = store.Get(keyWevmos.Bytes())
				if len(value) == 0 {
					continue
				}
			} else {
				value = store.Get(key.Bytes())
				if len(value) == 0 {
					continue
				}
			}

			resultsCol = append(resultsCol, BalanceResult{address: account, balanceBytes: value, id: tokenId})
		}
	}

	err := executeConversion(ctx, resultsCol, bankKeeper, erc20Keeper, wrappedAddr, nativeTokenPairs, false)
	if err != nil {
		panic(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	filePath := filepath.Join(dir, "results-full.json")
	// Store in file
	// file, _ := json.MarshalIndent(jsonExport, "", " ")
	file, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var readResults []ExportResult
	err = json.Unmarshal(file, &readResults)
	fmt.Println("Finalized results: ", len(readResults))
	if err != nil {
		panic("Failed to unmarshal")
	}

	erc20Map := make(map[string]int)
	for i, erc20 := range nativeTokenPairs {
		erc20Map[erc20.Erc20Address] = i
	}

	// Generate the json to store in the file
	var finalizedResults []BalanceResult = make([]BalanceResult, len(readResults))
	for i, result := range readResults {
		b, _ := new(big.Int).SetString(result.Balance, 10)
		finalizedResults[i] = BalanceResult{
			address:      sdk.MustAccAddressFromBech32(result.Address),
			balanceBytes: b.Bytes(),
			id:           erc20Map[result.Erc20],
		}
	}
	// execute the actual conversion.
	err = executeConversion(ctx, finalizedResults, bankKeeper, erc20Keeper, wrappedAddr, nativeTokenPairs, true)
	if err != nil {
		panic(err)
	}

	// NOTE: if there are tokens left in the ERC-20 module account
	// we return an error because this implies that the migration of native
	// coins to ERC-20 tokens was not fully completed.
	erc20ModuleAccountAddress := authtypes.NewModuleAddress(erc20types.ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, erc20ModuleAccountAddress)
	if !balances.IsZero() {
		return fmt.Errorf("there are still tokens in the erc-20 module account: %s", balances.String())
	}
	duration := time.Since(timeBegin)
	logger.Info(fmt.Sprintf("STR v2 migration took %s\n", duration))
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
