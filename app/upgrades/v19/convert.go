// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	"encoding/json"
	"os"
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
	fixes "github.com/evmos/evmos/v18/app/upgrades/v19/fixes"
	evmostypes "github.com/evmos/evmos/v18/types"
	erc20keeper "github.com/evmos/evmos/v18/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v18/x/evm/keeper"
)

// storeKey contains the slot in which the balance is stored in the evm.
var (
	storeKey       []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	storeKeyWevmos []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
)

type parseTokenPairs = []common.Address

// BalanceResult contains the data needed to perform the balance conversion
type BalanceResult struct {
	address      sdk.AccAddress
	balanceBytes []byte
	id           int
}



func GetMissingWalletsFromAuthModule(ctx sdk.Context,
	accountKeeper authkeeper.AccountKeeper,
) (Addresses []sdk.AccAddress) {
	xenAccounts:=0
	missingAccounts:=0
	wallets := fixes.GetAllMissingWallets()

	filter := GenerateFilter(accountKeeper, ctx)

	for _, wallet := range wallets {
		ethAddr := common.HexToAddress(wallet)
		addr := sdk.AccAddress(ethAddr.Bytes())

		if accountKeeper.HasAccount(ctx, addr) {
			account := accountKeeper.GetAccount(ctx, addr)
			ethAccount, ok := account.(*evmostypes.EthAccount)
			if !ok {
				fmt.Println("Account existed")
				continue
			}

			if IsAccountValid(ethAccount.EthAddress().Hex(), ethAccount.GetCodeHash(), filter) {
				fmt.Println("Account existed")
				continue
			}
			xenAccounts++
		}else{
			missingAccounts++
		}


		Addresses = append(Addresses, addr)
	}
	fmt.Println("xen",xenAccounts)
	fmt.Println("missingAccounts",missingAccounts)

	return Addresses
}


// executeConversion receives the whole set of adress with erc20 balances
// it sends the equivalent coin from the escrow address into the holder address
// it doesnt need to burn the erc20 balance, because the evm storage will be deleted later
func executeConversion(
	ctx sdk.Context,
	results []BalanceResult,
	bankKeeper bankkeeper.Keeper,
	wrappedEvmosAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	wevmosAccount := sdk.AccAddress(wrappedEvmosAddr.Bytes())
	// Go trough every address with an erc20 balance
	t := 0
	for _, result := range results {
		if t% 1_000 == 0{
			fmt.Println("exec",t)
		}
		t++
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
	var finalizedResults []BalanceResult

	missingAccounts := GetMissingWalletsFromAuthModule(ctx, accountKeeper)
	filter := GenerateFilter(accountKeeper, ctx)

	for _, account := range missingAccounts {
		addBalances(ctx, account, evmKeeper, wrappedAddr.Hex(), nativeTokenPairs, &finalizedResults)
	}

	i := 0
	j := 0
	// should ignore the xen token accounts
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		i++
		if i % 100_000 == 0{
			fmt.Println("processing account", i)
		}

		ethAccount, ok := account.(*evmostypes.EthAccount)
		if ok {
			if !IsAccountValid(ethAccount.EthAddress().Hex(), ethAccount.GetCodeHash(), filter) {
				j++
				if j % 100_000 == 0{
					fmt.Println("ignoring accounts",j)
				}
				return false
			} 
		}

		addBalances(ctx, account.GetAddress(), evmKeeper, wrappedAddr.Hex(), nativeTokenPairs, &finalizedResults)
		return false
	})

	logger.Info(fmt.Sprint("Finalized results: ", len(finalizedResults)))

	// Save addrs
	logger.Info("Saving valid addrs to json file")
	type ValidAddrs struct{
		Values []string `json:"values"`
	}
	addresses := ValidAddrs{Values:[]string{}}
	for _,v:=range finalizedResults{
		addresses.Values = append(addresses.Values, v.address.String())
	}
	jsonData, _ := json.Marshal(addresses)
	file, err := os.Create("addresses_valid.json")
	if err != nil{
		fmt.Println("error creating json file", err.Error())
	} else{
		defer file.Close()
		_, err := file.Write(jsonData)
		if err != nil{
			fmt.Println("error saving json file", err.Error())
		}
	}

	// execute the actual conversion.
	err = executeConversion(ctx, finalizedResults, bankKeeper, wrappedAddr, nativeTokenPairs)
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
		return fmt.Errorf("there are still tokens in the erc-20 module account: %s", balances.String())
	}
	duration := time.Since(timeBegin)

	// Panic at the end to stop execution
	panic(fmt.Sprintf("Finalized results len %d %s", len(finalizedResults), duration.String()))
}

func addBalances(
	ctx sdk.Context,
	account sdk.AccAddress,
	evmKeeper evmkeeper.Keeper,
	wrappedAddr string,
	nativeTokenPairs []erc20types.TokenPair,
	balances *[]BalanceResult) {
	concatBytes := append(common.LeftPadBytes(account.Bytes(), 32), storeKey...)
	key := crypto.Keccak256Hash(concatBytes)

	concatBytesWevmos := append(common.LeftPadBytes(account.Bytes(), 32), storeKeyWevmos...)
	keyWevmos := crypto.Keccak256Hash(concatBytesWevmos)
	var value []byte
	for tokenId, tokenPair := range nativeTokenPairs {
		if tokenPair.Erc20Address == wrappedAddr {
			value = evmKeeper.GetFastState(ctx, tokenPair.GetERC20Contract(), keyWevmos)
		} else {
			value = evmKeeper.GetFastState(ctx, tokenPair.GetERC20Contract(), key)
		}
		if len(value) == 0 {
			continue
		}
		*balances = append(*balances, BalanceResult{address: account, balanceBytes: value, id: tokenId})
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
