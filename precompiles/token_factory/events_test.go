package tokenfactory_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	tokenfactory "github.com/evmos/evmos/v18/precompiles/token_factory"
	"math/big"
)

func (s *PrecompileTestSuite) TestCreateERC20Event() {
	fromAddr := s.keyring.GetKey(0).Addr

	testCases := []struct {
		name          string
		tokenName     string
		symbol        string
		decimals      uint8
		initialSupply *big.Int
		expectedError bool
	}{
		{
			"pass - creates an ERC20 token factory token and emits the correct events",
			TokenName,
			TokenDenom,
			uint8(18),
			big.NewInt(1e18),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			account := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), s.precompile.Address().Bytes())
			if account == nil {
				account = s.network.App.AccountKeeper.NewAccountWithAddress(s.network.GetContext(), s.precompile.Address().Bytes())
			}

			address := crypto.CreateAddress(s.precompile.Address(), account.GetSequence())

			err := s.precompile.EmitCreateERC20Event(s.network.GetContext(), stateDB, fromAddr, address, tc.tokenName, tc.symbol, tc.decimals, tc.initialSupply)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				logs := stateDB.Logs()

				log := logs[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[tokenfactory.EventTypeCreateERC20]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var erc20CreatedEvent tokenfactory.EventERC20Created
				err := cmn.UnpackLog(s.precompile.ABI, &erc20CreatedEvent, tokenfactory.EventTypeCreateERC20, *log)
				s.Require().NoError(err, "unable to unpack log into ERC20 created event")

				s.Require().Equal(fromAddr, erc20CreatedEvent.Creator, "expected different creator")
				s.Require().Equal(tc.initialSupply, erc20CreatedEvent.InitialSupply, "expected different initial supply")
				s.Require().Equal(tc.tokenName, erc20CreatedEvent.Name, "expected different name")
				s.Require().Equal(tc.symbol, erc20CreatedEvent.Symbol, "expected different symbol")
				s.Require().Equal(tc.decimals, erc20CreatedEvent.Decimals, "expected different decimals")
			}
		})
	}
}

func (s *PrecompileTestSuite) TestMintEvent() {
	fromAddr := s.keyring.GetKey(0).Addr

	testCases := []struct {
		name          string
		amount        *big.Int
		expectedError bool
	}{
		{
			"pass - creates a Mint event",
			big.NewInt(1e18),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()
			err := s.precompile.EmitEventMint(s.network.GetContext(), stateDB, fromAddr, tc.amount)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				logs := stateDB.Logs()

				log := logs[0]
				s.Require().Equal(log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[tokenfactory.EventTypeMint]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight()))

				// Check the fully unpacked event matches the one emitted
				var erc20CreatedEvent tokenfactory.EventMint
				err := cmn.UnpackLog(s.precompile.ABI, &erc20CreatedEvent, tokenfactory.EventTypeMint, *log)
				s.Require().NoError(err, "unable to unpack log into Mint event")

				s.Require().Equal(fromAddr, erc20CreatedEvent.To, "expected different to address")
				s.Require().Equal(tc.amount, erc20CreatedEvent.Value, "expected different amount")
			}
		})
	}
}
