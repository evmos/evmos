// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	fixes "github.com/evmos/evmos/v16/app/upgrades/v17/fixes"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	"golang.org/x/sync/errgroup"
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
	wrappedEvmosAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	wevmosAccount := sdk.AccAddress(wrappedEvmosAddr.Bytes())
	// Go trough every address with an erc20 balance
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
	numWorkers := 1         // ensure an efficient amount of workers

	// each coroutine will query `batchSize` amount of accounts
	batchSize := 1000

	g := new(errgroup.Group)
	// Create a context to cancel the workers in case of an error
	g, workerCtx := errgroup.WithContext(context.Background())

	// Create buffered channels for accountsBatch and results
	accountsBatch := make(chan []sdk.AccAddress, numWorkers)
	balancesResults := make(chan []BalanceResult, numWorkers)

	// simplify the list of erc20 token pairs to handle less data
	tokenPairs := make(parseTokenPairs, len(nativeTokenPairs))
	for i := range nativeTokenPairs {
		tokenPairs[i] = nativeTokenPairs[i].GetERC20Contract()
	}

	fmt.Println("This is the number of token pairs: ", len(tokenPairs))

	// Fan-out: Create worker goroutines
	// each worker will handle a batch of accounts, and query each of those accounts
	// if any account holds erc20 balance, its added to the balancesResults channel
	for w := 1; w <= numWorkers; w++ {
		func(workerId int) {
			g.Go(func() error {
				return Worker(ctx, workerCtx, workerId, accountsBatch, balancesResults, evmKeeper, tokenPairs, wrappedAddr)
			})
		}(w)
	}

	missingAccounts := fixes.GetMissingWalletsFromAuthModule(ctx, accountKeeper)
	fmt.Println("Missing accounts ", len(missingAccounts))
	accountsBatch <- missingAccounts

	// Create a goroutine to send tasks to workers
	// Once the orchestrator has finished processing all the accounts
	// Close the channel so the workers now they can end the processing
	go func() {
		orchestrator(ctx, workerCtx, accountsBatch, accountKeeper, batchSize)
		close(accountsBatch)
	}()

	// Create a goroutine to wait for all workers to finish
	// check if there is an error and close the results channel
	go func() {
		if err := g.Wait(); err == nil {
			fmt.Println("All workers have finalized")
		} else {
			fmt.Println("Error received: ", err)
		}
		close(balancesResults)
	}()

	// Process results as they come in
	finalizedResults := processResults(balancesResults)

	// If the wait group didnt close graciously, store the error
	if g.Wait() != nil {
		err := g.Wait()
		fmt.Println("Context is cancelled we are destroying everything")
		fmt.Println(err)
	} else {
		fmt.Println("Completed Finalized results: ", len(finalizedResults))
	}

	// Generate the json to store in the file
	var jsonExport []ExportResult = make([]ExportResult, len(finalizedResults))
	for i, result := range finalizedResults {
		jsonExport[i] = ExportResult{
			Address: result.address.String(),
			Balance: new(big.Int).SetBytes(result.balanceBytes).String(),
			Erc20:   nativeTokenPairs[result.id].Erc20Address,
		}
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Store in file
	file, _ := json.MarshalIndent(jsonExport, "", " ")
	_ = os.WriteFile(fmt.Sprint(userHomeDir, "/results-full.json"), file, os.ModePerm)

	fmt.Println("Finalized results: ", len(finalizedResults))
	// execute the actual conversion.
	err = executeConversion(ctx, finalizedResults, bankKeeper, wrappedAddr, nativeTokenPairs)
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

	// Panic at the end to stop execution
	panic(fmt.Sprintf("Finalized results len %d %s", len(finalizedResults), duration.String()))
}

// orchestrator is in charge to distribute the work among the workers.
// It iterates trough all the accounts on the database and creates the accounts batches that each worker is gonna receive
// when a worker is ready, it grabs one of the available accounts batches and start processing it
func orchestrator(ctx sdk.Context, workerCtx context.Context, tasks chan<- []sdk.AccAddress, accountKeeper authkeeper.AccountKeeper, batchSize int) {
	currentBatch := make([]sdk.AccAddress, batchSize)
	i := 0
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		if workerCtx.Err() != nil {
			if workerCtx.Err() == context.Canceled {
				fmt.Println("Context is cancelled")
			} else if workerCtx.Err() == context.DeadlineExceeded {
				fmt.Println("Deadline has been exceeded")
			}
			// If the context is already cancelled, stop sending tasks
			return true
		}

		currentBatch[i] = account.GetAddress()
		i++
		// Check if the current batch is filled or it's the last element.
		if i == batchSize {
			copyBatch := make([]sdk.AccAddress, batchSize)
			copy(copyBatch, currentBatch)
			// post the current batch of accounts to the tasks channel.
			// if the tasks channel is full, it wait heres until a worker grabs one of the batches
			tasks <- copyBatch
			i = 0
		}
		return false
	})

	// if last batch is empty return
	if i == 0 {
		return
	}
	// After iterating trough everything
	// send the remaining accounts on the last batch
	var copyBatch = make([]sdk.AccAddress, i)
	copy(copyBatch, currentBatch[:i])
	tasks <- copyBatch
}

// processResults continuously receives the results from the workers getting the erc20 balances
// each result is appended to the final results slice
func processResults(results <-chan []BalanceResult) []BalanceResult {
	finalizedResults := make([]BalanceResult, 0)
	for batchResults := range results {
		finalizedResults = append(finalizedResults, batchResults...)
	}
	return finalizedResults
}

// worker performs the task on jobs received and sends results to the results channel.
func Worker(
	sdkCtx sdk.Context,
	ctx context.Context,
	id int,
	tasks <-chan []sdk.AccAddress,
	results chan<- []BalanceResult,
	evmKeeper evmkeeper.Keeper,
	nativeTokenPairs parseTokenPairs,
	wrappedAddr common.Address,
) error {
	logger := sdkCtx.Logger()
	var resultsCol []BalanceResult

	wevmosId := 0
	tokenPairStores := make([]sdk.KVStore, len(nativeTokenPairs))
	for i, pair := range nativeTokenPairs {
		tokenPairStores[i] = evmKeeper.GetStoreDummy(sdkCtx, pair)
		if wrappedAddr.Hex() == pair.Hex() {
			wevmosId = i
		}
	}

	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				// fmt.Printf("Worker %d stopping due to channel closed\n", id)
				return nil // Channel closed, stop the worker
			}

			logger.Info(fmt.Sprintf("Worker %d got accounts", id))
			now := time.Now()
			if id == 1 {
				logger.Info(now.String())
			}
			for _, account := range task {
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

			if id == 1 {
				logger.Info(time.Since(now).String())
			}

			if len(resultsCol) > 0 {
				copyBatch := make([]BalanceResult, len(resultsCol))
				copy(copyBatch, resultsCol)
				results <- copyBatch
				resultsCol = nil
			}

		case <-ctx.Done():
			// Context cancelled, stop the worker
			return nil
		}
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
