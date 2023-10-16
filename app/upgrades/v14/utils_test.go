// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v14_test

import (
	"encoding/json"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/v7/testing/mock"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmosapp "github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/vesting"
	evmosutil "github.com/evmos/evmos/v15/testutil"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
	"github.com/evmos/evmos/v15/utils"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/types"
)

// SetupWithGenesisValSet initializes a new EvmosApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func (s *UpgradesTestSuite) SetupWithGenesisValSet(valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) {
	appI, genesisState := evmosapp.SetupTestingApp(cmn.DefaultChainID)()
	app, ok := appI.(*evmosapp.Evmos)
	s.Require().True(ok)

	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.TokensFromConsensusPower(1, evmostypes.PowerReduction)

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		s.Require().NoError(err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		s.Require().NoError(err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))
	}
	s.validators = validators

	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	// set bond demon to be aevmos
	stakingParams.BondDenom = utils.BaseDenom
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalBondAmt := sdk.ZeroInt()
	for range validators {
		totalBondAmt = totalBondAmt.Add(bondAmt)
	}
	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(utils.BaseDenom, totalBondAmt))...)
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(utils.BaseDenom, totalBondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	s.Require().NoError(err)

	header := evmosutil.NewHeader(
		2,
		time.Now().UTC(),
		cmn.DefaultChainID,
		sdk.ConsAddress(validators[0].GetOperator()),
		tmhash.Sum([]byte("app")),
		tmhash.Sum([]byte("validators")),
	)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			ChainId:         cmn.DefaultChainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: evmosapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// create Context
	s.ctx = app.BaseApp.NewContext(false, header)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	s.app = app
}

func (s *UpgradesTestSuite) DoSetupTest() {
	nValidators := 3
	validators := make([]*tmtypes.Validator, 0, nValidators)

	for i := 0; i < nValidators; i++ {
		privVal := mock.NewPV()
		pubKey, err := privVal.GetPubKey()
		s.Require().NoError(err)
		validator := tmtypes.NewValidator(pubKey, 1)
		validators = append(validators, validator)
	}

	valSet := tmtypes.NewValidatorSet(validators)

	// generate genesis account
	addr, priv := testutiltx.NewAddrKey()
	s.address = addr

	baseAcc := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)

	acc := &evmostypes.EthAccount{
		BaseAccount: baseAcc,
		CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
	}

	amount := sdk.TokensFromConsensusPower(5, evmostypes.PowerReduction)

	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount)),
	}

	s.SetupWithGenesisValSet(valSet, []authtypes.GenesisAccount{acc}, balance)

	// Create StateDB
	s.stateDB = statedb.New(s.ctx, s.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(s.ctx.HeaderHash().Bytes())))

	// bond denom
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	stakingParams.MinCommissionRate = sdk.ZeroDec()
	s.bondDenom = stakingParams.BondDenom
	err := s.app.StakingKeeper.SetParams(s.ctx, stakingParams)
	s.Require().NoError(err, "failed to set params")

	s.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	precompile, err := vesting.NewPrecompile(s.app.VestingKeeper, s.app.AuthzKeeper)
	s.Require().NoError(err)
	s.precompile = precompile

	coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(5000000000000000000)))
	distrCoins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(2000000000000000000)))
	err = s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, authtypes.FeeCollectorName, distrCoins)
	s.Require().NoError(err)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, s.app.EvmKeeper)
	s.queryClientEVM = evmtypes.NewQueryClient(queryHelperEvm)

	s.NextBlock()
}

// NextBlock commits the current block and sets up the next block.
func (s *UpgradesTestSuite) NextBlock() {
	var err error
	s.ctx, err = evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Second, nil)
	s.Require().NoError(err)
}

// MigrationTestAccount is a struct to hold the test account address, its private key
// as well as its balances and delegations before and after the migration.
type MigrationTestAccount struct {
	Addr            sdk.AccAddress
	PrivKey         *ethsecp256k1.PrivKey
	BalancePre      sdk.Coins
	BalancePost     sdk.Coins
	DelegationsPost stakingtypes.Delegations
}

