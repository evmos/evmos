// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"fmt"
	"math"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"

	gethparams "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/version"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	commonnetwork "github.com/evmos/evmos/v16/testutil/integration/common/network"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
	revtypes "github.com/evmos/evmos/v16/x/revenue/v1/types"
)

// Network is the interface that wraps the methods to interact with integration test network.
//
// It was designed to avoid users to access module's keepers directly and force integration tests
// to be closer to the real user's behavior.
type Network interface {
	commonnetwork.Network

	GetEIP155ChainID() *big.Int
	GetEVMChainConfig() *gethparams.ChainConfig

	// Clients
	GetERC20Client() erc20types.QueryClient
	GetEvmClient() evmtypes.QueryClient
	GetGovClient() govtypes.QueryClient
	GetRevenueClient() revtypes.QueryClient
	GetInflationClient() infltypes.QueryClient
	GetFeeMarketClient() feemarkettypes.QueryClient
}

var _ Network = (*IntegrationNetwork)(nil)

// IntegrationNetwork is the implementation of the Network interface for integration tests.
type IntegrationNetwork struct {
	cfg        Config
	ctx        sdktypes.Context
	validators []stakingtypes.Validator
	app        *app.Evmos

	// This is only needed for IBC chain testing setup
	valSet     *cmttypes.ValidatorSet
	valSigners map[string]cmttypes.PrivValidator
}

// New configures and initializes a new integration Network instance with
// the given configuration options. If no configuration options are provided
// it uses the default configuration.
//
// It panics if an error occurs.
func New(opts ...ConfigOption) *IntegrationNetwork {
	cfg := DefaultConfig()
	// Modify the default config with the given options
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx := sdktypes.Context{}
	network := &IntegrationNetwork{
		cfg:        cfg,
		ctx:        ctx,
		validators: []stakingtypes.Validator{},
	}

	err := network.configureAndInitChain()
	if err != nil {
		panic(err)
	}
	return network
}

var (
	// DefaultBondedAmount is the amount of tokens that each validator will have initially bonded
	DefaultBondedAmount = sdktypes.TokensFromConsensusPower(1, types.PowerReduction)
	// PrefundedAccountInitialBalance is the amount of tokens that each prefunded account has at genesis
	PrefundedAccountInitialBalance, _ = sdkmath.NewIntFromString("100000000000000000000000") // 100k
)

// configureAndInitChain initializes the network with the given configuration.
// It creates the genesis state and starts the network.
func (n *IntegrationNetwork) configureAndInitChain() error {
	// Create validator set with the amount of validators specified in the config
	// with the default power of 1.
	valSet, valSigners := createValidatorSetAndSigners(n.cfg.amountOfValidators)
	totalBonded := DefaultBondedAmount.Mul(sdkmath.NewInt(int64(n.cfg.amountOfValidators)))

	// Build staking type validators and delegations
	validators, err := createStakingValidators(valSet.Validators, DefaultBondedAmount, n.cfg.operatorsAddrs)
	if err != nil {
		return err
	}

	// Create genesis accounts and funded balances based on the config
	genAccounts, fundedAccountBalances := getGenAccountsAndBalances(n.cfg, validators)

	fundedAccountBalances = addBondedModuleAccountToFundedBalances(
		fundedAccountBalances,
		sdktypes.NewCoin(n.cfg.denom, totalBonded),
	)

	delegations := createDelegations(validators, genAccounts[0].GetAddress())

	// Create a new EvmosApp with the following params
	evmosApp := createEvmosApp(n.cfg.chainID, n.cfg.customBaseAppOpts...)

	stakingParams := StakingCustomGenesisState{
		denom:       n.cfg.denom,
		validators:  validators,
		delegations: delegations,
	}
	totalSupply := calculateTotalSupply(fundedAccountBalances)
	bankParams := BankCustomGenesisState{
		totalSupply: totalSupply,
		balances:    fundedAccountBalances,
	}

	// Get the corresponding slashing info and missed block info
	// for the created validators
	slashingParams, err := getValidatorsSlashingGen(validators, evmosApp.StakingKeeper)
	if err != nil {
		return err
	}

	// Configure Genesis state
	genesisState := newDefaultGenesisState(
		evmosApp,
		defaultGenesisParams{
			genAccounts: genAccounts,
			staking:     stakingParams,
			bank:        bankParams,
			slashing:    slashingParams,
		},
	)

	// modify genesis state if there're any custom genesis state
	// for specific modules
	genesisState, err = customizeGenesis(evmosApp, n.cfg.customGenesisState, genesisState)
	if err != nil {
		return err
	}

	// Init chain
	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	if err != nil {
		return err
	}

	// Consensus module does not have a genesis state on the app,
	// but can customize the consensus parameters of the chain on initialization
	consensusParams := app.DefaultConsensusParams
	if gen, ok := n.cfg.customGenesisState[consensustypes.ModuleName]; ok {
		consensusParams, ok = gen.(*cmtproto.ConsensusParams)
		if !ok {
			return fmt.Errorf("invalid type for consensus parameters. Expected: cmtproto.ConsensusParams, got %T", gen)
		}
	}

	now := time.Now().UTC()
	if _, err := evmosApp.InitChain(
		&abcitypes.RequestInitChain{
			Time:            now,
			ChainId:         n.cfg.chainID,
			Validators:      []abcitypes.ValidatorUpdate{},
			ConsensusParams: consensusParams,
			AppStateBytes:   stateBytes,
		},
	); err != nil {
		return err
	}

	req := &abcitypes.RequestFinalizeBlock{
		Height:             evmosApp.LastBlockHeight() + 1,
		Hash:               evmosApp.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
		ProposerAddress:    valSet.Proposer.Address,
		Time:               now,
	}

	if _, err := evmosApp.FinalizeBlock(req); err != nil {
		return err
	}

	header := cmtproto.Header{
		ChainID:            n.cfg.chainID,
		Height:             req.Height,
		AppHash:            req.Hash,
		Time:               now,
		ValidatorsHash:     req.NextValidatorsHash,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
	}
	// TODO - this might not be the best way to initilize the context
	n.ctx = evmosApp.BaseApp.NewContextLegacy(false, header)

	// Commit genesis changes
	if _, err := evmosApp.Commit(); err != nil {
		return err
	}

	// Set networks global parameters
	var blockMaxGas uint64 = math.MaxUint64
	if consensusParams.Block != nil && consensusParams.Block.MaxGas > 0 {
		blockMaxGas = uint64(consensusParams.Block.MaxGas)
	}

	n.app = evmosApp
	n.ctx = n.ctx.WithConsensusParams(*consensusParams)
	n.ctx = n.ctx.WithBlockGasMeter(types.NewInfiniteGasMeterWithLimit(blockMaxGas))

	n.validators = validators
	n.valSet = valSet
	n.valSigners = valSigners

	return nil
}

