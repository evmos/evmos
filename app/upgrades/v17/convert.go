// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v16/contracts"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"golang.org/x/sync/errgroup"
)

var storeKey []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}

type TelemetryResult struct {
	address string
	balance string
	id      int
}

// worker performs the task on jobs received and sends results to the results channel.
func worker(
	workerCtx context.Context,
	logger log.Logger,
	id int,
	tasks <-chan []string,
	results chan<- []TelemetryResult,
	ctx sdk.Context,
	evmKeeper evmkeeper.Keeper,
	nativeTokenPairs parseTokenPairs,
) error {
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return nil // Channel closed, stop the worker
			}

			if id%10 == 0 {
				logger.Info(fmt.Sprintf("Worker %d received task", id))
			}
			processResults, err := PerformTask(logger, task, id, ctx, evmKeeper, nativeTokenPairs)
			if err != nil {
				return err
			}
			if len(processResults) == 0 {
				continue
			}
			logger.Info("adding to results channel")
			results <- processResults
			logger.Info(fmt.Sprintf("Worker %d sent %d results to main results channel", id, len(processResults)))
		case <-workerCtx.Done():
			logger.Error(fmt.Sprintf("worker %d is done", id))
			return nil
		}
	}
}

var balancesCounter int

func PerformTask(logger log.Logger, task []string, id int,
	ctx sdk.Context, evmKeeper evmkeeper.Keeper, tokenPairs parseTokenPairs,
) ([]TelemetryResult, error) {
	results := make([]TelemetryResult, 0, len(task))

	for _, account := range task {
		found := false
		cosmosAddress := sdk.MustAccAddressFromBech32(account)
		ethAddress := common.BytesToAddress(cosmosAddress.Bytes())
		addrBytes := ethAddress.Bytes()
		concatBytes := append(common.LeftPadBytes(addrBytes, 32), storeKey...)
		key := crypto.Keccak256Hash(concatBytes)
		for id, pair := range tokenPairs {
			state := evmKeeper.GetState(ctx, pair, key)
			stateHex := state.Hex()
			// TODO: move this to rama's branch
			if stateHex == "0x0000000000000000000000000000000000000000000000000000000000000000" {
				// we continue here early to save creating a new big int for the zero cases
				continue
			}
			found = true
			balance, _ := new(big.Int).SetString(stateHex, 0)
			if balance.Sign() > 0 {
				results = append(results, TelemetryResult{address: account, balance: balance.String(), id: id})
			}
		}

		// TODO: remove logging here
		if found {
			balancesCounter++
			logger.Info(fmt.Sprintf("found %d accounts with balances so far.", balancesCounter))
		}
	}
	logger.Info(fmt.Sprintf("Worker %d is done processed task and got %d results", id, len(task)))
	return results, nil
}

var batchCounter int

func orchestrator(workerCtx context.Context, logger log.Logger, tasks chan<- []string, accountKeeper authkeeper.AccountKeeper, batchSize int,
	ctx sdk.Context,
) {
	currentBatch := make([]string, 0, batchSize)
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

		currentBatch = append(currentBatch, account.GetAddress().String())
		// Check if the current batch is filled or it's the last element.
		if (i+1)%batchSize == 0 {
			batchCounter++
			logger.Info(fmt.Sprintf("Sending batch: %d (len: %d)", batchCounter, len(currentBatch)))
			tasks <- currentBatch
			currentBatch = nil // Reset current batch
		}
		i++
		return false
	})
	tasks <- currentBatch
}

var resultsCounter int

func processResults(results <-chan []TelemetryResult, logger log.Logger) []TelemetryResult {
	finalizedResults := make([]TelemetryResult, 0)
	for batchResults := range results {
		for i := range batchResults {
			logger.Info(
				fmt.Sprintf(
					"Processed results: %d, results size: %d",
					resultsCounter,
					len(finalizedResults),
				),
			)
			resultsCounter++
			finalizedResults = append(finalizedResults, batchResults[i])
		}
	}
	return finalizedResults
}

