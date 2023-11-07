package osmosis_test

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/ics20"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	commonnetwork "github.com/evmos/evmos/v15/testutil/integration/common/network"
	testutils "github.com/evmos/evmos/v15/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v15/testutil/integration/ibc/coordinator"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	// "github.com/evmos/evmos/v15/utils"
)

func (s *PrecompileTestSuite) TestSwap() {
	sender, senderAddr := s.keyring.GetAccAddr(0), s.keyring.GetAddr(0)
	acc, err := s.grpcHandler.GetAccount(sender.String())
	s.Require().NoError(err)

	coordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{s.unitNetwork},
	)

	// Account to sign IBC txs
	coordinator.SetDefaultSignerForChain(s.unitNetwork.GetChainID(), s.keyring.GetPrivKey(0), acc)
	dummyChainsIDs := coordinator.GetDummyChainsIds()
	coordinator.Setup(s.unitNetwork.GetChainID(), dummyChainsIDs[0])

	receiverOsmo := "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2"
	// receiverAtom := "cosmos1c2m73hdt6f37w9jqpqps5t3ha3st99dcsp7lf5"

	randomHex := common.HexToAddress("0x1FD55A1B9FC24967C4dB09C513C3BA0DFa7FF687")
	testSlippagePercentage := uint8(10)
	testWindowSeconds := uint64(20)
	transferAmount := big.NewInt(1e18)
	gas := uint64(2000)
	randomAddress := utiltx.GenerateAddress()

	method := s.precompile.Methods[osmosis.SwapMethod]
	testCases := []struct {
		name        string
		sender      common.Address
		receiver    string
		malleate    func() []interface{}
		expError    bool
		errContains string
	}{
		{
			name:     "fail - invalid number of args",
			sender:   senderAddr,
			receiver: receiverOsmo,
			malleate: func() []interface{} {
				return []interface{}{}
			},
			expError:    true,
			errContains: fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 7, 0),
		}, {
			name:     "fail - origin different from sender",
			sender:   senderAddr,
			receiver: receiverOsmo,
			malleate: func() []interface{} {
				return []interface{}{
					randomAddress,
					randomHex,
					randomHex,
					transferAmount,
					testSlippagePercentage,
					testWindowSeconds,
					receiverOsmo,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(ics20.ErrDifferentOriginFromSender, senderAddr, randomAddress),
		}, {
			name:     "fail - missing input token denom",
			sender:   senderAddr,
			receiver: receiverOsmo,
			malleate: func() []interface{} {
				// Register Evmos token as an erc20
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")
				evmosERC20 := evmosTokenPair.GetERC20Contract()

				// ibcOsmoDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				// osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, ibcOsmoDenomTrace)
				// s.Require().NoError(err, "expected no error during ibc osmo erc20 registration")
				// osmoERC20 := osmoTokenPair.GetERC20Contract()

				return []interface{}{
					randomAddress,
					evmosERC20,
					randomHex,
					transferAmount,
					testSlippagePercentage,
					testWindowSeconds,
					receiverOsmo,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(ics20.ErrDifferentOriginFromSender, senderAddr, randomAddress),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := coordinator.CommitAll()
			s.Require().NoError(err)

			contract := vm.NewContract(vm.AccountRef(tc.sender), s.precompile, big.NewInt(0), gas)

			stateDB := s.unitNetwork.GetStateDB()

			_, err = s.precompile.Swap(
				s.unitNetwork.GetContext(),
				tc.sender,
				stateDB,
				contract,
				&method,
				tc.malleate(),
			)
			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				// tc.postCheck()
			}
		})
	}
}