// GetContext returns the network's context
func (n *IntegrationNetwork) GetContext() sdktypes.Context {
	return n.ctx
}

// WithIsCheckTxCtx switches the network's checkTx property
func (n *IntegrationNetwork) WithIsCheckTxCtx(isCheckTx bool) sdktypes.Context {
	n.ctx = n.ctx.WithIsCheckTx(isCheckTx)
	return n.ctx
}

// GetChainID returns the network's chainID
func (n *IntegrationNetwork) GetChainID() string {
	return n.cfg.chainID
}

// GetEIP155ChainID returns the network EIp-155 chainID number
func (n *IntegrationNetwork) GetEIP155ChainID() *big.Int {
	return n.cfg.eip155ChainID
}

// GetChainConfig returns the network's chain config
func (n *IntegrationNetwork) GetEVMChainConfig() *gethparams.ChainConfig {
	params := n.app.EvmKeeper.GetParams(n.ctx)
	return params.ChainConfig.EthereumConfig(n.cfg.eip155ChainID)
}

// GetDenom returns the network's denom
func (n *IntegrationNetwork) GetDenom() string {
	return n.cfg.denom
}

// GetOtherDenoms returns network's other supported denoms
func (n *IntegrationNetwork) GetOtherDenoms() []string {
	return n.cfg.otherCoinDenom
}

// GetValidators returns the network's validators
func (n *IntegrationNetwork) GetValidators() []stakingtypes.Validator {
	return n.validators
}

// BroadcastTxSync broadcasts the given txBytes to the network and returns the response.
// TODO - this should be change to gRPC
func (n *IntegrationNetwork) BroadcastTxSync(txBytes []byte) (abcitypes.ExecTxResult, error) {
	req := abcitypes.RequestFinalizeBlock{
		Time:               n.ctx.BlockTime(),
		Height:             n.app.LastBlockHeight() + 1,
		Hash:               n.app.LastCommitID().Hash,
		NextValidatorsHash: n.valSet.Hash(),
		ProposerAddress:    n.valSet.Proposer.Address,
		Txs:                [][]byte{txBytes},
	}
	blockRes, err := n.app.BaseApp.FinalizeBlock(&req)
	if err != nil {
		return abcitypes.ExecTxResult{}, err
	}
	if len(blockRes.TxResults) != 1 {
		return abcitypes.ExecTxResult{}, fmt.Errorf("unexpected number of tx results. Expected 1, got: %d", len(blockRes.TxResults))
	}
	return *blockRes.TxResults[0], nil
}

// Simulate simulates the given txBytes to the network and returns the simulated response.
// TODO - this should be change to gRPC
func (n *IntegrationNetwork) Simulate(txBytes []byte) (*txtypes.SimulateResponse, error) {
	gas, result, err := n.app.BaseApp.Simulate(txBytes)
	if err != nil {
		return nil, err
	}
	return &txtypes.SimulateResponse{
		GasInfo: &gas,
		Result:  result,
	}, nil
}

// CheckTx calls the BaseApp's CheckTx method with the given txBytes to the network and returns the response.
func (n *IntegrationNetwork) CheckTx(txBytes []byte) (*abcitypes.ResponseCheckTx, error) {
	req := &abcitypes.RequestCheckTx{Tx: txBytes}
	res, err := n.app.BaseApp.CheckTx(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
