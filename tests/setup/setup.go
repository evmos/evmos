// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package setup

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"

	ibctesting "github.com/cosmos/ibc-go/v6/testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/tests"
	evmostypes "github.com/evmos/evmos/v11/types"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
)

// DefaultTestingAppInit defines the IBC application used for testing
var DefaultTestingAppInit func() (ibctesting.TestingApp, map[string]json.RawMessage) = SetupTestingApp

// DefaultConsensusParams defines the default Tendermint consensus params used in
// Evmos testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   -1, // no limit
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

var DefaultOptions = simapp.SetupOptions{
	Logger:             log.NewNopLogger(),
	DB:                 dbm.NewMemDB(),
	InvCheckPeriod:     0,
	HomePath:           app.DefaultNodeHome,
	SkipUpgradeHeights: nil,
	EncConfig:          encoding.MakeConfig(app.ModuleBasics),
	AppOpts:            simapp.EmptyAppOptions{},
}

func NewAppOptions()

type TestingEnv struct {
	genesis           simapp.GenesisState
	setupOptions      simapp.SetupOptions
	baseAppOptions    []func(*baseapp.BaseApp)
	ctx               sdk.Context
	app               *app.Evmos
	accounts          []tests.Account
	validatorAccounts []tests.Account
	validators        []stakingtypes.Validator
	denom             string
}

func (s *TestingEnv) setupEOAs(
	t testing.TB,
	numAccounts,
	numValidators uint64,
) ([]authtypes.GenesisAccount, []banktypes.Balance) {
	s.accounts = make([]tests.Account, numAccounts)

	genAccounts := make([]authtypes.GenesisAccount, numAccounts)
	balances := make([]banktypes.Balance, numAccounts+1)

	amount := sdk.NewInt(1).Mul(evmostypes.PowerReduction)
	bondedAmt := amount.MulRaw(int64(numValidators))

	totalSupply := sdk.Coins{}
	for i := uint64(0); i < numAccounts; i++ {
		acc := tests.NewEOAAccount(t)
		s.accounts[i] = acc

		baseAcc := authtypes.NewBaseAccount(acc.Address, acc.PubKey, i, 0)
		genAccounts[i] = &evmostypes.EthAccount{BaseAccount: baseAcc, CodeHash: common.BytesToHash(evmtypes.EmptyCodeHash).Hex()}

		coins := sdk.Coins{evmostypes.NewAEvmosCoin(amount)}
		balances[i] = banktypes.Balance{
			Address: acc.Address.String(),
			Coins:   coins,
		}

		totalSupply = totalSupply.Add(coins...)
	}

	// add delegated tokens to total supply
	totalSupply = totalSupply.Add(sdk.NewCoin(s.denom, bondedAmt))

	// add bonded amount to bonded pool module account
	balances[numAccounts] = banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(s.denom, bondedAmt)},
	}

	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccounts)
	s.genesis[authtypes.ModuleName] = s.app.AppCodec().MustMarshalJSON(authGenesis)

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	s.genesis[banktypes.ModuleName] = s.app.AppCodec().MustMarshalJSON(bankGenesis)

	return genAccounts, balances
}

func (s *TestingEnv) setupValidators(
	t testing.TB,
	numValidators uint64,
) {
	s.validatorAccounts = make([]tests.Account, numValidators)
	tmValidators := make([]*tmtypes.Validator, numValidators)
	s.validators = make([]stakingtypes.Validator, numValidators)
	delegations := make([]stakingtypes.Delegation, numValidators)

	bondAmt := evmostypes.PowerReduction

	for i := 0; i < int(numValidators); i++ {
		validatorAcc := tests.NewValidatorAccount(t)
		tmValidator := tmtypes.NewValidator(validatorAcc.TmPubKey, sdk.TokensToConsensusPower(sdk.OneInt(), evmostypes.PowerReduction))
		s.validatorAccounts[i] = validatorAcc
		tmValidators[i] = tmValidator

		pkAny, err := codectypes.NewAnyWithValue(validatorAcc.PubKey)
		require.NoError(t, err)

		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(validatorAcc.Address).String(),
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
		s.validators[i] = validator

		delegations[i] = stakingtypes.NewDelegation(s.accounts[i].Address, validator.GetOperator(), sdk.OneDec())
	}

	// set validators and delegations
	params := stakingtypes.DefaultParams()
	params.BondDenom = s.denom
	stakingGenesis := stakingtypes.NewGenesisState(params, s.validators, delegations)
	s.genesis[stakingtypes.ModuleName] = s.app.AppCodec().MustMarshalJSON(stakingGenesis)
}

// Setup initializes a new Evmos. A Nop logger is set in Evmos.
func (s *TestingEnv) Setup(
	t testing.TB,
	chainID string,
	numValidators,
	numAccounts uint64,
) {
	genAccs, balances := s.setupEOAs(t, numAccounts, numValidators)

	s.setupValidators(t, numValidators)

	s.app = app.NewEvmos(
		s.setupOptions.Logger,
		s.setupOptions.DB,
		nil,
		true,
		s.setupOptions.SkipUpgradeHeights,
		s.setupOptions.HomePath,
		s.setupOptions.InvCheckPeriod,
		s.setupOptions.EncConfig,
		s.setupOptions.AppOpts,
		s.baseAppOptions...,
	)

	stateBytes, err := json.MarshalIndent(s.genesis, "", " ")
	require.NoError(t, err)

	// Initialize the chain
	req := abci.RequestInitChain{
		ChainId:         chainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	}

	res := s.app.InitChain(req)
	header := s.NewHeader(1, time.Now().UTC(), chainID, s.validatorAccounts[0].Address.Bytes(), res.AppHash)
	s.ctx = s.app.NewContext(false, header)
}

func (s TestingEnv) NewHeader(
	height int64,
	blockTime time.Time,
	chainID string,
	proposer sdk.ConsAddress,
	appHash []byte,
) tmproto.Header {
	return tmproto.Header{
		Height:          height,
		ChainID:         chainID,
		Time:            blockTime,
		ProposerAddress: proposer.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            appHash,
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}
}
