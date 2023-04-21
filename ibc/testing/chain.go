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

package ibctesting

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/cosmos/ibc-go/v6/testing/mock"

	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	evmostypes "github.com/evmos/evmos/v12/types"
	"github.com/evmos/evmos/v12/utils"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

// ChainIDPrefix defines the default chain ID prefix for Evmos test chains
var ChainIDPrefix = "evmos_9000-"

func init() {
	ibcgotesting.ChainIDPrefix = ChainIDPrefix
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on
// counterparty chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the Coordinator in order to ensure synchrony between chains.
// Each update of any chain increments the block header time for all chains by 5 seconds.
func NewTestChain(t *testing.T, coord *ibcgotesting.Coordinator, chainID string) *ibcgotesting.TestChain {
	// generate validator private/public key
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := make(map[string]tmtypes.PrivValidator)
	signers[pubKey.Address().String()] = privVal

	// generate genesis account
	senderPrivKey, err := ethsecp256k1.GenerateKey()
	if err != nil {
		panic(err)
	}

	baseAcc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)

	acc := &evmostypes.EthAccount{
		BaseAccount: baseAcc,
		CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
	}

	amount := sdk.TokensFromConsensusPower(1, evmostypes.PowerReduction)

	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount)),
	}

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, chainID, balance)

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := app.GetTxConfig()

	// create an account to send transactions from
	chain := &ibcgotesting.TestChain{
		T:             t,
		Coordinator:   coord,
		ChainID:       chainID,
		App:           app,
		CurrentHeader: header,
		QueryServer:   app.GetIBCKeeper(),
		TxConfig:      txConfig,
		Codec:         app.AppCodec(),
		Vals:          valSet,
		Signers:       signers,
		SenderPrivKey: senderPrivKey,
		SenderAccount: acc,
		NextVals:      valSet,
	}

	coord.CommitBlock(chain)

	return chain
}

func NewTransferPath(chainA, chainB *ibcgotesting.TestChain) *Path {
	path := NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibcgotesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibcgotesting.TransferPort

	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
}
