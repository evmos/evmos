// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"time"

	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/mock"

	sdkmath "cosmossdk.io/math"
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmtypes "github.com/cometbft/cometbft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/evmos/evmos/v14/x/epochs/types"
	infltypes "github.com/evmos/evmos/v14/x/inflation/types"
)

// createValidatorSet creates validator set with the amount of validators specified
// with the default power of 1.
func createValidatorSet(numberOfValidators int) *tmtypes.ValidatorSet {
	// Create validator set
	tmValidators := make([]*tmtypes.Validator, 0, numberOfValidators)
	for i := 0; i < numberOfValidators; i++ {
		privVal := mock.NewPV()
		pubKey, _ := privVal.GetPubKey()
		validator := tmtypes.NewValidator(pubKey, 1)
		tmValidators = append(tmValidators, validator)
	}

	return tmtypes.NewValidatorSet(tmValidators)
}

// createGenesisAccounts returns a slice of genesis accounts from the given
// account addresses.
func createGenesisAccounts(accounts []sdktypes.AccAddress) []authtypes.GenesisAccount {
	numberOfAccounts := len(accounts)
	genAccounts := make([]authtypes.GenesisAccount, 0, numberOfAccounts)
	for _, acc := range accounts {
		baseAcc := authtypes.NewBaseAccount(acc, nil, 0, 0)
		genAccounts = append(genAccounts, baseAcc)
	}
	return genAccounts
}

func createBalances(accounts []sdktypes.AccAddress, coin sdktypes.Coin) []banktypes.Balance {
	numberOfAccounts := len(accounts)
	fundedAccountBalances := make([]banktypes.Balance, 0, numberOfAccounts)
	for _, acc := range accounts {
		balance := banktypes.Balance{
			Address: acc.String(),
			Coins:   sdktypes.NewCoins(coin),
		}

		fundedAccountBalances = append(fundedAccountBalances, balance)
	}
	return fundedAccountBalances
}

func createEvmosApp(chainID string) *app.Evmos {
	// Create evmos app
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	loadLatest := true
	skipUpgradeHeights := map[int64]bool{}
	homePath := app.DefaultNodeHome
	invCheckPeriod := uint(5)
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	appOptions := simutils.NewAppOptionsWithFlagHome(app.DefaultNodeHome)
	baseAppOptions := []func(*baseapp.BaseApp){baseapp.SetChainID(chainID)}

	return app.NewEvmos(
		logger,
		db,
		nil,
		loadLatest,
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		encodingConfig,
		appOptions,
		baseAppOptions...,
	)
}

func createStakingValidator(val *tmtypes.Validator, amtOfBondedTokens sdkmath.Int) (stakingtypes.Validator, error) {
	pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	commission := stakingtypes.NewCommission(sdktypes.ZeroDec(), sdktypes.ZeroDec(), sdktypes.ZeroDec())
	validator := stakingtypes.Validator{
		OperatorAddress:   sdktypes.ValAddress(val.Address).String(),
		ConsensusPubkey:   pkAny,
		Jailed:            false,
		Status:            stakingtypes.Bonded,
		Tokens:            amtOfBondedTokens,
		DelegatorShares:   sdktypes.OneDec(),
		Description:       stakingtypes.Description{},
		UnbondingHeight:   int64(0),
		UnbondingTime:     time.Unix(0, 0).UTC(),
		Commission:        commission,
		MinSelfDelegation: sdktypes.ZeroInt(),
	}
	return validator, nil
}

func createStakingValidators(tmValidators []*tmtypes.Validator, bondedAmt sdkmath.Int) ([]stakingtypes.Validator, error) {
	amountOfValidators := len(tmValidators)
	stakingValidators := make([]stakingtypes.Validator, 0, amountOfValidators)
	for _, val := range tmValidators {
		validator, err := createStakingValidator(val, bondedAmt)
		if err != nil {
			return nil, err
		}
		stakingValidators = append(stakingValidators, validator)
	}
	return stakingValidators, nil
}

func createDelegations(tmValidators []*tmtypes.Validator, fromAccount sdktypes.AccAddress) []stakingtypes.Delegation {
	amountOfValidators := len(tmValidators)
	delegations := make([]stakingtypes.Delegation, 0, amountOfValidators)
	for _, val := range tmValidators {
		delegation := stakingtypes.NewDelegation(fromAccount, val.Address.Bytes(), sdktypes.OneDec())
		delegations = append(delegations, delegation)
	}
	return delegations
}

type StakingCustomGenesisState struct {
	denom string

	validators  []stakingtypes.Validator
	delegations []stakingtypes.Delegation
}

func setStakingGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState, overwriteParams StakingCustomGenesisState) simapp.GenesisState {
	// Set staking params
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = overwriteParams.denom

	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, overwriteParams.validators, overwriteParams.delegations)
	genesisState[stakingtypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(stakingGenesis)
	return genesisState
}

func setAuthGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState, genAccounts []authtypes.GenesisAccount) simapp.GenesisState {
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccounts)
	genesisState[authtypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(authGenesis)
	return genesisState
}

func setInflationGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState) simapp.GenesisState {
	inflationParams := infltypes.DefaultParams()
	inflationParams.EnableInflation = false

	inflationGenesis := infltypes.NewGenesisState(inflationParams, uint64(0), epochstypes.DayEpochID, 365, 0)
	genesisState[infltypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(&inflationGenesis)
	return genesisState
}

type BankCustomGenesisState struct {
	totalSupply sdktypes.Coins
	balances    []banktypes.Balance
}

func setBankGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState, overwriteParams BankCustomGenesisState) simapp.GenesisState {
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		overwriteParams.balances,
		overwriteParams.totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(bankGenesis)
	return genesisState
}

func calculateTotalSupply(fundedAccountsBalances []banktypes.Balance) sdktypes.Coins {
	totalSupply := sdktypes.NewCoins()
	for _, balance := range fundedAccountsBalances {
		totalSupply = totalSupply.Add(balance.Coins...)
	}
	return totalSupply
}

func addBondedModuleAccountToFundedBalances(fundedAccountsBalances []banktypes.Balance, totalBonded sdktypes.Coin) []banktypes.Balance {
	// add bonded amount to bonded pool module account and include it on funded accounts
	return append(fundedAccountsBalances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdktypes.Coins{totalBonded},
	})
}
