// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	"fmt"
	"math/big"

	"context"
	"sync"

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
)

// worker performs the task on jobs received and sends results to the results channel.
func worker(
	// worker
	workerCtx context.Context,
	id int,
	tasks <-chan []authtypes.AccountI,
	errs chan<- error,
	wg *sync.WaitGroup,

	ctx sdk.Context,
	logger log.Logger,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				fmt.Printf("Worker %d stopping due to channel closed\n", id)
				return // Channel closed, stop the worker
			}

			for i := range task {
				executeConversionBatch(ctx, logger, task, bankKeeper, erc20Keeper, wrappedAddr, nativeTokenPairs)
				fmt.Printf("Worker %d processing address: %s\n", id, task[i])
			}

		case <-workerCtx.Done():
			// Context cancelled, stop the worker
			fmt.Printf("Worker %d stopping due to cancellation\n", id)
			return
		}
	}
}
func executeConversionBatch(
	ctx sdk.Context,
	logger log.Logger,
	accounts []authtypes.AccountI,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) {
	for _, account := range accounts {
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
			if err := ConvertERC20Token(ctx, ethAddress, contract, cosmosAddress, erc20Keeper, tokenPair); err != nil {
				logger.Error(
					"failed to convert ERC20 to native Coin",
					"account", ethHexAddr,
					"erc20", contract.String(),
					"balance", balance.String(),
					"error", err.Error(),
				)
			}
		}
	}
}

type Task struct {
	accounts []authtypes.AccountI
	ctx      sdk.Context
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

	numWorkers := 4
	batchSize := 5

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the context is cancelled to prevent a context leak

	var wg sync.WaitGroup

	// Create buffered channels for tasks and results
	tasks := make(chan []authtypes.AccountI, numWorkers)
	// Results channel is buffered to ensure non-blocking send
	// This results will be used for telemetry
	// results := make(chan int, numWorkers)
	errs := make(chan error) // Buffered channel to ensure non-blocking send on error

	// Fan-out: Create worker goroutines
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(workerCtx, w, tasks, errs, &wg,
			ctx, logger, bankKeeper, erc20Keeper, wrappedAddr, nativeTokenPairs)
	}

	go func() {
		for err := range errs {
			if err != nil {
				fmt.Println("Received error from worker:", err)
				// Cancel all goroutines upon error
				cancel()
				break // Stop listening for more errors
			}
		}
	}()

	var currentBatch []authtypes.AccountI
	i := 0
	// iterate over all the accounts and convert the tokens to native coins
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {

		if workerCtx.Err() != nil {
			// If the context is already cancelled, stop sending tasks
			return true
		}
		currentBatch = append(currentBatch, account)
		i++
		// Check if the current batch is filled or it's the last element.
		if (i+1)%batchSize == 0 {
			tasks <- currentBatch
			currentBatch = nil // Reset current batch
		}

		return false
	})
	// Process the remaining batch
	tasks <- currentBatch

	close(tasks) // Close the jobs channel to signal no more jobs will be sent

	// Ensure all goroutines have finished after cancellation
	wg.Wait()
	close(errs)
	fmt.Println("All workers have been cancelled.")

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
