// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"fmt"
	"slices"
	"time"

	"github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/encoding"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v16/types"
	evmosutil "github.com/evmos/evmos/v16/utils"
	epochstypes "github.com/evmos/evmos/v16/x/epochs/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
)

// genSetupFn is the type for the module genesis setup functions
type genSetupFn func(evmosApp *app.Evmos, genesisState types.GenesisState, customGenesis interface{}) (types.GenesisState, error)

// defaultGenesisParams contains the params that are needed to
// setup the default genesis for the testing setup
type defaultGenesisParams struct {
	genAccounts []authtypes.GenesisAccount
	staking     StakingCustomGenesisState
	bank        BankCustomGenesisState
}

// genesisSetupFunctions contains the available genesis setup functions
// that can be used to customize the network genesis
var genesisSetupFunctions = map[string]genSetupFn{
	evmtypes.ModuleName:       genStateSetter[*evmtypes.GenesisState](evmtypes.ModuleName),
	govtypes.ModuleName:       genStateSetter[*govtypesv1.GenesisState](govtypes.ModuleName),
	infltypes.ModuleName:      genStateSetter[*infltypes.GenesisState](infltypes.ModuleName),
	feemarkettypes.ModuleName: genStateSetter[*feemarkettypes.GenesisState](feemarkettypes.ModuleName),
	banktypes.ModuleName:      setBankGenesisState,
}

// genStateSetter is a generic function to set module-specific genesis state
func genStateSetter[T proto.Message](moduleName string) genSetupFn {
	return func(evmosApp *app.Evmos, genesisState types.GenesisState, customGenesis interface{}) (types.GenesisState, error) {
		moduleGenesis, ok := customGenesis.(T)
		if !ok {
			return nil, fmt.Errorf("invalid type %T for %s module genesis state", customGenesis, moduleName)
		}

		genesisState[moduleName] = evmosApp.AppCodec().MustMarshalJSON(moduleGenesis)
		return genesisState, nil
	}
}

// createValidatorSetAndSigners creates validator set with the amount of validators specified
// with the default power of 1.
func createValidatorSetAndSigners(numberOfValidators int) (*cmttypes.ValidatorSet, map[string]cmttypes.PrivValidator) {
	// Create validator set
	tmValidators := make([]*cmttypes.Validator, 0, numberOfValidators)
	signers := make(map[string]cmttypes.PrivValidator, numberOfValidators)

	for i := 0; i < numberOfValidators; i++ {
		privVal := mock.NewPV()
		pubKey, _ := privVal.GetPubKey()
		validator := cmttypes.NewValidator(pubKey, 1)
		tmValidators = append(tmValidators, validator)
		signers[pubKey.Address().String()] = privVal
	}

	return cmttypes.NewValidatorSet(tmValidators), signers
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
		genAccounts = append(genAccounts, sdktypes.AccAddress(balance.Address))
	}
	return genAccounts
}

