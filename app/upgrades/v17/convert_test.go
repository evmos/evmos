package v17_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	v17 "github.com/evmos/evmos/v18/app/upgrades/v17"
	"github.com/evmos/evmos/v18/contracts"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	grpchandler "github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v18/testutil/tx"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

type WithdrawWEVMOSTestSuite struct {
	keyring testkeyring.Keyring
	network *network.UnitTestNetwork
	factory testfactory.TxFactory

	// account is the address of the account to withdraw WEVMOS from.
	account common.Address
	// wevmosContract is the address of the WEVMOS contract.
	wevmosContract common.Address
}

func TestWithdrawWEVMOS(t *testing.T) {
	t.Parallel()

	senderIdx := 0
	kr := testkeyring.New(1)

	testcases := []struct {
		name        string
		malleate    func(suite *WithdrawWEVMOSTestSuite) error
		expPass     bool
		errContains string
		expAmount   math.Int
	}{
		{
			name: "pass - empty account",
			malleate: func(ts *WithdrawWEVMOSTestSuite) error {
				// Deploy WEVMOS contract
				wevmosAddr, err := ts.factory.DeployContract(
					ts.keyring.GetPrivKey(erc20Deployer),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
				)
				require.NoError(t, err, "failed to deploy WEVMOS contract")

				ts.account = common.Address{}
				ts.wevmosContract = wevmosAddr

				return nil
			},
			expPass:   true,
			expAmount: sdk.ZeroInt(),
		},
		{
			name: "pass - withdraw WEVMOS",
			malleate: func(ts *WithdrawWEVMOSTestSuite) error {
				// Deploy WEVMOS contract
				wevmosAddr, err := ts.factory.DeployContract(
					ts.keyring.GetPrivKey(erc20Deployer),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
				)
				require.NoError(t, err, "failed to deploy WEVMOS contract")

				// Send WEVMOS to account
				_, err = ts.factory.ExecuteEthTx(
					ts.keyring.GetPrivKey(senderIdx),
					evmtypes.EvmTxArgs{
						To:     &wevmosAddr,
						Amount: sentWEVMOS.BigInt(),
						// FIXME: the gas simulation is not working correctly - otherwise results in out of gas
						GasLimit: 100_000,
					},
				)
				require.NoError(t, err, "failed to send WEVMOS to account")

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				return nil
			},
			expPass:   true,
			expAmount: sentWEVMOS,
		},
		{
			name: "fail - no contract at address",
			malleate: func(ts *WithdrawWEVMOSTestSuite) error {
				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = testutiltx.GenerateAddress()

				return nil
			},
			errContains: "failed to get WEVMOS balance for",
		},
		{
			name: "fail - no withdraw method on contract",
			malleate: func(ts *WithdrawWEVMOSTestSuite) error {
				// Deploy a contract without a withdrawal method
				contractAddr, err := ts.factory.DeployContract(
					ts.keyring.GetPrivKey(senderIdx),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{
						Contract: contracts.ERC20MinterBurnerDecimalsContract,
						ConstructorArgs: []interface{}{
							"TestToken", "TT", uint8(18),
						},
					},
				)
				if err != nil {
					return errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
				}

				// Mint some tokens to the contract
				_, err = ts.factory.ExecuteContractCall(
					ts.keyring.GetPrivKey(senderIdx),
					evmtypes.EvmTxArgs{To: &contractAddr},
					testfactory.CallArgs{
						ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
						MethodName:  "mint",
						Args:        []interface{}{ts.keyring.GetAddr(senderIdx), mintAmount},
					},
				)
				if err != nil {
					return errorsmod.Wrap(err, "failed to mint ERC-20 tokens")
				}

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = contractAddr

				return nil
			},
			errContains: "execution reverted",
		},
	}

	for _, tc := range testcases {
		tc := tc // capture range variable (for parallel testing)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Set up a new network
			nw := network.NewUnitTestNetwork(
				network.WithPreFundedAccounts(kr.GetAllAccAddrs()...),
			)
			handler := grpchandler.NewIntegrationHandler(nw)
			txFactory := testfactory.New(nw, handler)

			ts := &WithdrawWEVMOSTestSuite{
				keyring: kr,
				network: nw,
				factory: txFactory,
			}

			err := tc.malleate(ts)
			require.NoError(t, err, "failed to malleate test suite")

			amount, res, err := v17.WithdrawWEVMOS(
				nw.GetContext(),
				ts.account,
				ts.wevmosContract,
				nw.App.Erc20Keeper,
			)

			if tc.expPass {
				require.NoError(t, err, "expected no error")
				require.Equal(t, tc.expAmount.String(), amount.String(), "expected different amount to be withdrawn")
				if res != nil {
					require.Empty(t, res.VmError, "expected no VM error")
				}
			} else {
				require.Error(t, err, "expected error but got none")
				require.ErrorContains(t, err, tc.errContains, "expected different error message")
			}
		})
	}
}
