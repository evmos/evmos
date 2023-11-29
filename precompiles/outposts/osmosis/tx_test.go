package osmosis_test

import (
	"fmt"
	"math/big"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/precompiles/authorization"
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
	// Default variables used during tests.
	slippagePercentage := uint8(10)
	windowSeconds := uint64(20)
	transferAmount := big.NewInt(1e18)
	gas := uint64(2_000)
	senderAddress := utiltx.GenerateAddress()
	sender := sdktypes.AccAddress(senderAddress.Bytes())
	randomAddress := utiltx.GenerateAddress()
	receiver := "evmos1vl0x3xr0zwgrllhdzxxlkal7txnnk56q3552x7" //nolint:goconst

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
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				return []interface{}{}
			},
			expError:    true,
			errContains: fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 7, 0),
		}, {
			name:   "fail - origin different from sender",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				return []interface{}{
					randomAddress,
					randomAddress,
					randomAddress,
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(ics20.ErrDifferentOriginFromSender, senderAddress, randomAddress),
		}, {
			name:   "fail - missing input token denom",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddress,
					randomAddress,
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("token '%s' not registered", randomAddress),
		}, {
			name:   "fail - missing output token denom",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddress,
					evmosTokenPair.GetERC20Contract(),
					randomAddress,
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("token '%s' not registered", randomAddress),
		}, {
			name:   "fail - osmo token pair not registered (with osmo hardcoded address)",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddress,
					common.HexToAddress("0x1D54EcB8583Ca25895c512A8308389fFD581F9c9"),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("token '%s' not registered", common.HexToAddress("0x1D54EcB8583Ca25895c512A8308389fFD581F9c9")),
		}, {
			name:   "fail - osmo token pair registered with another channelID",
			sender: senderAddress,
			origin: senderAddress,
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
					senderAddress,
					wrongOsmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError: true,
			// Probably there is a better way than hardcoding the expected string
			errContains: fmt.Sprintf(osmosis.ErrInputTokenNotSupported, []string{"aevmos", "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"}),
		}, {
			name:   "fail - input equal to denom",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				return []interface{}{
					senderAddress,
					evmosTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(osmosis.ErrInputEqualOutput, "aevmos"),
		}, {
			name:   "fail - invalid input",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				wrongIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, "wrong")
				wrongTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, wrongIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddress,
					wrongTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError: true,
			// Probably there is a better way than hardcoding the expected string
			errContains: fmt.Sprintf(osmosis.ErrInputTokenNotSupported, []string{"aevmos", "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"}),
		}, {
			name:   "fail - receiver is not a valid bech32",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddress,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					"invalidbec32",
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(osmosis.ErrReceiverAddress, "not a valid evmos address"),
		}, {
			//  THIS PANICS INSIDE CheckAuthzExists
			name:   "fail - origin different from address caller",
			sender: senderAddress,
			origin: s.keyring.GetAddr(1),
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddress,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf(authorization.ErrAuthzDoesNotExistOrExpired, senderAddress, s.keyring.GetAddr(1)),
		}, {
			name:   "fail - ibc channel not open",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddress,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
				}
			},
			expError:    true,
			errContains: fmt.Sprintf("port ID (%s) channel ID (%s)", portID, channelID),
		}, {
			name:   "pass - correct swap",
			sender: senderAddress,
			origin: senderAddress,
			malleate: func() []interface{} {
				evmosTokenPair, err := testutils.RegisterEvmosERC20Coins(*s.unitNetwork, sender)
				s.Require().NoError(err, "expected no error during evmos erc20 registration")

				osmoIbcDenomTrace := utils.ComputeIBCDenomTrace(portID, channelID, osmosis.OsmosisDenom)
				osmoTokenPair, err := testutils.RegisterIBCERC20Coins(*s.unitNetwork, sender, osmoIbcDenomTrace)
				s.Require().NoError(err, "expected no error during ibc erc20 registration")

				return []interface{}{
					senderAddress,
					osmoTokenPair.GetERC20Contract(),
					evmosTokenPair.GetERC20Contract(),
					transferAmount,
					slippagePercentage,
					windowSeconds,
					receiver,
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
				ibcSender, ibcSenderPrivKey := s.keyring.GetAccAddr(0), s.keyring.GetPrivKey(0)
				// Account to sign IBC txs
				ibcAcc, err := s.grpcHandler.GetAccount(ibcSender.String())
				s.Require().NoError(err)

				coordinator := coordinator.NewIntegrationCoordinator(
					s.T(),
					[]commonnetwork.Network{s.unitNetwork},
				)

				coordinator.SetDefaultSignerForChain(s.unitNetwork.GetChainID(), ibcSenderPrivKey, ibcAcc)
				coordinator.Setup(s.unitNetwork.GetChainID(), coordinator.GetDummyChainsIds()[0])

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
			}
		})
	}
}