// createBalances creates balances for the given accounts and coin
func createBalances(accounts []sdktypes.AccAddress, denoms []string) []banktypes.Balance {
	slices.Sort(denoms)
	numberOfAccounts := len(accounts)
	coins := make([]sdktypes.Coin, len(denoms))
	for i, denom := range denoms {
		coins[i] = sdktypes.NewCoin(denom, PrefundedAccountInitialBalance)
	}
	fundedAccountBalances := make([]banktypes.Balance, 0, numberOfAccounts)
	for _, acc := range accounts {
		balance := banktypes.Balance{
			Address: acc.String(),
			Coins:   coins,
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
func createStakingValidator(val *cmttypes.Validator, bondedAmt sdkmath.Int) (stakingtypes.Validator, error) {
	pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	commission := stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec())
	validator := stakingtypes.Validator{
		OperatorAddress:   sdktypes.ValAddress(val.Address).String(),
		ConsensusPubkey:   pkAny,
		Jailed:            false,
		Status:            stakingtypes.Bonded,
		Tokens:            bondedAmt,
		DelegatorShares:   sdkmath.LegacyOneDec(),
		Description:       stakingtypes.Description{},
		UnbondingHeight:   int64(0),
		UnbondingTime:     time.Unix(0, 0).UTC(),
		Commission:        commission,
		MinSelfDelegation: sdkmath.ZeroInt(),
	}
	return validator, nil
}

// createStakingValidators creates staking validators from the given tm validators and bonded
// amounts
func createStakingValidators(tmValidators []*cmttypes.Validator, bondedAmt sdkmath.Int) ([]stakingtypes.Validator, error) {
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
func createDelegations(tmValidators []*cmttypes.Validator, fromAccount sdktypes.AccAddress) []stakingtypes.Delegation {
	amountOfValidators := len(tmValidators)
	delegations := make([]stakingtypes.Delegation, 0, amountOfValidators)
	for _, val := range tmValidators {
		delegation := stakingtypes.NewDelegation(fromAccount.String(), sdktypes.ValAddress(val.Address).String(), sdkmath.LegacyOneDec())
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

// setDefaultStakingGenesisState sets the default staking genesis state
func setDefaultStakingGenesisState(evmosApp *app.Evmos, genesisState types.GenesisState, overwriteParams StakingCustomGenesisState) types.GenesisState {
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

// setDefaultInflationGenesisState sets the default inflation genesis state
func setDefaultInflationGenesisState(evmosApp *app.Evmos, genesisState types.GenesisState) types.GenesisState {
	inflationParams := infltypes.DefaultParams()
	inflationParams.EnableInflation = false
	defaultGen := infltypes.NewGenesisState(inflationParams, uint64(0), epochstypes.DayEpochID, 365, 0)

	genesisState[infltypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(&defaultGen)
	return genesisState
}

type BankCustomGenesisState struct {
	totalSupply sdktypes.Coins
	balances    []banktypes.Balance
}

// setDefaultBankGenesisState sets the default bank genesis state
func setDefaultBankGenesisState(evmosApp *app.Evmos, genesisState types.GenesisState, overwriteParams BankCustomGenesisState) types.GenesisState {
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

// setBankGenesisState updates the bank genesis state with custom genesis state
func setBankGenesisState(evmosApp *app.Evmos, genesisState types.GenesisState, customGenesis interface{}) (types.GenesisState, error) {
	customGen, ok := customGenesis.(*banktypes.GenesisState)
	if !ok {
		return nil, fmt.Errorf("invalid type %T for bank module genesis state", customGenesis)
	}

	bankGen := &banktypes.GenesisState{}
	evmosApp.AppCodec().MustUnmarshalJSON(genesisState[banktypes.ModuleName], bankGen)

	if len(customGen.Balances) > 0 {
		coins := sdktypes.NewCoins()
		bankGen.Balances = append(bankGen.Balances, customGen.Balances...)
		for _, b := range customGen.Balances {
			coins = append(coins, b.Coins...)
		}
		bankGen.Supply = bankGen.Supply.Add(coins...)
	}
	if len(customGen.DenomMetadata) > 0 {
		bankGen.DenomMetadata = append(bankGen.DenomMetadata, customGen.DenomMetadata...)
	}

	if len(customGen.SendEnabled) > 0 {
		bankGen.SendEnabled = append(bankGen.SendEnabled, customGen.SendEnabled...)
	}

	bankGen.Params = customGen.Params

	genesisState[banktypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(bankGen)
	return genesisState, nil
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

// setDefaultAuthGenesisState sets the default auth genesis state
func setDefaultAuthGenesisState(evmosApp *app.Evmos, genesisState types.GenesisState, genAccs []authtypes.GenesisAccount) types.GenesisState {
	defaultAuthGen := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = evmosApp.AppCodec().MustMarshalJSON(defaultAuthGen)
	return genesisState
}

// setDefaultGovGenesisState sets the default gov genesis state
func setDefaultGovGenesisState(evmosApp *app.Evmos, genesisState types.GenesisState) types.GenesisState {
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
func newDefaultGenesisState(evmosApp *app.Evmos, params defaultGenesisParams) types.GenesisState {
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
func customizeGenesis(evmosApp *app.Evmos, customGen CustomGenesisState, genesisState types.GenesisState) (types.GenesisState, error) {
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
