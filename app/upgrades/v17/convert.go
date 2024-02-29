// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	"context"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"golang.org/x/sync/errgroup"
)

type TelemetryResult struct {
	address string
	balance string
	id      int
}

// worker performs the task on jobs received and sends results to the results channel.
func worker(
	workerCtx context.Context,
	id int,
	tasks <-chan []string,
	results chan<- []TelemetryResult,
	ctx sdk.Context,
	erc20Keeper erc20keeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				fmt.Printf("Worker %d stopping due to channel closed\n", id)
				return nil // Channel closed, stop the worker
			}

			processResults, err := performTask(task, id, ctx, erc20Keeper, nativeTokenPairs)
			if err != nil {
				fmt.Println("Error received: ", err)
				return err
			}
			results <- processResults
		case <-workerCtx.Done():
			// Context cancelled, stop the worker
			fmt.Printf("Worker %d stopping due to cancellation\n", id)
			return nil
		}
	}
}

func performTask(task []string, id int,
	ctx sdk.Context, erc20Keeper erc20keeper.Keeper, tokenPairs []erc20types.TokenPair,
) ([]TelemetryResult, error) {
	results := []TelemetryResult{}
	i := 0
	for _, account := range task {
		for tokenID, pair := range tokenPairs {
			// if account == "evmos1qqqvgqnylf3qtts70jmjxtn668w9gu4yhvz2ms" {
			// 	continue
			// }
			cosmosAddress := sdk.MustAccAddressFromBech32(account)
			ethAddress := common.BytesToAddress(cosmosAddress.Bytes())
			balance := erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), ethAddress)
			if balance == nil {
				return nil, fmt.Errorf("failed to get ERC20 balance (contract %q) for %s", pair.GetERC20Contract(), account)
			}

			if balance.Sign() > 0 {
				results = append(results, TelemetryResult{address: account, balance: balance.String(), id: tokenID})
				i++
			}
		}

	}
	return results, nil
}

func orchestrator(workerCtx context.Context, tasks chan<- []string, accountKeeper authkeeper.AccountKeeper, batchSize int,
	ctx sdk.Context,
) {
	var currentBatch []string
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
			tasks <- currentBatch
			currentBatch = nil // Reset current batch
		}
		i++
		return false
	})
	tasks <- currentBatch
}

func processResults(results <-chan []TelemetryResult) []TelemetryResult {
	finalizedResults := make([]TelemetryResult, 0)
	for batchResults := range results {
		for i := range batchResults {
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

	numWorkers := 100
	batchSize := 1000

	g := new(errgroup.Group)
	// Create a context to cancel the workers in case of an error
	g, workerCtx := errgroup.WithContext(context.Background())

	// Create buffered channels for tasks and results
	tasks := make(chan []string, numWorkers)
	results := make(chan []TelemetryResult, numWorkers)

	// Fan-out: Create worker goroutines
	for w := 1; w <= numWorkers; w++ {
		func(w int) {
			g.Go(func() error {
				return worker(workerCtx, w, tasks, results, ctx, erc20Keeper, wrappedAddr, nativeTokenPairs)
			})
		}(w)
	}

	// Create a goroutine to send tasks to workers
	go func() {
		orchestrator(workerCtx, tasks, accountKeeper, batchSize, ctx)
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
	finalizedResults := processResults(results)
	if g.Wait() != nil {
		err := g.Wait()
		fmt.Println("Context is cancelled we are destroying everything")
		fmt.Println(err)
		return err
	} else {
		fmt.Println("Completed Finalized results: ", len(finalizedResults))
	}

	executeConversionBatch(ctx, logger, finalizedResults, bankKeeper, erc20Keeper, wrappedAddr, nativeTokenPairs)

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
	tokenPair erc20types.TokenPair,
) error {
	balance := erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, contract, from)
	if balance == nil {
		return fmt.Errorf("failed to get ERC20 balance (contract %q) for %s", contract.String(), from.String())
	}

	if balance.Sign() <= 0 {
		return nil
	}

	msg := erc20types.NewMsgConvertERC20(sdk.NewIntFromBigInt(balance), receiver, contract, from)
	_, err := erc20Keeper.ConvertSTRV2(ctx, tokenPair, msg, receiver, from, balance)

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
