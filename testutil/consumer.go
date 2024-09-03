// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package testutil

import (
	"time"

	ibctypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ccvprovidertypes "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"
	ccvtypes "github.com/cosmos/interchain-security/v4/x/ccv/types"
)

// This function creates consumer module genesis state that is used as starting point for modifications
// that allow Evmos chain to be started locally without having to start the provider chain and the relayer.
// Ref: https://github.com/Stride-Labs/stride/blob/4cfda614e8fb9664ce72861d32824d72430d4436/testutil/consumer.go#L16-L36
func CreateMinimalConsumerTestGenesis() *ccvtypes.ConsumerGenesisState {
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
