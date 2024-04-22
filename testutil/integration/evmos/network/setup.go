// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"fmt"
	"time"

	"github.com/evmos/evmos/v17/app"
	"github.com/evmos/evmos/v17/encoding"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/gogoproto/proto"

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
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/evmos/evmos/v17/x/epochs/types"
	infltypes "github.com/evmos/evmos/v17/x/inflation/v1/types"

	evmosutil "github.com/evmos/evmos/v17/utils"
	evmtypes "github.com/evmos/evmos/v17/x/evm/types"
)

// createValidatorSetAndSigners creates validator set with the amount of validators specified
// with the default power of 1.
func createValidatorSetAndSigners(numberOfValidators int) (*tmtypes.ValidatorSet, map[string]tmtypes.PrivValidator) {
	// Create validator set
	tmValidators := make([]*tmtypes.Validator, 0, numberOfValidators)
	signers := make(map[string]tmtypes.PrivValidator, numberOfValidators)

	for i := 0; i < numberOfValidators; i++ {
		privVal := mock.NewPV()
		pubKey, _ := privVal.GetPubKey()
		validator := tmtypes.NewValidator(pubKey, 1)
		tmValidators = append(tmValidators, validator)
		signers[pubKey.Address().String()] = privVal
	}

	return tmtypes.NewValidatorSet(tmValidators), signers
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

// getAccAddrsFromBalances returns a slice of genesis accounts from the
// given balances.
func getAccAddrsFromBalances(balances []banktypes.Balance) []sdktypes.AccAddress {
	numberOfBalances := len(balances)
	genAccounts := make([]sdktypes.AccAddress, 0, numberOfBalances)
	for _, balance := range balances {
		genAccounts = append(genAccounts, balance.GetAddress())
	}
	return genAccounts
}

// createBalances creates balances for the given accounts and coin
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

// createEvmosApp creates an evmos app
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

// createStakingValidator creates a staking validator from the given tm validator and bonded
func createStakingValidator(val *tmtypes.Validator, bondedAmt sdkmath.Int) (stakingtypes.Validator, error) {
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
		Tokens:            bondedAmt,
		DelegatorShares:   sdktypes.OneDec(),
		Description:       stakingtypes.Description{},
		UnbondingHeight:   int64(0),
		UnbondingTime:     time.Unix(0, 0).UTC(),
		Commission:        commission,
		MinSelfDelegation: sdktypes.ZeroInt(),
	}
	return validator, nil
}

// createStakingValidators creates staking validators from the given tm validators and bonded
// amounts
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

// createDelegations creates delegations for the given validators and account
func createDelegations(tmValidators []*tmtypes.Validator, fromAccount sdktypes.AccAddress) []stakingtypes.Delegation {
	amountOfValidators := len(tmValidators)
	delegations := make([]stakingtypes.Delegation, 0, amountOfValidators)
	for _, val := range tmValidators {
		delegation := stakingtypes.NewDelegation(fromAccount, val.Address.Bytes(), sdktypes.OneDec())
		delegations = append(delegations, delegation)
	}
	return delegations
}

// StakingCustomGenesisState defines the staking genesis state
type StakingCustomGenesisState struct {
	denom string

	validators  []stakingtypes.Validator
	delegations []stakingtypes.Delegation
}

