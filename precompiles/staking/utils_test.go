package staking_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmosapp "github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/precompiles/authorization"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/staking"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/precompiles/testutil/contracts"
	evmosutil "github.com/evmos/evmos/v15/testutil"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
	"github.com/evmos/evmos/v15/utils"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/types"
	"golang.org/x/exp/slices"
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
	s.bondDenom = stakingParams.BondDenom
	err := s.app.StakingKeeper.SetParams(s.ctx, stakingParams)
	s.Require().NoError(err)

	s.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	precompile, err := staking.NewPrecompile(s.app.StakingKeeper, s.app.AuthzKeeper)
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
}

// ApproveAndCheckAuthz is a helper function to approve a given authorization method and check if the authorization was created.
func (s *PrecompileTestSuite) ApproveAndCheckAuthz(method abi.Method, msgType string, amount *big.Int) {
	approveArgs := []interface{}{
		s.address,
		amount,
		[]string{msgType},
	}
	resp, err := s.precompile.Approve(s.ctx, s.address, s.stateDB, &method, approveArgs)
	s.Require().NoError(err)
	s.Require().Equal(resp, cmn.TrueValue)

	auth, _ := s.CheckAuthorization(staking.DelegateAuthz, s.address, s.address)
	s.Require().NotNil(auth)
	s.Require().Equal(auth.AuthorizationType, staking.DelegateAuthz)
	s.Require().Equal(auth.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: sdk.NewIntFromBigInt(amount)})
}

// CheckAuthorization is a helper function to check if the authorization is set and if it is the correct type.
func (s *PrecompileTestSuite) CheckAuthorization(authorizationType stakingtypes.AuthorizationType, grantee, granter common.Address) (*stakingtypes.StakeAuthorization, *time.Time) {
	stakingAuthz := stakingtypes.StakeAuthorization{AuthorizationType: authorizationType}
	auth, expirationTime := s.app.AuthzKeeper.GetAuthorization(s.ctx, grantee.Bytes(), granter.Bytes(), stakingAuthz.MsgTypeURL())

	stakeAuthorization, ok := auth.(*stakingtypes.StakeAuthorization)
	if !ok {
		return nil, expirationTime
	}

	return stakeAuthorization, expirationTime
}

// CreateAuthorization is a helper function to create a new authorization of the given type for a spender address
// (=grantee).
// The authorization will be created to spend the given Coin.
// For testing purposes, this function will create a new authorization for all available validators,
// that are not jailed.
func (s *PrecompileTestSuite) CreateAuthorization(grantee common.Address, authzType stakingtypes.AuthorizationType, coin *sdk.Coin) error {
	// Get all available validators and filter out jailed validators
	validators := make([]sdk.ValAddress, 0)
	s.app.StakingKeeper.IterateValidators(
		s.ctx, func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
			if validator.IsJailed() {
				return
			}
			validators = append(validators, validator.GetOperator())
			return
		},
	)

	stakingAuthz, err := stakingtypes.NewStakeAuthorization(validators, nil, authzType, coin)
	if err != nil {
		return err
	}

	expiration := time.Now().Add(cmn.DefaultExpirationDuration).UTC()
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee.Bytes(), s.address.Bytes(), stakingAuthz, &expiration)
	if err != nil {
		return err
	}

	return nil
}

// SetupApproval sets up an approval, that authorizes the grantee to spend the given amount for the granter
// in transactions, that target the specified message types.
func (s *PrecompileTestSuite) SetupApproval(
	granterPriv types.PrivKey,
	grantee common.Address,
	amount *big.Int,
	msgTypes []string,
) {
	approveArgs := contracts.CallArgs{
		ContractAddr: s.precompile.Address(),
		ContractABI:  s.precompile.ABI,
		PrivKey:      granterPriv,
		MethodName:   authorization.ApproveMethod,
		Args: []interface{}{
			grantee, amount, msgTypes,
		},
	}

	logCheckArgs := testutil.LogCheckArgs{
		ABIEvents: s.precompile.Events,
		ExpEvents: []string{authorization.EventTypeApproval},
		ExpPass:   true,
	}

	res, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, approveArgs, logCheckArgs)
	Expect(err).To(BeNil(), "error while calling the contract to approve")

	s.NextBlock()

	// Check if the approval event is emitted
	granterAddr := common.BytesToAddress(granterPriv.PubKey().Address().Bytes())
	testutil.CheckAuthorizationEvents(
		s.precompile.Events[authorization.EventTypeApproval],
		s.precompile.Address(),
		granterAddr,
		grantee,
		res,
		s.ctx.BlockHeight()-1,
		msgTypes,
		amount,
	)
}