// GenerateMigrationTestAccount returns a new MigrationTestAccount with a new address and private key.
// All expected pre- and post-migration balances and delegations are initialized to be empty.
func GenerateMigrationTestAccount() MigrationTestAccount {
	addr, priv := testutiltx.NewAccAddressAndKey()
	return MigrationTestAccount{
		Addr:            addr,
		PrivKey:         priv,
		BalancePre:      sdk.Coins{},
		BalancePost:     sdk.Coins{},
		DelegationsPost: stakingtypes.Delegations{},
	}
}

// requireMigratedAccount checks that the account has the expected balances and delegations
// after the migration.
func (s *UpgradesTestSuite) requireMigratedAccount(account MigrationTestAccount) {
	// Check the new delegations
	delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, account.Addr)
	s.Require().Len(delegations, len(account.DelegationsPost), "expected different number of delegations after migration of account %s", account.Addr.String())
	s.Require().ElementsMatch(delegations, account.DelegationsPost, "expected different delegations after migration of account %s", account.Addr.String())

	// There should not be any unbonding delegations
	unbondingDelegations := s.app.StakingKeeper.GetAllUnbondingDelegations(s.ctx, account.Addr)
	s.Require().Len(unbondingDelegations, 0, "expected no unbonding delegations after migration for account %s", account.Addr.String())

	// Check balances
	balances := s.app.BankKeeper.GetAllBalances(s.ctx, account.Addr)
	s.Require().Len(balances, len(account.BalancePost), "expected different number of balances after migration for account %s", account.Addr.String())

	for _, balance := range balances {
		expAmount := account.BalancePost.AmountOf(balance.Denom)
		s.Require().Equal(expAmount.String(), balance.Amount.String(), "expected different balance of %q after migration for account %s", balance.Denom, account.Addr.String())
	}
}

// getDelegationSharesMap returns a map of validator operator addresses to the
// total shares delegated to them.
func (s *UpgradesTestSuite) getDelegationSharesMap() map[string]sdk.Dec {
	allValidators := s.app.StakingKeeper.GetAllValidators(s.ctx)
	sharesMap := make(map[string]sdk.Dec, len(allValidators))
	for _, validator := range allValidators {
		sharesMap[validator.OperatorAddress] = validator.DelegatorShares
	}
	return sharesMap
}

// CreateDelegationWithZeroTokens is a helper script, which creates a delegation and
// slashes the validator afterwards, so that the shares are not zero but the query
// for token amount returns zero tokens.
//
// NOTE: This is replicating an edge case that was found in mainnet data, which led to
// the account migrations not succeeding.
func CreateDelegationWithZeroTokens(
	ctx sdk.Context,
	app *evmosapp.Evmos,
	priv *ethsecp256k1.PrivKey,
	delegator sdk.AccAddress,
	validator stakingtypes.Validator,
	amount int64,
) (stakingtypes.Delegation, error) {
	delegation, err := Delegate(ctx, app, priv, delegator, validator, amount)
	if err != nil {
		return stakingtypes.Delegation{}, err
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return stakingtypes.Delegation{}, fmt.Errorf("failed to get validator consensus address: %w", err)
	}

	// Slash the validator
	app.SlashingKeeper.Slash(s.ctx, consAddr, sdk.NewDecWithPrec(5, 2), 1, 0)

	return delegation, nil
}

// Delegate is a helper function to delegation one atto-Evmos to a validator.
func Delegate(
	ctx sdk.Context,
	app *evmosapp.Evmos,
	priv *ethsecp256k1.PrivKey,
	delegator sdk.AccAddress,
	validator stakingtypes.Validator,
	amount int64,
) (stakingtypes.Delegation, error) {
	stakingDenom := app.StakingKeeper.BondDenom(ctx)

	msgDelegate := stakingtypes.NewMsgDelegate(delegator, validator.GetOperator(), sdk.NewInt64Coin(stakingDenom, amount))
	_, err := evmosutil.DeliverTx(ctx, app, priv, nil, msgDelegate)
	if err != nil {
		return stakingtypes.Delegation{}, fmt.Errorf("failed to delegate: %w", err)
	}

	delegation, found := app.StakingKeeper.GetDelegation(s.ctx, delegator, validator.GetOperator())
	if !found {
		return stakingtypes.Delegation{}, fmt.Errorf("delegation not found")
	}

	return delegation, nil
}