// setDefaultStakingGenesisState sets the staking genesis state
func setDefaultStakingGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState, overwriteParams StakingCustomGenesisState) simapp.GenesisState {
	// Set staking params
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = overwriteParams.denom

	stakingGenesis := stakingtypes.NewGenesisState(
		stakingParams,
		overwriteParams.validators,
		overwriteParams.delegations,
	)
	genesisState[stakingtypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(stakingGenesis)
	return genesisState
}

// setDefaultInflationGenesisState sets the inflation genesis state
func setDefaultInflationGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState) simapp.GenesisState {
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

// setDefaultBankGenesisState sets the bank genesis state
func setDefaultBankGenesisState(
	evmosApp *app.Evmos,
	genesisState simapp.GenesisState,
	overwriteParams BankCustomGenesisState,
) simapp.GenesisState {
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

// genSetupFn is the type for the module genesis setup functions
type genSetupFn func(
	evmosApp *app.Evmos,
	genesisState simapp.GenesisState,
	customGenesis interface{},
) (simapp.GenesisState, error)

// defaultGenesisParams contains the params that are needed to
// setup the default genesis for the testing setup
type defaultGenesisParams struct {
	genAccounts []authtypes.GenesisAccount
	staking     StakingCustomGenesisState
	bank        BankCustomGenesisState
}

// genStateSetter is a generic function to set module-specific genesis state
func genStateSetter[T proto.Message](moduleName string) genSetupFn {
	return func(
		evmosApp *app.Evmos,
		genesisState simapp.GenesisState,
		customGenesis interface{},
	) (simapp.GenesisState, error) {
		moduleGenesis, ok := customGenesis.(T)
		if !ok {
			return nil, fmt.Errorf("invalid type %T for %s module genesis state", customGenesis, moduleName)
		}

		genesisState[moduleName] = evmosApp.AppCodec().MustMarshalJSON(moduleGenesis)
		return genesisState, nil
	}
}

// genesisSetupFunctions contains the available genesis setup functions
// that can be used to customize the network genesis
var genesisSetupFunctions = map[string]genSetupFn{
	evmtypes.ModuleName:  genStateSetter[*evmtypes.GenesisState](evmtypes.ModuleName),
	govtypes.ModuleName:  genStateSetter[*govtypesv1.GenesisState](govtypes.ModuleName),
	infltypes.ModuleName: genStateSetter[*infltypes.GenesisState](infltypes.ModuleName),
}

// setDefaultAuthGenesisState sets the default auth genesis state
func setDefaultAuthGenesisState(
	evmosApp *app.Evmos,
	genesisState simapp.GenesisState,
	genAccs []authtypes.GenesisAccount,
) simapp.GenesisState {
	defaultAuthGen := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(defaultAuthGen)
	return genesisState
}

// setDefaultGovGenesisState sets the default gov genesis state
func setDefaultGovGenesisState(evmosApp *app.Evmos, genesisState simapp.GenesisState) simapp.GenesisState {
	govGen := govtypesv1.DefaultGenesisState()
	updatedParams := govGen.Params
	// set 'aevmos' as deposit denom
	updatedParams.MinDeposit = sdktypes.NewCoins(sdktypes.NewCoin(evmosutil.BaseDenom, sdkmath.NewInt(1e18)))
	govGen.Params = updatedParams
	genesisState[govtypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(govGen)
	return genesisState
}

// defaultAuthGenesisState sets the default genesis state
// for the testing setup
func newDefaultGenesisState(evmosApp *app.Evmos, params defaultGenesisParams) simapp.GenesisState {
	genesisState := app.NewDefaultGenesisState()

	genesisState = setDefaultAuthGenesisState(evmosApp, genesisState, params.genAccounts)
	genesisState = setDefaultStakingGenesisState(evmosApp, genesisState, params.staking)
	genesisState = setDefaultBankGenesisState(evmosApp, genesisState, params.bank)
	genesisState = setDefaultInflationGenesisState(evmosApp, genesisState)
	genesisState = setDefaultGovGenesisState(evmosApp, genesisState)

	return genesisState
}

// customizeGenesis modifies genesis state if there're any custom genesis state
// for specific modules
func customizeGenesis(
	evmosApp *app.Evmos,
	customGen CustomGenesisState,
	genesisState simapp.GenesisState,
) (simapp.GenesisState, error) {
	var err error
	for mod, modGenState := range customGen {
		if fn, found := genesisSetupFunctions[mod]; found {
			genesisState, err = fn(evmosApp, genesisState, modGenState)
			if err != nil {
				return genesisState, err
			}
		}
	}
	return genesisState, err
}

// calculateTotalSupply calculates the total supply from the given balances
func calculateTotalSupply(fundedAccountsBalances []banktypes.Balance) sdktypes.Coins {
	totalSupply := sdktypes.NewCoins()
	for _, balance := range fundedAccountsBalances {
		totalSupply = totalSupply.Add(balance.Coins...)
	}
	return totalSupply
}

// addBondedModuleAccountToFundedBalances adds bonded amount to bonded pool module account and include it on funded accounts
func addBondedModuleAccountToFundedBalances(
	fundedAccountsBalances []banktypes.Balance,
	totalBonded sdktypes.Coin,
) []banktypes.Balance {
	return append(fundedAccountsBalances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdktypes.Coins{totalBonded},
	})
}
