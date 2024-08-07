package distribution_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
<<<<<<< HEAD

	"github.com/evmos/evmos/v18/precompiles/staking"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	inflationtypes "github.com/evmos/evmos/v18/x/inflation/v1/types"
)

=======
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmosapp "github.com/evmos/evmos/v19/app"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/precompiles/distribution"
	evmosutil "github.com/evmos/evmos/v19/testutil"
	evmosutiltx "github.com/evmos/evmos/v19/testutil/tx"
	evmostypes "github.com/evmos/evmos/v19/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v19/x/inflation/v1/types"
)

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

	totalBondAmt := bondAmt.Add(bondAmt)
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

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			ChainId:         cmn.DefaultChainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: evmosapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	app.Commit()

	// instantiate new header
	header := evmosutil.NewHeader(
		2,
		time.Now().UTC(),
		cmn.DefaultChainID,
		sdk.ConsAddress(validators[0].GetOperator()),
		tmhash.Sum([]byte("app")),
		tmhash.Sum([]byte("validators")),
	)

	app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// create Context
	s.ctx = app.BaseApp.NewContext(false, header)
	s.app = app
}

func (s *PrecompileTestSuite) DoSetupTest() {
	// generate validator private/public key
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	s.Require().NoError(err)

	privVal2 := mock.NewPV()
	pubKey2, err := privVal2.GetPubKey()
	s.Require().NoError(err)

	// create validator set with two validators
	validator := tmtypes.NewValidator(pubKey, 1)
	validator2 := tmtypes.NewValidator(pubKey2, 2)
	s.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator, validator2})
	signers := make(map[string]tmtypes.PrivValidator)
	signers[pubKey.Address().String()] = privVal
	signers[pubKey2.Address().String()] = privVal2

	// generate genesis account
	addr, priv := evmosutiltx.NewAddrKey()
	s.privKey = priv
	s.address = addr
	s.signer = evmosutiltx.NewSigner(priv)

	baseAcc := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)

	amount := sdk.TokensFromConsensusPower(5, evmostypes.PowerReduction)

	balance := banktypes.Balance{
		Address: baseAcc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount)),
	}

	s.SetupWithGenesisValSet(s.valSet, []authtypes.GenesisAccount{baseAcc}, balance)

	// Create StateDB
	s.stateDB = statedb.New(s.ctx, s.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(s.ctx.HeaderHash().Bytes())))

	// bond denom
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	s.bondDenom = stakingParams.BondDenom
	err = s.app.StakingKeeper.SetParams(s.ctx, stakingParams)
	s.Require().NoError(err)

	s.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	precompile, err := distribution.NewPrecompile(s.app.DistrKeeper, s.app.StakingKeeper, s.app.AuthzKeeper)
	s.Require().NoError(err)
	s.precompile = precompile

	coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(5_000_000_000_000_000_000)))
	inflCoins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(2_000_000_000_000_000_000)))
	distrCoins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(3_000_000_000_000_000_000)))
	err = s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, authtypes.FeeCollectorName, inflCoins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, distrtypes.ModuleName, distrCoins)
	s.Require().NoError(err)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, s.app.EvmKeeper)
	s.queryClientEVM = evmtypes.NewQueryClient(queryHelperEvm)
}

// DeployContract deploys a contract that calls the distribution precompile's methods for testing purposes.
func (s *PrecompileTestSuite) DeployContract(contract evmtypes.CompiledContract) (addr common.Address, err error) {
	addr, err = evmosutil.DeployContract(
		s.ctx,
		s.app,
		s.privKey,
		s.queryClientEVM,
		contract,
	)
	return
}

>>>>>>> main
type stakingRewards struct {
	Delegator sdk.AccAddress
	Validator stakingtypes.Validator
	RewardAmt math.Int
}

var (
	testRewardsAmt, _       = math.NewIntFromString("1000000000000000000")
	validatorCommPercentage = math.LegacyNewDecWithPrec(5, 2) // 5% commission
	validatorCommAmt        = math.LegacyNewDecFromInt(testRewardsAmt).Mul(validatorCommPercentage).TruncateInt()
	expRewardsAmt           = testRewardsAmt.Sub(validatorCommAmt) // testRewardsAmt - commission
)

// prepareStakingRewards prepares the test suite for testing delegation rewards.
//
// Specified rewards amount are allocated to the specified validator using the distribution keeper,
// such that the given amount of tokens is outstanding as a staking reward for the account.
//
// The setup is done in the following way:
//   - Fund distribution module to pay for rewards.
//   - Allocate rewards to the validator.
func (s *PrecompileTestSuite) prepareStakingRewards(ctx sdk.Context, stkRs ...stakingRewards) (sdk.Context, error) {
	for _, r := range stkRs {
		// set distribution module account balance which pays out the rewards
		coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, r.RewardAmt))
		if err := s.mintCoinsForDistrMod(ctx, coins); err != nil {
			return ctx, err
		}

		// allocate rewards to validator
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(s.bondDenom, r.RewardAmt))
		if err := s.network.App.DistrKeeper.AllocateTokensToValidator(ctx, r.Validator, allocatedRewards); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// mintCoinsForDistrMod is a helper function to mint a specific amount of coins from the
// distribution module to pay for staking rewards.
func (s *PrecompileTestSuite) mintCoinsForDistrMod(ctx sdk.Context, amount sdk.Coins) error {
	// Minting tokens for the FeeCollector to simulate fee accrued.
	if err := s.network.App.BankKeeper.MintCoins(
		ctx,
		inflationtypes.ModuleName,
		amount,
	); err != nil {
		return err
	}

	return s.network.App.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		inflationtypes.ModuleName,
		distrtypes.ModuleName,
		amount,
	)
}

// fundAccountWithBaseDenom is a helper function to fund a given address with the chain's
// base denomination.
func (s *PrecompileTestSuite) fundAccountWithBaseDenom(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error {
	coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, amount))
	if err := s.network.App.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, coins); err != nil {
		return err
	}
	return s.network.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, addr, coins)
}

func (s *PrecompileTestSuite) getStakingPrecompile() (*staking.Precompile, error) {
	return staking.NewPrecompile(
		s.network.App.StakingKeeper,
		s.network.App.AuthzKeeper,
	)
}

func generateKeys(count int) []keyring.Key {
	accs := make([]keyring.Key, 0, count)
	for i := 0; i < count; i++ {
		acc := keyring.NewKey()
		accs = append(accs, acc)
	}
	return accs
}