// SetupApprovalWithContractCalls is a helper function used to setup the allowance for the given spender.
func (s *PrecompileTestSuite) SetupApprovalWithContractCalls(approvalArgs contracts.CallArgs) {
	msgTypes, ok := approvalArgs.Args[1].([]string)
	Expect(ok).To(BeTrue(), "failed to convert msgTypes to []string")
	expAmount, ok := approvalArgs.Args[2].(*big.Int)
	Expect(ok).To(BeTrue(), "failed to convert amount to big.Int")

	logCheckArgs := testutil.LogCheckArgs{
		ABIEvents: s.precompile.Events,
		ExpEvents: []string{authorization.EventTypeApproval},
		ExpPass:   true,
	}

	_, _, err := contracts.CallContractAndCheckLogs(s.ctx, s.app, approvalArgs, logCheckArgs)
	Expect(err).To(BeNil(), "error while approving: %v", err)

	// iterate over args
	var expectedAuthz stakingtypes.AuthorizationType
	for _, msgType := range msgTypes {
		switch msgType {
		case staking.DelegateMsg:
			expectedAuthz = staking.DelegateAuthz
		case staking.UndelegateMsg:
			expectedAuthz = staking.UndelegateAuthz
		case staking.RedelegateMsg:
			expectedAuthz = staking.RedelegateAuthz
		case staking.CancelUnbondingDelegationMsg:
			expectedAuthz = staking.CancelUnbondingDelegationAuthz
		}
		authz, expirationTime := s.CheckAuthorization(expectedAuthz, approvalArgs.ContractAddr, s.address)
		Expect(authz).ToNot(BeNil(), "expected authorization to be set")
		Expect(authz.MaxTokens.Amount).To(Equal(sdk.NewInt(expAmount.Int64())), "expected different allowance")
		Expect(authz.MsgTypeURL()).To(Equal(msgType), "expected different message type")
		Expect(expirationTime).ToNot(BeNil(), "expected expiration time to not be nil")
	}
}

// DeployContract deploys a contract that calls the staking precompile's methods for testing purposes.
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

// NextBlock commits the current block and sets up the next block.
func (s *PrecompileTestSuite) NextBlock() {
	var err error
	s.ctx, err = evmosutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Second, nil)
	Expect(err).To(BeNil(), "failed to commit block")
}

// CheckAllowanceChangeEvent is a helper function used to check the allowance change event arguments.
func (s *PrecompileTestSuite) CheckAllowanceChangeEvent(log *ethtypes.Log, methods []string, amounts []*big.Int) {
	s.Require().Equal(log.Address, s.precompile.Address())
	// Check event signature matches the one emitted
	event := s.precompile.ABI.Events[authorization.EventTypeAllowanceChange]
	s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
	s.Require().Equal(log.BlockNumber, uint64(s.ctx.BlockHeight()))

	var approvalEvent authorization.EventAllowanceChange
	err := cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeAllowanceChange, *log)
	s.Require().NoError(err)
	s.Require().Equal(s.address, approvalEvent.Grantee)
	s.Require().Equal(s.address, approvalEvent.Granter)
	s.Require().Equal(len(methods), len(approvalEvent.Methods))

	for i, method := range methods {
		s.Require().Equal(method, approvalEvent.Methods[i])
		s.Require().Equal(amounts[i], approvalEvent.Values[i])
	}
}

// ExpectAuthorization is a helper function for tests using the Ginkgo BDD style tests, to check that the
// authorization is correctly set.
func (s *PrecompileTestSuite) ExpectAuthorization(authorizationType stakingtypes.AuthorizationType, grantee, granter common.Address, maxTokens *sdk.Coin) {
	authz, expirationTime := s.CheckAuthorization(authorizationType, grantee, granter)
	Expect(authz).ToNot(BeNil(), "expected authorization to be set")
	Expect(authz.AuthorizationType).To(Equal(authorizationType), "expected different authorization type")
	Expect(authz.MaxTokens).To(Equal(maxTokens), "expected different max tokens")
	Expect(expirationTime).ToNot(BeNil(), "expected expiration time to be not be nil")
}

// assertValidatorsResponse asserts all the fields on the validators response
func (s *PrecompileTestSuite) assertValidatorsResponse(validators []staking.ValidatorInfo, expLen int) {
	// returning order can change
	valOrder := []int{0, 1}
	if validators[0].OperatorAddress != s.validators[0].OperatorAddress {
		valOrder = []int{1, 0}
	}
	for i := 0; i < expLen; i++ {
		j := valOrder[i]
		s.Require().Equal(s.validators[j].OperatorAddress, validators[i].OperatorAddress)
		s.Require().Equal(uint8(s.validators[j].Status), validators[i].Status)
		s.Require().Equal(s.validators[j].Tokens.Uint64(), validators[i].Tokens.Uint64())
		s.Require().Equal(s.validators[j].DelegatorShares.BigInt(), validators[i].DelegatorShares)
		s.Require().Equal(s.validators[j].Jailed, validators[i].Jailed)
		s.Require().Equal(s.validators[j].UnbondingHeight, validators[i].UnbondingHeight)
		s.Require().Equal(int64(0), validators[i].UnbondingTime)
		s.Require().Equal(int64(0), validators[i].Commission.Int64())
		s.Require().Equal(int64(0), validators[i].MinSelfDelegation.Int64())
		s.Require().Contains(validators[i].ConsensusPubkey, fmt.Sprintf("%v", s.validators[j].ConsensusPubkey.Value))
	}
}

