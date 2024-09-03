// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package app

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	consumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"

	ibctypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	ccvprovidertypes "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"
	ccvtypes "github.com/cosmos/interchain-security/v4/x/ccv/types"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/encoding"
	"github.com/evmos/evmos/v19/utils"
)

// EthDefaultConsensusParams defines the default Tendermint consensus params used in
// EvmosApp testing.
var EthDefaultConsensusParams = &tmtypes.ConsensusParams{
	Block: &tmtypes.BlockParams{
		MaxBytes: 200000,
		MaxGas:   -1, // no limit
	},
	Evidence: &tmtypes.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmtypes.ValidatorParams{
		PubKeyTypes: []string{
			cmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// EthSetup initializes a new EvmosApp. A Nop logger is set in EvmosApp.
func EthSetup(isCheckTx bool, patchGenesis func(*Evmos, simapp.GenesisState) simapp.GenesisState) *Evmos {
	return EthSetupWithDB(isCheckTx, patchGenesis, dbm.NewMemDB())
}

// EthSetupWithDB initializes a new EvmosApp. A Nop logger is set in EvmosApp.
func EthSetupWithDB(isCheckTx bool, patchGenesis func(*Evmos, simapp.GenesisState) simapp.GenesisState, db dbm.DB) *Evmos {
	chainID := utils.TestnetChainID + "-1"
	app := NewEvmos(log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		5,
		encoding.MakeConfig(ModuleBasics),
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewTestGenesisState(app.AppCodec())
		if patchGenesis != nil {
			genesisState = patchGenesis(app, genesisState)
		}

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				ChainId:         chainID,
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// NewTestGenesisState generate genesis state with single validator
func NewTestGenesisState(codec codec.Codec) simapp.GenesisState {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}
	// create validator set with single validator
	validator := cmtypes.NewValidator(pubKey, 1)
	valSet := cmtypes.NewValidatorSet([]*cmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	genesisState := NewDefaultGenesisState()
	return genesisStateWithValSet(codec, genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)
}

func genesisStateWithValSet(codec codec.Codec, genesisState simapp.GenesisState,
	valSet *cmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) simapp.GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))
	// Initial validator powers is required to start the consumer chain InitGenesis.
	initValPowers := []abci.ValidatorUpdate{}

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			panic(err)
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			panic(err)
		}
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

		protoVal, _ := val.ToProto()
		initValPowers = append(initValPowers, abci.ValidatorUpdate{
			Power:  val.VotingPower,
			PubKey: protoVal.PubKey,
		})
	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	vals, err := cmtypes.PB2TM.ValidatorUpdates(initValPowers)
	if err != nil {
		panic("failed to get vals")
	}

	// Define the cross-chain validation module genesis.
	// Ref: https://github.com/Stride-Labs/stride/blob/4cfda614e8fb9664ce72861d32824d72430d4436/app/test_setup.go#L171-L175
	consumerGenesisState := createMinimalConsumerTestGenesis()
	consumerGenesisState.Provider.InitialValSet = initValPowers
	consumerGenesisState.Provider.ConsensusState.NextValidatorsHash = cmtypes.NewValidatorSet(vals).Hash()
	consumerGenesisState.Params.Enabled = true
	genesisState[consumertypes.ModuleName] = codec.MustMarshalJSON(consumerGenesisState)

	return genesisState
}

// This function creates consumer module genesis state that is used as starting point for modifications
// that allow Evmos chain to be started locally without having to start the provider chain and the relayer.
// Ref: https://github.com/Stride-Labs/stride/blob/4cfda614e8fb9664ce72861d32824d72430d4436/testutil/consumer.go#L16-L36
func createMinimalConsumerTestGenesis() *ccvtypes.ConsumerGenesisState {
	genesisState := ccvtypes.DefaultConsumerGenesisState()
	genesisState.Params.Enabled = true
	genesisState.NewChain = true
	genesisState.Provider.ClientState = ccvprovidertypes.DefaultParams().TemplateClient
	genesisState.Provider.ClientState.ChainId = "evmos"
	genesisState.Provider.ClientState.LatestHeight = ibctypes.Height{RevisionNumber: 0, RevisionHeight: 1}
	genesisState.Params.UnbondingPeriod = stakingtypes.DefaultUnbondingTime
	unbondingPeriod := genesisState.Params.UnbondingPeriod
	trustPeriod, err := ccvtypes.CalculateTrustPeriod(unbondingPeriod, ccvprovidertypes.DefaultTrustingPeriodFraction)
	if err != nil {
		panic("provider client trusting period error")
	}
	genesisState.Provider.ClientState.TrustingPeriod = trustPeriod
	genesisState.Provider.ClientState.UnbondingPeriod = unbondingPeriod
	genesisState.Provider.ClientState.MaxClockDrift = ccvprovidertypes.DefaultMaxClockDrift
	genesisState.Provider.ConsensusState = &ibctmtypes.ConsensusState{
		Timestamp: time.Now().UTC(),
		Root:      types.MerkleRoot{Hash: []byte("dummy")},
	}

	return genesisState
}
