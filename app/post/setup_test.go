// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post_test

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v20/x/inflation/v1/types"

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

	from common.Address
	to   common.Address
}

func (s *PostTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)

	// TxBuilder is used to create Ethereum and Cosmos Tx to test
	// the fee burner.
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(codec, authtx.DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()

	s.from = keyring.GetAddr(0)
	s.to = keyring.GetAddr(1)
	s.unitNetwork = unitNetwork
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.txBuilder = txBuilder
}

func TestPostTestSuite(t *testing.T) {
	suite.Run(t, new(PostTestSuite))
}

func (s *PostTestSuite) BuildEthTx() sdk.Tx {
	chainID := evmtypes.GetChainConfig().ChainID

	nonce := s.unitNetwork.App.EvmKeeper.GetNonce(
		s.unitNetwork.GetContext(),
		common.BytesToAddress(s.from.Bytes()),
	)

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &s.to,
		GasLimit:  gasLimit,
		GasPrice:  big.NewInt(1),
		GasTipCap: big.NewInt(1),
	}

	msgEthereumTx := evmtypes.NewTx(ethTxParams)
	msgEthereumTx.From = s.from.String()
	tx, err := msgEthereumTx.BuildTx(s.txBuilder, "evmos")
	s.Require().NoError(err)
	return tx
}

// BuildCosmosTxWithNSendMsg is an utils function to create an sdk.Tx containing
// a single message of type MsgSend from the bank module.
func (s *PostTestSuite) BuildCosmosTxWithNSendMsg(n int, feeAmount sdk.Coins) sdk.Tx {
	messages := make([]sdk.Msg, n)

	sendMsg := banktypes.MsgSend{
		FromAddress: s.from.String(),
		ToAddress:   s.to.String(),
		Amount:      feeAmount,
	}

	for i := range messages {
		messages[i] = &sendMsg
	}

	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetFeeAmount(feeAmount)
	err := s.txBuilder.SetMsgs(messages...)
	s.Require().NoError(err)
	return s.txBuilder.GetTx()
}

// MintCoinsForFeeCollector allows to mint a specific amount of coins from the bank
// and to transfer them to the FeeCollector.
func (s *PostTestSuite) MintCoinsForFeeCollector(amount sdk.Coins) {
	// Minting tokens for the FeeCollector to simulate fee accrued.
	err := s.unitNetwork.App.BankKeeper.MintCoins(
		s.unitNetwork.GetContext(),
		inflationtypes.ModuleName,
		amount,
	)
	s.Require().NoError(err)

	err = s.unitNetwork.App.BankKeeper.SendCoinsFromModuleToModule(
		s.unitNetwork.GetContext(),
		inflationtypes.ModuleName,
		authtypes.FeeCollectorName,
		amount,
	)
	s.Require().NoError(err)

	balance := s.GetFeeCollectorBalance()
	s.Require().Equal(amount, balance)
}

// GetFeeCollectorBalance is an utility function to query the balance
// of the FeeCollector module.
func (s *PostTestSuite) GetFeeCollectorBalance() sdk.Coins {
	address := s.unitNetwork.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	balance := s.unitNetwork.App.BankKeeper.GetAllBalances(
		s.unitNetwork.GetContext(),
		address,
	)
	return balance
}
