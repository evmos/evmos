// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package stride_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	erc20types "github.com/evmos/evmos/v15/x/erc20/types"

	"github.com/evmos/evmos/v15/precompiles/outposts/stride"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	evmosapp "github.com/evmos/evmos/v15/app"
	evmosibc "github.com/evmos/evmos/v15/ibc/testing"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	evmosutil "github.com/evmos/evmos/v15/testutil"
	evmosutiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmostypes "github.com/evmos/evmos/v15/types"
	"github.com/evmos/evmos/v15/utils"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v15/x/feemarket/types"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/types"
)

const (
	portID    = "transfer"
	channelID = "channel-0"
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

	totalBondAmt := bondAmt.Mul(sdk.NewInt(int64(len(validators))))
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

	feeGenesis := feemarkettypes.NewGenesisState(feemarkettypes.DefaultGenesisState().Params, 0)
	genesisState[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feeGenesis)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			ChainId:         cmn.DefaultChainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: evmosapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
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

	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// create Contexts
	s.ctx = app.BaseApp.NewContext(false, header)
	s.app = app
}

func (s *PrecompileTestSuite) DoSetupTest() {
	s.defaultExpirationDuration = s.ctx.BlockTime().Add(cmn.DefaultExpirationDuration).UTC()

	// generate validators private/public key
	var (
		validatorsPerChain = 2
		validators         []*tmtypes.Validator
		signersByAddress   = make(map[string]tmtypes.PrivValidator, validatorsPerChain)
	)

	for i := 0; i < validatorsPerChain; i++ {
		privVal := mock.NewPV()
		pubKey, err := privVal.GetPubKey()
		s.Require().NoError(err)
		validators = append(validators, tmtypes.NewValidator(pubKey, 1))
		signersByAddress[pubKey.Address().String()] = privVal
	}

	// construct validator set;
	// Note that the validators are sorted by voting power
	// or, if equal, by address lexical order
	s.valSet = tmtypes.NewValidatorSet(validators)

	// Create a coordinator and 2 test chains that will be used in the testing suite
	chains := make(map[string]*ibctesting.TestChain)
	s.coordinator = &ibctesting.Coordinator{
		T: s.T(),
		// NOTE: This year has to be updated otherwise the client will be shown as expired
		CurrentTime: time.Date(time.Now().Year()+1, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	// Create 2 Evmos chains
	chains[cmn.DefaultChainID] = s.NewTestChainWithValSet(s.coordinator, s.valSet, signersByAddress)
	chainID2 := utils.MainnetChainID + "-2"
	chains[chainID2] = ibctesting.NewTestChain(s.T(), s.coordinator, chainID2)
	s.coordinator.Chains = chains

	// Setup Chains in the testing suite
	s.chainA = s.coordinator.GetChain(cmn.DefaultChainID)
	s.chainB = s.coordinator.GetChain(chainID2)

	if s.suiteIBCTesting {
		s.setupIBCTest()
	}
}

func (s *PrecompileTestSuite) NewTestChainWithValSet(coord *ibctesting.Coordinator, valSet *tmtypes.ValidatorSet, signers map[string]tmtypes.PrivValidator) *ibctesting.TestChain {
	// generate genesis account
	addr, priv := evmosutiltx.NewAddrKey()
	s.privKey = priv
	s.address = addr
	// differentAddr is an address generated for testing purposes that e.g. raises the different origin error
	s.differentAddr = evmosutiltx.GenerateAddress()
	s.signer = evmosutiltx.NewSigner(priv)

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

	s.SetupWithGenesisValSet(s.valSet, []authtypes.GenesisAccount{acc}, balance)

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: cmn.DefaultChainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := s.app.GetTxConfig()

	// Create StateDB
	s.stateDB = statedb.New(s.ctx, s.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(s.ctx.HeaderHash().Bytes())))

	// bond denom
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	s.bondDenom = stakingParams.BondDenom
	err := s.app.StakingKeeper.SetParams(s.ctx, stakingParams)
	s.Require().NoError(err)

	s.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	// Setting up the fee market to 0 so the transactions don't fail in IBC testing
	s.app.FeeMarketKeeper.SetBaseFee(s.ctx, big.NewInt(0))
	s.app.FeeMarketKeeper.SetBlockGasWanted(s.ctx, 0)
	s.app.FeeMarketKeeper.SetTransientBlockGasWanted(s.ctx, 0)

	precompile, err := stride.NewPrecompile(portID, channelID, s.app.TransferKeeper, s.app.Erc20Keeper, s.app.AuthzKeeper, s.app.StakingKeeper)
	s.Require().NoError(err)
	s.precompile = precompile

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, s.app.EvmKeeper)
	s.queryClientEVM = evmtypes.NewQueryClient(queryHelperEvm)

	// Registered the supported Coin and ERC20
	s.registerStrideCoinERC20()

	// create an account to send transactions from
	chain := &ibctesting.TestChain{
		T:              s.T(),
		Coordinator:    coord,
		ChainID:        cmn.DefaultChainID,
		App:            s.app,
		CurrentHeader:  header,
		QueryServer:    s.app.GetIBCKeeper(),
		TxConfig:       txConfig,
		Codec:          s.app.AppCodec(),
		Vals:           valSet,
		NextVals:       valSet,
		Signers:        signers,
		SenderPrivKey:  priv,
		SenderAccount:  acc,
		SenderAccounts: []ibctesting.SenderAccount{{SenderPrivKey: priv, SenderAccount: acc}},
	}

	coord.CommitBlock(chain)

	return chain
}

