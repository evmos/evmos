// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/precompiles/vesting"
	evmosutil "github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
=======
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmosapp "github.com/evmos/evmos/v19/app"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/precompiles/testutil/contracts"
	"github.com/evmos/evmos/v19/precompiles/vesting"
	"github.com/evmos/evmos/v19/precompiles/vesting/testdata"
	evmosutil "github.com/evmos/evmos/v19/testutil"
	testutiltx "github.com/evmos/evmos/v19/testutil/tx"
	evmostypes "github.com/evmos/evmos/v19/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v19/x/inflation/v1/types"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
>>>>>>> main

	"github.com/evmos/evmos/v19/precompiles/authorization"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

<<<<<<< HEAD
=======
// SetupWithGenesisValSet initializes a new EvmosApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func (s *PrecompileTestSuite) SetupWithGenesisValSet(valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) {
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
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			MinSelfDelegation: math.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), math.LegacyOneDec()))
	}
	s.validators = validators

	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	// set bond demon to be aevmos
	stakingParams.BondDenom = utils.BaseDenom
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalBondAmt := math.ZeroInt()
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

func (s *PrecompileTestSuite) DoSetupTest() {
	nValidators := 3
	signers := make(map[string]tmtypes.PrivValidator, nValidators)
	validators := make([]*tmtypes.Validator, 0, nValidators)

	for i := 0; i < nValidators; i++ {
		privVal := mock.NewPV()
		pubKey, err := privVal.GetPubKey()
		s.Require().NoError(err)
		signers[pubKey.Address().String()] = privVal
		validator := tmtypes.NewValidator(pubKey, 1)
		validators = append(validators, validator)
	}

	valSet := tmtypes.NewValidatorSet(validators)

	// generate genesis account
	addr, priv := testutiltx.NewAddrKey()
	s.privKey = priv
	s.address = addr
	s.signer = testutiltx.NewSigner(priv)

	baseAcc := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)

	amount := sdk.TokensFromConsensusPower(5, evmostypes.PowerReduction)

	balance := banktypes.Balance{
		Address: baseAcc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount)),
	}

	s.SetupWithGenesisValSet(valSet, []authtypes.GenesisAccount{baseAcc}, balance)

	// Create StateDB
	s.stateDB = statedb.New(s.ctx, s.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(s.ctx.HeaderHash().Bytes())))

	// bond denom
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	s.bondDenom = stakingParams.BondDenom
	err = s.app.StakingKeeper.SetParams(s.ctx, stakingParams)
	s.Require().NoError(err, "failed to set params")

	s.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	precompile, err := vesting.NewPrecompile(s.app.VestingKeeper, s.app.AuthzKeeper)
	s.Require().NoError(err)
	s.precompile = precompile

	coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(5000000000000000000)))
	distrCoins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(2000000000000000000)))
	err = s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, authtypes.FeeCollectorName, distrCoins)
	s.Require().NoError(err)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, s.app.EvmKeeper)
	s.queryClientEVM = evmtypes.NewQueryClient(queryHelperEvm)

	vestingCallerContract, err := testdata.LoadVestingCallerContract()
	s.Require().NoError(err)
	s.vestingCallerContract = vestingCallerContract
}

>>>>>>> main
// CallType is a struct that represents the type of call to be made to the
// precompile - either direct or through a smart contract.
type CallType struct {
	// name is the name of the call type
	name string
	// directCall is true if the call is to be made directly to the precompile
	directCall bool
}

// BuildCallArgs builds the call arguments for the integration test suite
// depending on the type of interaction. `contractAddr` is used as the `to`
// of the transaction only if the `callType` is direct, Otherwise, this
// field is ignored and the `to` is the vesting precompile.
// FIX: should be renamed
func (s *PrecompileTestSuite) BuildCallArgs(
	callType CallType,
	contractAddr common.Address,
) (factory.CallArgs, evmtypes.EvmTxArgs) {
	var (
		to       common.Address
		callArgs = factory.CallArgs{}
		txArgs   = evmtypes.EvmTxArgs{}
	)

	if callType.directCall {
		callArgs.ContractABI = s.precompile.ABI
		to = s.precompile.Address()
	} else {
		callArgs.ContractABI = vestingCaller.ABI
		to = contractAddr
	}
	txArgs.To = &to
	return callArgs, txArgs
}

