// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/werc20"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/utils"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

type PrecompileUnitTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     keyring.Keyring

	// WEVMOS related fields
	precompile        *werc20.Precompile
	precompileAddrHex string
}

func TestPrecompileUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileUnitTestSuite))
}

// SetupTest allows to configure the testing suite embedding a network with a
// custom chainID. This is important to check that the correct address is used
// for the precompile.
func (s *PrecompileUnitTestSuite) SetupTest(chainID string) {
	keyring := keyring.New(2)

	integrationNetwork := network.NewUnitTestNetwork(
		network.WithChainID(chainID),
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	s.network = integrationNetwork
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring

	s.precompileAddrHex = erc20types.GetWEVMOSContractHex(chainID)

	ctx := integrationNetwork.GetContext()

	tokenPairID := s.network.App.Erc20Keeper.GetTokenPairID(ctx, evmtypes.GetEVMCoinDenom())
	tokenPair, found := s.network.App.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
	s.Require().True(found, "expected wevmos precompile to be registered in the tokens map")
	s.Require().Equal(s.precompileAddrHex, tokenPair.Erc20Address, "expected a different address of the contract")

	precompile, err := werc20.NewPrecompile(
		tokenPair,
		s.network.App.BankKeeper,
		s.network.App.AuthzKeeper,
		s.network.App.TransferKeeper,
	)
	s.Require().NoError(err, "failed to instantiate the werc20 precompile")
	s.Require().NotNil(precompile)
	s.precompile = precompile
}

type DepositEvent struct {
	Dst common.Address
	Wad *big.Int
}

type WithdrawalEvent struct {
	Src common.Address
	Wad *big.Int
}

//nolint:dupl
func (s *PrecompileUnitTestSuite) TestEmitDepositEvent() {
	testCases := []struct {
		name    string
		chainID string
	}{
		{
			name:    "mainnet",
			chainID: utils.MainnetChainID + "-1",
		}, {
			name:    "testnet",
			chainID: utils.TestnetChainID + "-1",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest(tc.chainID)
			caller := s.keyring.GetAddr(0)
			amount := new(big.Int).SetInt64(1_000)

			stateDB := s.network.GetStateDB()

			err := s.precompile.EmitDepositEvent(
				s.network.GetContext(),
				stateDB,
				caller,
				amount,
			)
			s.Require().NoError(err, "expected deposit event to be emitted successfully")

			log := stateDB.Logs()[0]

			// Check on the address
			s.Require().Equal(log.Address, s.precompile.Address())

			// Check on the topics
			event := s.precompile.ABI.Events[werc20.EventTypeDeposit]
			s.Require().Equal(
				crypto.Keccak256Hash([]byte(event.Sig)),
				common.HexToHash(log.Topics[0].Hex()),
			)
			var adddressTopic common.Hash
			copy(adddressTopic[common.HashLength-common.AddressLength:], caller[:])
			s.Require().Equal(adddressTopic, log.Topics[1])

			s.Require().EqualValues(log.BlockNumber, s.network.GetContext().BlockHeight())

			// Verify data
			var depositEvent DepositEvent
			err = cmn.UnpackLog(s.precompile.ABI, &depositEvent, werc20.EventTypeDeposit, *log)
			s.Require().NoError(err, "unable to unpack log into deposit event")

			s.Require().Equal(caller, depositEvent.Dst, "expected different destination address")
			s.Require().Equal(amount, depositEvent.Wad, "expected different amount")
		})
	}
}

//nolint:dupl
func (s *PrecompileUnitTestSuite) TestEmitWithdrawalEvent() {
	testCases := []struct {
		name    string
		chainID string
	}{
		{
			name:    "mainnet",
			chainID: utils.MainnetChainID + "-1",
		}, {
			name:    "testnet",
			chainID: utils.TestnetChainID + "-1",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest(tc.chainID)
			caller := s.keyring.GetAddr(0)
			amount := new(big.Int).SetInt64(1_000)

			stateDB := s.network.GetStateDB()

			err := s.precompile.EmitWithdrawalEvent(
				s.network.GetContext(),
				stateDB,
				caller,
				amount,
			)
			s.Require().NoError(err, "expected withdrawal event to be emitted successfully")

			log := stateDB.Logs()[0]

			// Check on the address
			s.Require().Equal(log.Address, s.precompile.Address())

			// Check on the topics
			event := s.precompile.ABI.Events[werc20.EventTypeWithdrawal]
			s.Require().Equal(
				crypto.Keccak256Hash([]byte(event.Sig)),
				common.HexToHash(log.Topics[0].Hex()),
			)
			var adddressTopic common.Hash
			copy(adddressTopic[common.HashLength-common.AddressLength:], caller[:])
			s.Require().Equal(adddressTopic, log.Topics[1])

			s.Require().EqualValues(log.BlockNumber, s.network.GetContext().BlockHeight())

			// Verify data
			var withdrawalEvent WithdrawalEvent
			err = cmn.UnpackLog(s.precompile.ABI, &withdrawalEvent, werc20.EventTypeWithdrawal, *log)
			s.Require().NoError(err, "unable to unpack log into withdrawal event")

			s.Require().Equal(caller, withdrawalEvent.Src, "expected different source address")
			s.Require().Equal(amount, withdrawalEvent.Wad, "expected different amount")
		})
	}
}
