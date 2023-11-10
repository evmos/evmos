// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post_test

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/suite"
)

const (
	gasLimit = 100_000
)

type PostTestSuite struct {
	suite.Suite

	unitNetwork *network.UnitTestNetwork
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	txBuilder client.TxBuilder
}

func (s *PostTestSuite) SetupTest() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	interfaceRegistry := types.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(codec, tx.DefaultSignModes)

	s.unitNetwork = unitNetwork
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.txBuilder = txConfig.NewTxBuilder()
}

func TestPostTestSuite(t *testing.T) {
	suite.Run(t, new(PostTestSuite))
}

func (s *PostTestSuite) BuildEthTx(from, to common.Address) sdk.Tx {
	chainID := s.unitNetwork.App.EvmKeeper.ChainID()
	nonce := s.unitNetwork.App.EvmKeeper.GetNonce(
		s.unitNetwork.GetContext(),
		common.BytesToAddress(from.Bytes()),
	)

	amount := big.NewInt(1)
	input := make([]byte, 0)
	gasPrice := big.NewInt(1)
	gasFeeCap := big.NewInt(1)
	gasTipCap := big.NewInt(1)
	accesses := &ethtypes.AccessList{}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &to,
		Amount:    amount,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Input:     input,
		Accesses:  accesses,
	}

	msgEthereumTx := evmtypes.NewTx(ethTxParams)
	msgEthereumTx.From = from.String()
	tx, err := msgEthereumTx.BuildTx(s.txBuilder, "aevmos")
	s.Require().NoError(err)
	return tx
}

func (s *PostTestSuite) BuildCosmosTx(from, to common.Address, feeAmount sdk.Coins) sdk.Tx {
	sendMsg := banktypes.MsgSend{
		FromAddress: from.String(),
		ToAddress:   to.String(),
		Amount:      feeAmount,
	}
	s.txBuilder.SetGasLimit(gasLimit)
	// s.txBuilder.SetFeeAmount(feeAmount)
	err := s.txBuilder.SetMsgs(&sendMsg)
	s.Require().NoError(err)
	return s.txBuilder.GetTx()
}
