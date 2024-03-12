// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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
)

var storeKey []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}

func executeConversionBatch(
	ctx sdk.Context,
	logger log.Logger,
	results []TelemetryResult2,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	wrappedAddr common.Address,
	nativeTokenPairs []erc20types.TokenPair,
) error {
	totalBalance := big.NewInt(0)
	for _, result := range results {
		ethAddress := common.BytesToAddress(result.address)
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
			balance := new(big.Int).SetBytes(result.balance)
			totalBalance = totalBalance.Add(totalBalance, balance)
			coins := sdk.Coins{sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(balance)}}

			// Unescrow coins and send to receiver
			err := bankKeeper.SendCoinsFromModuleToAccount(ctx, erc20types.ModuleName, result.address, coins)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("Total balance: ", totalBalance.String())
	return nil
}

type parseTokenPairs = []common.Address

type TelemetryResult2 struct {
	address sdk.AccAddress
	balance []byte
	id      int
}

type ExportResult struct {
	Address string
	Balance string
	Erc20   string
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
	timeBegin := time.Now()

	tokenPairs := make(parseTokenPairs, len(nativeTokenPairs))
	for i := range nativeTokenPairs {
		tokenPairs[i] = nativeTokenPairs[i].GetERC20Contract()
	}
	fmt.Println("This is the number of token pairs: ", len(tokenPairs))

	var resultsCol []TelemetryResult2
	var jsonExport []ExportResult = make([]ExportResult, 0)

	tokenPairStores := make([]sdk.KVStore, len(nativeTokenPairs))
	for i, pair := range tokenPairs {
		tokenPairStores[i] = evmKeeper.GetStoreDummy(ctx, pair)
	}

	logCounter := 0
	from := time.Now()
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		accountAddr := account.GetAddress()
		concatBytes := append(common.LeftPadBytes(accountAddr.Bytes(), 32), storeKey...)
		key := crypto.Keccak256Hash(concatBytes)
		for tokenId, store := range tokenPairStores {
			value := store.Get(key.Bytes())
			if len(value) == 0 {
				continue
			}
			resultsCol = append(
				resultsCol,
				TelemetryResult2{
					address: accountAddr,
					balance: value,
					id:      tokenId,
				})

			balance := new(big.Int).SetBytes(value)
			jsonExport = append(
				jsonExport,
				ExportResult{
					Address: accountAddr.String(),
					Balance: balance.String(),
					Erc20:   nativeTokenPairs[tokenId].Erc20Address,
				})
		}
		logCounter++
		if logCounter == 12000 {
			logCounter = 0
			fmt.Println("Accounts with balances: ", len(resultsCol))
			fmt.Printf("Time per batch: %v \n", time.Since(from).String())
			from = time.Now()
		}
		return false
	})

	fmt.Println("Finalized results: ", len(resultsCol))
	err := executeConversionBatch(ctx, logger, resultsCol, bankKeeper, erc20Keeper, wrappedAddr, nativeTokenPairs)
	if err != nil {
		panic(err)
	}

	file, _ := json.MarshalIndent(jsonExport, "", " ")
	_ = os.WriteFile("results.json", file, os.ModePerm)

	// NOTE: if there are tokens left in the ERC-20 module account
	// we return an error because this implies that the migration of native
	// coins to ERC-20 tokens was not fully completed.
	erc20ModuleAccountAddress := authtypes.NewModuleAddress(erc20types.ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, erc20ModuleAccountAddress)
	if !balances.IsZero() {
		return fmt.Errorf("there are still tokens in the erc-20 module account: %s", balances.String())
	}
	duration := time.Since(timeBegin)
	fmt.Println("Duration: ", duration)
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