// NewPrecompileContract creates a new precompile contract and sets the gas meter
func (s *PrecompileTestSuite) NewPrecompileContract(gas uint64) *vm.Contract {
	contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), gas)

	s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	initialGas := s.ctx.GasMeter().GasConsumed()
	s.Require().Zero(initialGas)

	return contract
}

// NewTransferPath creates a new path between two chains with the specified portIds and version.
func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = transfertypes.PortID
	path.EndpointB.ChannelConfig.PortID = transfertypes.PortID
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

// setupIBCTest makes the necessary setup of chains A & B
// for integration tests
func (s *PrecompileTestSuite) setupIBCTest() {
	s.coordinator.CommitNBlocks(s.chainA, 2)
	s.coordinator.CommitNBlocks(s.chainB, 2)

	s.app = s.chainA.App.(*evmosapp.Evmos)
	evmParams := s.app.EvmKeeper.GetParams(s.chainA.GetContext())
	evmParams.EvmDenom = utils.BaseDenom
	err := s.app.EvmKeeper.SetParams(s.chainA.GetContext(), evmParams)
	s.Require().NoError(err)

	// Set block proposer once, so its carried over on the ibc-go-testing suite
	validators := s.app.StakingKeeper.GetValidators(s.chainA.GetContext(), 2)
	cons, err := validators[0].GetConsAddr()
	s.Require().NoError(err)
	s.chainA.CurrentHeader.ProposerAddress = cons.Bytes()

	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.chainA.GetContext(), validators[0])
	s.Require().NoError(err)

	_, err = s.app.EvmKeeper.GetCoinbaseAddress(s.chainA.GetContext(), sdk.ConsAddress(s.chainA.CurrentHeader.ProposerAddress))
	s.Require().NoError(err)

	// Mint coins locked on the evmos account generated with secp.
	amt, ok := sdk.NewIntFromString("1000000000000000000000")
	s.Require().True(ok)
	coinEvmos := sdk.NewCoin(utils.BaseDenom, amt)
	coins := sdk.NewCoins(coinEvmos)
	err = s.app.BankKeeper.MintCoins(s.chainA.GetContext(), inflationtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), inflationtypes.ModuleName, s.chainA.SenderAccount.GetAddress(), coins)
	s.Require().NoError(err)

	s.transferPath = evmosibc.NewTransferPath(s.chainA, s.chainB) // clientID, connectionID, channelID empty
	evmosibc.SetupPath(s.coordinator, s.transferPath)             // clientID, connectionID, channelID filled
	s.Require().Equal("07-tendermint-0", s.transferPath.EndpointA.ClientID)
	s.Require().Equal("connection-0", s.transferPath.EndpointA.ConnectionID)
	s.Require().Equal("channel-0", s.transferPath.EndpointA.ChannelID)
}

// registerStrideCoinERC20 registers stEvmos and Evmos coin as an ERC20 token
func (s *PrecompileTestSuite) registerStrideCoinERC20() {
	// Register EVMOS ERC20 equivalent
	bondDenom := s.app.StakingKeeper.BondDenom(s.ctx)
	evmosMetadata := banktypes.Metadata{
		Description: "The native token of Evmos",
		Base:        bondDenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    bondDenom,
				Exponent: 0,
				Aliases:  []string{"aevmos"},
			},
			{
				Denom:    "aevmos",
				Exponent: 18,
			},
		},
		Name:    "Evmos",
		Symbol:  "EVMOS",
		Display: "evmos",
	}

	coin := sdk.NewCoin(evmosMetadata.Base, sdk.NewInt(2e18))
	err := s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, sdk.NewCoins(coin))
	s.Require().NoError(err)

	// Register some Token Pairs
	_, err = s.app.Erc20Keeper.RegisterCoin(s.ctx, evmosMetadata)
	s.Require().NoError(err)

	// Register stEvmos Token Pair
	denomTrace := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: "st" + bondDenom,
	}
	s.app.TransferKeeper.SetDenomTrace(s.ctx, denomTrace)
	stEvmosMetadata := banktypes.Metadata{
		Description: "The native token of Evmos",
		Base:        denomTrace.IBCDenom(),
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denomTrace.IBCDenom(),
				Exponent: 0,
				Aliases:  []string{"stEvmos"},
			},
			{
				Denom:    "stEvmos",
				Exponent: 18,
			},
		},
		Name:    "stEvmos",
		Symbol:  "STEVMOS",
		Display: "stEvmos",
	}

	stEvmos := sdk.NewCoin(stEvmosMetadata.Base, sdk.NewInt(9e18))
	err = s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, sdk.NewCoins(stEvmos))
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, inflationtypes.ModuleName, s.address.Bytes(), sdk.NewCoins(stEvmos))
	s.Require().NoError(err)

	// Register some Token Pairs
	_, err = s.app.Erc20Keeper.RegisterCoin(s.ctx, stEvmosMetadata)
	s.Require().NoError(err)

	convertCoin := erc20types.NewMsgConvertCoin(
		stEvmos,
		s.address,
		s.address.Bytes(),
	)

	_, err = s.app.Erc20Keeper.ConvertCoin(s.ctx, convertCoin)
	s.Require().NoError(err)
}