// FundTestClawbackVestingAccount funds the clawback vesting account with some tokens
func (s *PrecompileTestSuite) FundTestClawbackVestingAccount() {
	method := s.precompile.Methods[vesting.FundVestingAccountMethod]
<<<<<<< HEAD
	createArgs := []interface{}{s.keyring.GetAddr(0), toAddr, uint64(time.Now().Unix()), lockupPeriods, vestingPeriods}
	msg, _, _, _, _, err := vesting.NewMsgFundVestingAccount(createArgs, &method) //nolint:dogsled
=======
	createArgs := []interface{}{s.address, toAddr, uint64(time.Now().Unix()), lockupPeriods, vestingPeriods}
	msg, _, _, _, _, err := vesting.NewMsgFundVestingAccount(createArgs, &method) //nolint:dogsled
	s.Require().NoError(err)
	_, err = s.app.VestingKeeper.FundVestingAccount(s.ctx, msg)
>>>>>>> main
	s.Require().NoError(err)
	_, err = s.network.App.VestingKeeper.FundVestingAccount(s.network.GetContext(), msg)
	s.Require().NoError(err)
	vestingAcc, err := s.network.App.VestingKeeper.Balances(s.network.GetContext(), &vestingtypes.QueryBalancesRequest{Address: sdk.AccAddress(toAddr.Bytes()).String()})
	s.Require().NoError(err)
	s.Require().Equal(vestingAcc.Locked, balancesSdkCoins)
	s.Require().Equal(vestingAcc.Unvested, balancesSdkCoins)
}

// CreateTestClawbackVestingAccount creates a vesting account that can clawback.
// Useful for unit tests only
func (s *PrecompileTestSuite) CreateTestClawbackVestingAccount(ctx sdk.Context, funder, vestingAddr common.Address) {
	msgArgs := []interface{}{funder, vestingAddr, false}
	msg, _, _, err := vesting.NewMsgCreateClawbackVestingAccount(msgArgs)
<<<<<<< HEAD
=======
	s.Require().NoError(err)
	err = evmosutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
>>>>>>> main
	s.Require().NoError(err)
	err = evmosutil.FundAccount(ctx, s.network.App.BankKeeper, vestingAddr.Bytes(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(100))))
	s.Require().NoError(err)
	_, err = s.network.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
	s.Require().NoError(err)
}

// ExpectSimpleVestingAccount checks that the vesting account has the expected funder address.
func (s *PrecompileTestSuite) ExpectSimpleVestingAccount(vestingAddr, funderAddr common.Address) {
	vestingAcc := s.GetVestingAccount(vestingAddr)
	funder, err := sdk.AccAddressFromBech32(vestingAcc.FunderAddress)
	Expect(err).ToNot(HaveOccurred(), "vesting account should have a valid funder address")
	Expect(funder.Bytes()).To(Equal(funderAddr.Bytes()), "vesting account should have the correct funder address")
}

// ExpectVestingAccount checks that the vesting account has the expected lockup and vesting periods.
func (s *PrecompileTestSuite) ExpectVestingAccount(vestingAddr common.Address, lockupPeriods, vestingPeriods []vesting.Period) {
	vestingAcc := s.GetVestingAccount(vestingAddr)
	// TODO: check for multiple lockup or vesting periods
	Expect(vestingAcc.LockupPeriods).To(HaveLen(len(lockupPeriods)), "vesting account should have the correct number of lockup periods")
	Expect(vestingAcc.LockupPeriods[0].Length).To(Equal(lockupPeriods[0].Length), "vesting account should have the correct lockup period length")
	Expect(vestingAcc.LockupPeriods[0].Amount[0].Denom).To(Equal(lockupPeriods[0].Amount[0].Denom), "vesting account should have the correct vestingPeriod amount")
	Expect(vestingAcc.LockupPeriods[0].Amount[0].Amount.BigInt()).To(Equal(lockupPeriods[0].Amount[0].Amount), "vesting account should have the correct vesting amount")
	Expect(vestingAcc.VestingPeriods).To(HaveLen(len(vestingPeriods)), "vesting account should have the correct number of vesting periods")
	Expect(vestingAcc.VestingPeriods[0].Length).To(Equal(vestingPeriods[0].Length), "vesting account should have the correct vesting period length")
	Expect(vestingAcc.VestingPeriods[0].Amount[0].Denom).To(Equal(vestingPeriods[0].Amount[0].Denom), "vesting account should have the correct vesting amount")
	Expect(vestingAcc.VestingPeriods[0].Amount[0].Amount.BigInt()).To(Equal(vestingPeriods[0].Amount[0].Amount), "vesting account should have the correct vesting amount")
}

