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
	"github.com/evmos/evmos/v15/utils"
)

func (s *PrecompileTestSuite) TestSwap() {
	sender, senderAddr, senderPrivKey := s.keyring.GetAccAddr(0), s.keyring.GetAddr(0), s.keyring.GetPrivKey(0)
	// Account to sign IBC txs
	acc, err := s.grpcHandler.GetAccount(sender.String())
	s.Require().NoError(err)

	validSlippagePercentage := uint8(10)
	validWindowSeconds := uint64(20)
	transferAmount := big.NewInt(1e18)
	gas := uint64(2000)
	randomAddress := utiltx.GenerateAddress()
	osmoAddress := "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2"

	method := s.precompile.Methods[osmosis.SwapMethod]
	testCases := []struct {
		name        string
		sender      common.Address
		origin      common.Address
		malleate    func() []interface{}
		ibcSetup    bool
		expError    bool
		errContains string
	}{
		{
			name:   "fail - invalid number of args",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				return []interface{}{}
			},
			expError:    true,
			errContains: fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 7, 0),
		}, {
			name:   "fail - origin different from sender",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				return []interface{}{
					randomAddress,
					randomAddress,
					randomAddress,
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(ics20.ErrDifferentOriginFromSender, senderAddr, randomAddress),
		}, {
			name:   "fail - missing input token denom",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddr,
					randomAddress,
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("token '%s' not registered", randomAddress),
		}, {
			name:   "fail - missing output token denom",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddr,
					evmosTokenPair.GetERC20Contract(),
					randomAddress,
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("token '%s' not registered", randomAddress),
		}, {
			name:   "fail - osmo token pair not registered (with osmo hardcoded address)",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddr,
					common.HexToAddress("0x1D54EcB8583Ca25895c512A8308389fFD581F9c9"),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("token '%s' not registered", common.HexToAddress("0x1D54EcB8583Ca25895c512A8308389fFD581F9c9")),
		}, {
			name:   "fail - osmo token pair registered with another channelID",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				_, err = testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				wrongOsmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, "channel-1", osmosis.OsmosisDenom)
				wrongOsmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, wrongOsmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddr,
					wrongOsmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError: true,
			// Probably there is a better way than hardcoding the expected string
			errContains: fmt.Sprintf(osmosis.ErrInputTokenNotSupported, []string{"aevmos", "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"}),
		}, {
			name:   "fail - input equal to denom",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddr,
					evmosTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(osmosis.ErrInputEqualOutput),
		}, {
			name:   "fail - invalid input",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				wrongIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, "wrong")
				wrongTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, wrongIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddr,
					wrongTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError: true,
			// Probably there is a better way than hardcoding the expected string
			errContains: fmt.Sprintf(osmosis.ErrInputTokenNotSupported, []string{"aevmos", "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"}),
		}, {
			name:   "fail - receiver is not a valid bech32",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddr,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					"invalidbec32",
				}
			},
			expError:    true,
			errContains: "invalid separator",
		}, {
			//  THIS PANICS INSIDE CheckAuthzExists
			// 	name:   "fail - origin different from address caller",
			// 	sender: senderAddr,
			// 	origin: s.keyring.GetAddr(1),
			// 	malleate: func() []interface{} {
			// 		evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
			// 		s.Require().NoError(err, "expected no error during evmos erc20 registration")
			//
			// 		osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
			// 		osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
			// 		s.Require().NoError(err, "expected no error during ibc erc20 registration")
			//
			// 		return []interface{}{
			// 			senderAddr,
			// 			osmoTokenPair.GetERC20Contract(),
			// 			evmosTokenPair.GetERC20Contract(),
			// 			transferAmount,
			// 			validSlippagePercentage,
			// 			validWindowSeconds,
			// 			osmoAddress,
			// 		}
			// 	},
			// 	expError:    true,
			// 	errContains: "invalid separator",
			// }, {
			name:   "fail - ibc channel not open",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddr,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("port ID (%s) channel ID (%s)", portID, channelID),
		}, {
			name:   "pass - correct swap",
			sender: senderAddr,
			origin: senderAddr,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddr,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					validSlippagePercentage,
					validWindowSeconds,
					osmoAddress,
				}
			},
			expError: false,
			ibcSetup: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			contract := vm.NewContract(vm.AccountRef(tc.sender), s.precompile, big.NewInt(0), gas)

			stateDB := s.unitNetwork.GetStateDB()

			if tc.ibcSetup {
				coordinator := coordinator.NewIntegrationCoordinator(
					s.T(),
					[]commonnetwork.Network{s.unitNetwork},
				)

				coordinator.SetDefaultSignerForChain(s.unitNetwork.GetChainID(), senderPrivKey, acc)
				dummyChainsIDs := coordinator.GetDummyChainsIds()
				coordinator.Setup(s.unitNetwork.GetChainID(), dummyChainsIDs[0])

				err = coordinator.CommitAll()
				s.Require().NoError(err)
			}

			_, err := s.precompile.Swap(
				s.unitNetwork.GetContext(),
				tc.origin,
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
