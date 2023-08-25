package network

import (
	"encoding/json"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	abci "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	infltypes "github.com/evmos/evmos/v14/x/inflation/types"
	revtypes "github.com/evmos/evmos/v14/x/revenue/v1/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
)

// --------------------------------------

// NetworkManager is the interface that wraps the basic methods of the network.
// TODO: Add a more detailed description of the purpose of the manager interface
type NetworkManager interface {
	// GetContext returns the context
	GetContext() sdktypes.Context
	// GetChainID returns the chain id
	GetChainID() string
	// GetDenom returns the denom
	GetDenom() string

	// CommitBlock commits a block
	CommitBlock() error

    // GRPC Clients
    GetEvmClient() evmtypes.QueryClient
    GetRevenueClient() revtypes.QueryClient
    GetInflationClient() infltypes.QueryClient
    GetBankClient() banktypes.QueryClient
    GetFeeMarketClient() feemarkettypes.QueryClient


	// Because to update the params on a conventional manner governance
	// would be require, we should provide an easier way to update the params
	UpdateRevenueParams(params revtypes.Params) error
	UpdateInflationParams(params infltypes.Params) error
	UpdateEvmParams(params evmtypes.Params) error
}

var (
	_ NetworkManager = (*Network)(nil)
)

type Network struct {
	cfg NetworkConfig
	ctx sdktypes.Context

	App        *app.Evmos
	Validators []stakingtypes.Validator
}

// New creates a new Network instance with the given options.
// It panics if an error occurs.
func New(config NetworkConfig) *Network {
	ctx := sdktypes.Context{}
	network := &Network{
		cfg:        config,
		ctx:        ctx,
		Validators: []stakingtypes.Validator{},
	}

	err := network.configureAndInitChain()
	if err != nil {
		panic(err)
	}
	return network
}

// ------ Initial Setup ------

var (
	// bondedAmt is the amount of tokens that each validator will have initially bonded
	bondedAmt = sdktypes.TokensFromConsensusPower(1, types.PowerReduction)
	// PrefundedAccountInitialBalance is the amount of tokens that each prefunded account have at genesis
	PrefundedAccountInitialBalance = sdktypes.NewInt(int64(math.Pow10(18) * 4))
)

// configureAndInitChain initializes the network with the given configuration.
// It creates the genesis state and starts the network.
func (n *Network) configureAndInitChain() error {
	// Create funded accounts based on the config and
	// create genesis accounts
	coin := sdktypes.NewCoin(n.cfg.denom, PrefundedAccountInitialBalance)
	genAccounts := createGenesisAccounts(n.cfg.preFundedAccounts, coin)
	fundedAccountBalances := createBalances(n.cfg.preFundedAccounts, coin)

	// Create validator set with the amount of validators specified in the config
	// with the default power of 1.
	valSet := createValidatorSet(n.cfg.amountOfValidators)
	totalBonded := bondedAmt.Mul(sdktypes.NewInt(int64(n.cfg.amountOfValidators)))

	// Build staking type validators and delegations
	validators, err := createStakingValidators(valSet.Validators, bondedAmt)
	if err != nil {
		return err
	}

	fundedAccountBalances = addBondedModuleAccountToFundedBalances(fundedAccountBalances, sdktypes.NewCoin(n.cfg.denom, totalBonded))

	delegations := createDelegations(valSet.Validators, genAccounts[0].GetAddress())

	// Create a new EvmosApp with the following params
	evmosApp := createEvmosApp(n.cfg.chainID)

	// Configure Genesis state
	genesisState := app.NewDefaultGenesisState()

	genesisState = setAuthGenesisState(evmosApp, genesisState, genAccounts)

	stakingParams := StakingCustomGenesisState{
		denom:       n.cfg.denom,
		validators:  validators,
		delegations: delegations,
	}
	genesisState = setStakingGenesisState(evmosApp, genesisState, stakingParams)

	genesisState = setInflationGenesisState(evmosApp, genesisState)

	totalSupply := calculateTotalSupply(fundedAccountBalances)
	bankParams := BankCustomGenesisState{
		totalSupply: totalSupply,
		balances:    fundedAccountBalances,
	}
	genesisState = setBankGenesisState(evmosApp, genesisState, bankParams)

	// Init chain
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		return err
	}

	evmosApp.InitChain(
		abci.RequestInitChain{
			ChainId:         n.cfg.chainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: app.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	// Commit genesis changes
	evmosApp.Commit()

	header := tmproto.Header{
		ChainID:            n.cfg.chainID,
		Height:             evmosApp.LastBlockHeight() + 1,
		AppHash:            evmosApp.LastCommitID().Hash,
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
		ProposerAddress:    valSet.Proposer.Address,
	}
	evmosApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Set networks global parameters
	n.App = evmosApp
	// TODO - this might not be the best way to initilize the context
	n.ctx = evmosApp.BaseApp.NewContext(false, header)
	n.Validators = validators
	return nil
}

// ------------------------------------------

func (n *Network) GetContext() sdktypes.Context {
	return n.ctx
}

func (n *Network) GetChainID() string {
	return n.cfg.chainID
}

func (n *Network) GetDenom() string {
	return n.cfg.denom
}

// Module functions

// ----- EVM -----

func (n *Network) UpdateEvmParams(params evmtypes.Params) error {
	return n.App.EvmKeeper.SetParams(n.ctx, params)
}

func (n *Network) GetEvmPrecompiles(precompiles ...common.Address) map[common.Address]vm.PrecompiledContract {
	return n.App.EvmKeeper.Precompiles(
		precompiles...,
	)
}

// ----- Bank -----
func (n *Network) FundAccount(addr sdktypes.AccAddress, coins sdktypes.Coins) error {
	if err := n.App.BankKeeper.MintCoins(n.ctx, infltypes.ModuleName, coins); err != nil {
		return err
	}

	return n.App.BankKeeper.SendCoinsFromModuleToAccount(n.ctx, infltypes.ModuleName, addr, coins)
}

// ----- Revenue -----

func (n *Network) UpdateRevenueParams(params revtypes.Params) error {
	return n.App.RevenueKeeper.SetParams(n.ctx, params)
}

// ----- Inflation -----

func (n *Network) UpdateInflationParams(params infltypes.Params) error {
	return n.App.InflationKeeper.SetParams(n.ctx, params)
}