// ExpectVestingFunder checks that the vesting funder of a given vesting account address is the given one.
func (s *PrecompileTestSuite) ExpectVestingFunder(vestingAddr common.Address, funderAddr common.Address) {
	vestingAcc := s.GetVestingAccount(vestingAddr)
	Expect(vestingAcc.FunderAddress).To(Equal(sdk.AccAddress(funderAddr.Bytes()).String()), "expected a different funder for the vesting account")
}

// GetVestingAccount returns the vesting account for the given address.
func (s *PrecompileTestSuite) GetVestingAccount(addr common.Address) *vestingtypes.ClawbackVestingAccount {
	acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), addr.Bytes())
	Expect(acc).ToNot(BeNil(), "vesting account should exist")
	vestingAcc, ok := acc.(*vestingtypes.ClawbackVestingAccount)
	Expect(ok).To(BeTrue(), "vesting account should be of type VestingAccount")
	return vestingAcc
}

// CreateFundVestingAuthorization creates an approval authorization for the grantee to use granter's balance
// to send the specified message. The method check that this is the only authorization stored for the pair
// (granter, grantee) and returns an error if this is not true.
func (s *PrecompileTestSuite) CreateVestingMsgAuthorization(granter keyring.Key, grantee common.Address, msg string) {
	approvalCallArgs := factory.CallArgs{
		ContractABI: s.precompile.ABI,
		MethodName:  "approve",
		Args: []interface{}{
			grantee,
			msg,
		},
	}

	precompileAddr := s.precompile.Address()
	logCheck := passCheck.WithExpEvents(authorization.EventTypeApproval)

	_, _, err = s.factory.CallContractAndCheckLogs(granter.Priv, evmtypes.EvmTxArgs{To: &precompileAddr}, approvalCallArgs, logCheck)
	Expect(err).To(BeNil(), "error while creating the generic authorization: %v", err)
	Expect(s.network.NextBlock()).To(BeNil())

	auths, err := s.grpcHandler.GetAuthorizations(sdk.AccAddress(grantee.Bytes()).String(), granter.AccAddr.String())
	Expect(err).To(BeNil())
	Expect(auths).To(HaveLen(1))
}

// GetBondBalances returns the balances of the bonded denom for the given addresses. The
// testing suite checks for error to be nil during the queries.
func (s *PrecompileTestSuite) GetBondBalances(addresses ...sdk.AccAddress) []math.Int {
	balances := make([]math.Int, 0, len(addresses))

	for _, acc := range addresses {
		balResp, err := s.grpcHandler.GetBalance(acc, s.bondDenom)
		Expect(err).To(BeNil())
		balances = append(balances, balResp.Balance.Amount)
	}

	return balances
}

// mergeEventMaps is a helper function to merge events maps from different contracts.
// If duplicates events are present, map2 override map1 values.
func mergeEventMaps(map1, map2 map[string]abi.Event) map[string]abi.Event {
	// Create a new map to hold the merged result
	mergedMap := make(map[string]abi.Event)

	// Copy all key-value pairs from map1 to mergedMap
	for k, v := range map1 {
		mergedMap[k] = v
	}

	// Copy all key-value pairs from map2 to mergedMap
	// If there are duplicate keys, values from map2 will overwrite those from map1
	for k, v := range map2 {
		mergedMap[k] = v
	}

	return mergedMap
}

// mergeEventMaps is a helper function to merge events maps from different contracts.
// If duplicates events are present, map2 override map1 values.
func mergeEventMaps(map1, map2 map[string]abi.Event) map[string]abi.Event {
	// Create a new map to hold the merged result
	mergedMap := make(map[string]abi.Event)

	// Copy all key-value pairs from map1 to mergedMap
	for k, v := range map1 {
		mergedMap[k] = v
	}

	// Copy all key-value pairs from map2 to mergedMap
	// If there are duplicate keys, values from map2 will overwrite those from map1
	for k, v := range map2 {
		mergedMap[k] = v
	}

	return mergedMap
}