func executeConversionBatch(
	ctx sdk.Context,
	logger log.Logger,
	results []TelemetryResult,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	for _, result := range results {

		cosmosAddress := sdk.MustAccAddressFromBech32(result.address)
		ethAddress := common.BytesToAddress(cosmosAddress.Bytes())
		ethHexAddr := ethAddress.String()
		tokenPair := nativeTokenPairs[result.id]

		if tokenPair.GetERC20Contract() == wrappedAddr {

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
				return err
			} else if res != nil && res.VmError != "" {
				logger.Error(
					"withdraw WEVMOS reverted",
					"account", ethHexAddr,
					"balance", bs,
					"vm-error", res.VmError,
				)
			}
		} else {

			n := new(big.Int)
			n, _ = n.SetString(result.balance, 10)
			coins := sdk.Coins{sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(n)}}

			// Unescrow coins and send to receiver
			err := bankKeeper.SendCoinsFromModuleToAccount(ctx, erc20types.ModuleName, cosmosAddress, coins)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

type parseTokenPairs = []common.Address

type TelemetryResult2 struct {
	address sdk.AccAddress
	balance []byte
	id      int
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
	fmt.Println("CORESSS WE ARE USING", runtime.NumCPU())
	numWorkers := runtime.NumCPU()
	batchSize := 1000
	g := new(errgroup.Group)
	// Create a context to cancel the workers in case of an error
	g, workerCtx := errgroup.WithContext(context.Background())

	// Create buffered channels for tasks and results
	tasks := make(chan []sdk.AccAddress, numWorkers)
	results := make(chan []TelemetryResult2, numWorkers)

	tokenPairs := make(parseTokenPairs, len(nativeTokenPairs))
	for i := range nativeTokenPairs {
		tokenPairs[i] = nativeTokenPairs[i].GetERC20Contract()
	}

	fmt.Println("This is the number of token pairs: ", len(tokenPairs))

	// Fan-out: Create worker goroutines
	for w := 1; w <= numWorkers; w++ {
		func(w int) {
			g.Go(func() error {
				return Worker2(ctx, workerCtx, w, tasks, results, evmKeeper, tokenPairs)
			})
		}(w)
	}

	// Create a goroutine to send tasks to workers
	go func() {
		orchestrator2(ctx, workerCtx, tasks, accountKeeper, batchSize)
		close(tasks)
	}()

	// Create a goroutine to wait for all workers to finish
	// check if there is an error and close the results channel
	go func() {
		if err := g.Wait(); err == nil {
			fmt.Println("All workers have finalized")
		} else {
			fmt.Println("Error received: ", err)
		}
		close(results)
	}()

	// Process results as they come in
	finalizedResults := processResults2(results)
	if g.Wait() != nil {
		err := g.Wait()
		fmt.Println("Context is cancelled we are destroying everything")
		fmt.Println(err)
	} else {
		fmt.Println("Completed Finalized results: ", len(finalizedResults))
	}
	return nil
}

func orchestrator2(ctx sdk.Context, workerCtx context.Context, tasks chan<- []sdk.AccAddress, accountKeeper authkeeper.AccountKeeper, batchSize int) {
	logger := ctx.Logger()
	currentBatch := make([]sdk.AccAddress, batchSize)
	counter := 0
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
			logger.Info(fmt.Sprintf("----------- Sending account # %v", counter))
			tasks <- currentBatch
			i = 0
		}
		counter++
		return false
	})
	tasks <- currentBatch[:i]
}

func processResults2(results <-chan []TelemetryResult2) []TelemetryResult2 {
	finalizedResults := make([]TelemetryResult2, 0)
	for batchResults := range results {
		fmt.Println("Results with balances: ", len(finalizedResults))
		finalizedResults = append(finalizedResults, batchResults...)
	}
	return finalizedResults
}

// worker performs the task on jobs received and sends results to the results channel.
func Worker2(
	sdkCtx sdk.Context,
	ctx context.Context,
	id int,
	tasks <-chan []sdk.AccAddress,
	results chan<- []TelemetryResult2,
	evmKeeper evmkeeper.Keeper,
	nativeTokenPairs parseTokenPairs,
) error {
	logger := sdkCtx.Logger()
	var resultsCol []TelemetryResult2
	evmKeeper.SetStorageDummy(sdkCtx)

	// leftPad := make([]byte, 64)
	// for k := range storeKey {
	// 	leftPad[32+k] = storeKey[k]
	// }
	tp := make([][]byte, len(nativeTokenPairs))
	for i, pair := range nativeTokenPairs {
		tp[i] = evmtypes.AddressStoragePrefix(pair)
	}

	// tokenPairStores := make([]sdk.KVStore, len(nativeTokenPairs))
	// for i, pair := range nativeTokenPairs {
	// 	tokenPairStores[i] = evmKeeper.GetStoreDummy(sdkCtx, pair)
	// }

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
				logger.Error(now.String())
			}
			for _, account := range task {
				concatBytes := append(common.LeftPadBytes(account.Bytes(), 32), storeKey...)
				key := crypto.Keccak256Hash(concatBytes)
				for _, pair := range tp {
					value := evmKeeper.PerformantGet(key.Bytes(), pair)
					if len(value) == 0 {
						continue
					}
					resultsCol = append(resultsCol, TelemetryResult2{address: account, balance: value, id: id})
				}
			}
			if id == 1 {
				logger.Error(time.Since(now).String())
			}

			// logger.Info(fmt.Sprintf("Worker %d is done processed task and got %d results", id, len(resultsCol)))
			if len(resultsCol) > 0 {
				results <- resultsCol
				resultsCol = nil
			}

			// resultsCol := make([]TelemetryResult2, len(task))
			// // resp := fmt.Sprintf("Worker %d processing address: %s\n", id, task[i])
			// // create a slice of slice of strings
			// for i, account := range task {

			// 	for x := range nativeTokenPairs {

			// 	}
			// 	resultsCol[i] = TelemetryResult2{address: account, balance: nil, id: 3}
			// }
			// results <- resultsCol
			// resultsCol = nil

		case <-ctx.Done():
			fmt.Sprintf("This is my important exit")
			// Context cancelled, stop the worker
			// fmt.Printf("Worker %d stopping due to cancellation\n", id)
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

