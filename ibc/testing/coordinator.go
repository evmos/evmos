// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ibctesting

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/evmos/evmos/v17/app"
	"github.com/stretchr/testify/require"
)

const DefaultFeeAmt = int64(150_000_000_000_000_000) // 0.15 EVMOS

var globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

// NewCoordinator initializes Coordinator with N EVM TestChain's (Evmos apps) and M Cosmos chains (Simulation Apps)
func NewCoordinator(t *testing.T, nEVMChains, mCosmosChains int) *ibctesting.Coordinator {
	chains := make(map[string]*ibctesting.TestChain)
	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: globalStartTime,
	}

	for i := 1; i <= nEVMChains; i++ {
		chainID := ibctesting.GetChainID(i)
		// setup EVM chains
		ibctesting.DefaultTestingAppInit = DefaultTestingAppInit(chainID)
		chains[chainID] = NewTestChain(t, coord, chainID)
	}

	// setup Cosmos chains
	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp

	for j := 1 + nEVMChains; j <= nEVMChains+mCosmosChains; j++ {
		chainID := ibctesting.GetChainID(j)
		chains[chainID] = ibctesting.NewTestChain(t, coord, chainID)
	}

	coord.Chains = chains

	return coord
}

// SetupPath constructs a TM client, connection, and channel on both chains provided. It will
// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
// for both chains. The channels created are connected to the ibc-transfer application.
func SetupPath(coord *ibctesting.Coordinator, path *Path) {
	SetupConnections(coord, path)

	// channels can also be referenced through the returned connections
	CreateChannels(coord, path)
}

// SetupClientConnections is a helper function to create clients and the appropriate
// connections on both the source and counterparty chain. It assumes the caller does not
// anticipate any errors.
func SetupConnections(coord *ibctesting.Coordinator, path *Path) {
	SetupClients(coord, path)

	CreateConnections(coord, path)
}

// CreateChannel constructs and executes channel handshake messages in order to create
// OPEN channels on chainA and chainB. The function expects the channels to be successfully
// opened otherwise testing will fail.
func CreateChannels(coord *ibctesting.Coordinator, path *Path) {
	err := path.EndpointA.ChanOpenInit()
	require.NoError(coord.T, err)

	err = path.EndpointB.ChanOpenTry()
	require.NoError(coord.T, err)

	err = path.EndpointA.ChanOpenAck()
	require.NoError(coord.T, err)

	err = path.EndpointB.ChanOpenConfirm()
	require.NoError(coord.T, err)

	// ensure counterparty is up to date
	err = path.EndpointA.UpdateClient()
	require.NoError(coord.T, err)
}

// CreateConnection constructs and executes connection handshake messages in order to create
// OPEN channels on chainA and chainB. The connection information of for chainA and chainB
// are returned within a TestConnection struct. The function expects the connections to be
// successfully opened otherwise testing will fail.
func CreateConnections(coord *ibctesting.Coordinator, path *Path) {
	err := path.EndpointA.ConnOpenInit()
	require.NoError(coord.T, err)

	err = path.EndpointB.ConnOpenTry()
	require.NoError(coord.T, err)

	err = path.EndpointA.ConnOpenAck()
	require.NoError(coord.T, err)

	err = path.EndpointB.ConnOpenConfirm()
	require.NoError(coord.T, err)

	// ensure counterparty is up to date
	err = path.EndpointA.UpdateClient()
	require.NoError(coord.T, err)
}

// SetupClients is a helper function to create clients on both chains. It assumes the
// caller does not anticipate any errors.
func SetupClients(coord *ibctesting.Coordinator, path *Path) {
	err := path.EndpointA.CreateClient()
	require.NoError(coord.T, err)

	err = path.EndpointB.CreateClient()
	require.NoError(coord.T, err)
}

func SendMsgs(chain *ibctesting.TestChain, feeAmt int64, msgs ...sdk.Msg) (*sdk.Result, error) {
	var bondDenom string
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	if evmosChain, ok := chain.App.(*app.Evmos); ok {
		bondDenom = evmosChain.StakingKeeper.BondDenom(chain.GetContext())
	} else {
		bondDenom = chain.GetSimApp().StakingKeeper.BondDenom(chain.GetContext())
	}

	fee := sdk.Coins{sdk.NewInt64Coin(bondDenom, feeAmt)}
	_, r, err := SignAndDeliver(
		chain.T,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		msgs,
		fee,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true, chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// NextBlock calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

// SignAndDeliver signs and delivers a transaction. No simulation occurs as the
// ibc testing package causes checkState and deliverState to diverge in block time.
//
// CONTRACT: BeginBlock must be called before this function.
// Is a customization of IBC-go function that allows to modify the fee denom and amount
// IBC-go implementation: https://github.com/cosmos/ibc-go/blob/d34cef7e075dda1a24a0a3e9b6d3eff406cc606c/testing/simapp/test_helpers.go#L332-L364
func SignAndDeliver(
	t *testing.T, txCfg client.TxConfig, app *baseapp.BaseApp, msgs []sdk.Msg,
	fee sdk.Coins,
	chainID string, accNums, accSeqs []uint64, expPass bool, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		txCfg,
		msgs,
		fee,
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)

	// Simulate a sending a transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	return gInfo, res, err
}
