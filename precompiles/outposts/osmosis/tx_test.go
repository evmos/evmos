package osmosis_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	// evmosutiltx "github.com/evmos/evmos/v15/testutil/tx"
	"github.com/evmos/evmos/v15/utils"
)

func (s *PrecompileTestSuite) TestSwap() {
	// s.SetupTest()

	method := s.precompile.Methods[osmosis.SwapMethod]

	bondDenom := s.network.App.StakingKeeper.BondDenom(s.network.GetContext())

	// Retrieve Evmos token information useful for the testing
	evmosDenomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), bondDenom)
	evmosTokenPair, _ := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), evmosDenomID)

	// Retrieve Osmo token information useful for the testing
	osmoIBCDenom := utils.ComputeIBCDenom(portID, channelID, osmosis.OsmosisDenom)
	osmoDenomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), osmoIBCDenom)
	osmoTokenPair, _ := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), osmoDenomID)

	sender := s.keyring.GetAddr(0)
	receiverOsmo := "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2"
	// receiverAtom := "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5"
	transferAmount := big.NewInt(3)

	gas := uint64(0)

	input := evmosTokenPair.GetERC20Contract()
	output := osmoTokenPair.GetERC20Contract()
	testSlippagePercentage := uint8(10)
	testWindowSeconds := uint64(20)

	testCases := []struct {
		name     string
		sender   common.Address
		input    common.Address
		output   common.Address
		amount   *big.Int
		receiver string
		args     []interface{}
	}{
		{
			name:     "pass - correct swap",
			sender:   sender,
			input:    evmosTokenPair.GetERC20Contract(),
			output:   osmoTokenPair.GetERC20Contract(),
			amount:   transferAmount,
			receiver: receiverOsmo,
			args: []interface{}{
				sender,
				input,
				output,
				transferAmount,
				testSlippagePercentage,
				testWindowSeconds,
				receiverOsmo,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.network.NextBlock()
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(tc.sender), s.precompile, big.NewInt(0), gas)

			_, err = s.precompile.Swap(
				s.chainA.GetContext(),
				sender,
				s.stateDB,
				contract,
				&method,
				tc.args,
			)
			s.Require().NoError(err)
		})
	}
}