// assertRedelegation asserts the redelegationOutput struct and its fields
func (s *PrecompileTestSuite) assertRedelegationsOutput(data []byte, redelTotalCount uint64, expAmt *big.Int, expCreationHeight int64, hasPagination bool) {
	var redOut staking.RedelegationsOutput
	err := s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationsMethod, data)
	s.Require().NoError(err, "failed to unpack output")

	s.Require().Len(redOut.Response, 1)
	// check pagination - total count should be 2
	s.Require().Equal(redelTotalCount, redOut.PageResponse.Total)
	if hasPagination {
		s.Require().NotEmpty(redOut.PageResponse.NextKey)
	} else {
		s.Require().Empty(redOut.PageResponse.NextKey)
	}
	// check redelegation entry
	// order may change, one redelegation has 2 entries
	// and the other has one
	if len(redOut.Response[0].Entries) == 2 {
		s.assertRedelegation(redOut.Response[0],
			2,
			s.validators[0].OperatorAddress,
			s.validators[1].OperatorAddress,
			expAmt,
			expCreationHeight,
		)
	} else {
		s.assertRedelegation(redOut.Response[0],
			1,
			s.validators[0].OperatorAddress,
			sdk.ValAddress(s.address.Bytes()).String(),
			expAmt,
			expCreationHeight,
		)
	}
}

// assertRedelegation asserts all the fields on the redelegations response
// should specify the amount of entries expected and the expected amount for this
// the same amount is considered for all entries
func (s *PrecompileTestSuite) assertRedelegation(res staking.RedelegationResponse, entriesCount int, expValSrcAddr, expValDstAddr string, expAmt *big.Int, expCreationHeight int64) {
	// check response
	s.Require().Equal(res.Redelegation.DelegatorAddress, sdk.AccAddress(s.address.Bytes()).String())
	s.Require().Equal(res.Redelegation.ValidatorSrcAddress, expValSrcAddr)
	s.Require().Equal(res.Redelegation.ValidatorDstAddress, expValDstAddr)
	// check redelegation entries - should be empty
	s.Require().Empty(res.Redelegation.Entries)
	// check response entries, should be 2
	s.Require().Len(res.Entries, entriesCount)
	// check redelegation entries
	for _, e := range res.Entries {
		s.Require().Equal(e.Balance, expAmt)
		s.Require().True(e.RedelegationEntry.CompletionTime > 1600000000)
		s.Require().Equal(expCreationHeight, e.RedelegationEntry.CreationHeight)
		s.Require().Equal(e.RedelegationEntry.InitialBalance, expAmt)
	}
}

// setupRedelegations setups 2 entries for redelegation from validator[0]
// to validator[1], creates a validator using s.address
// and creates a redelegation from validator[0] to the new validator
func (s *PrecompileTestSuite) setupRedelegations(redelAmt *big.Int) error {
	msg := stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    sdk.AccAddress(s.address.Bytes()).String(),
		ValidatorSrcAddress: s.validators[0].OperatorAddress,
		ValidatorDstAddress: s.validators[1].OperatorAddress,
		Amount:              sdk.NewCoin(s.bondDenom, sdk.NewIntFromBigInt(redelAmt)),
	}

	msgSrv := stakingkeeper.NewMsgServerImpl(&s.app.StakingKeeper)
	// create 2 entries for same redelegation
	for i := 0; i < 2; i++ {
		if _, err := msgSrv.BeginRedelegate(s.ctx, &msg); err != nil {
			return err
		}
	}

	// create a validator with s.address and s.privKey
	// then create a redelegation from validator[0] to this new validator
	testutil.CreateValidator(s.ctx, s.T(), s.privKey.PubKey(), s.app.StakingKeeper, math.NewInt(100))
	msg.ValidatorDstAddress = sdk.ValAddress(s.address.Bytes()).String()
	_, err := msgSrv.BeginRedelegate(s.ctx, &msg)
	return err
}

// CheckValidatorOutput checks that the given validator output
func (s *PrecompileTestSuite) CheckValidatorOutput(valOut staking.ValidatorInfo) {
	validatorAddrs := make([]string, len(s.validators))
	for i, v := range s.validators {
		validatorAddrs[i] = v.OperatorAddress
	}
	Expect(slices.Contains(validatorAddrs, valOut.OperatorAddress)).To(BeTrue(), "operator address not found in test suite validators")
	Expect(valOut.DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
}
