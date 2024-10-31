package werc20_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/evmos/evmos/v20/precompiles/werc20"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"

	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

var s *PrecompileUnitTestSuite

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
	s = new(PrecompileUnitTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileUnitTestSuite) SetupTest() {
	keyring := keyring.New(2)

	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	s.network = integrationNetwork
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring

	chainID := s.network.GetChainID()
	s.precompileAddrHex = erc20types.GetWEVMOSContractHex(chainID)

	ctx := integrationNetwork.GetContext()

	tokenPairID := s.network.App.Erc20Keeper.GetTokenPairID(ctx, evmtypes.GetEVMCoinDenom())
	tokenPair, found := s.network.App.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
	s.Require().True(found, "expected wevmos precompile to be registered in the tokens map")
	s.Require().Equal(tokenPair.Erc20Address, s.precompileAddrHex)

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

func (s *PrecompileUnitTestSuite) TestEmitDepositEvent() {
	testCasees := []struct {
		name string
	}{
		{
			name: "pass",
		},
	}

	for _, tc := range testCasees {

		s.SetupTest()
		s.Run(tc.name, func() {
			caller := s.keyring.GetAddr(0)
			amount := new(big.Int).SetInt64(1_000)

			stateDB := s.network.GetStateDB()

			err := s.precompile.EmitDepositEvent(
				s.network.GetContext(),
				stateDB,
				caller,
				amount,
			)
			s.Require().NoError(err, "expected transfer event to be emitted successfully")

			// log := stateDB.Logs()[0]
			// s.Require().Equal(log.Address, s.precompile.Address())
			//
			// event := s.precompile.ABI.Events[werc20.EventTypeDeposit]
			//
			// // First topic should match the event signature.
			// s.Require().Equal(
			// 	crypto.Keccak256Hash([]byte(event.Sig)),
			// 	common.HexToHash(string(log.Topics[0].Hex())),
			// )
			// s.Require().Equal(
			// 	crypto.Keccak256Hash([]byte(caller.String())),
			// 	common.HexToHash(string(log.Topics[1].Hex())),
			// )
			// s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

			// // Verify data
			// var transferEvent werc20.EventTypeDeposit
			// err = cmn.UnpackLog(s.precompile.ABI, &transferEvent, erc20precompile.EventTypeTransfer, *log)
			// s.Require().NoError(err, "unable to unpack log into transfer event")
			//
			// s.Require().Equal(tc.from, transferEvent.From, "expected different from address")
			// s.Require().Equal(tc.to, transferEvent.To, "expected different to address")
			// s.Require().Equal(tc.amount, transferEvent.Value, "expected different amount")
		})
	}
}
