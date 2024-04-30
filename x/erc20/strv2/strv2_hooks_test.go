package strv2_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	grpchandler "github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

type STRV2WEVMOSHooksTestSuite struct {
	keyring testkeyring.Keyring
	network *network.UnitTestNetwork
	factory testfactory.TxFactory

	// account is the address of the account to withdraw WEVMOS from.
	account common.Address
	// wevmosContract is the address of the WEVMOS contract.
	wevmosContract common.Address
}

const (

	// erc20Deployer is the index for the account that deploys the ERC-20 contract.
	erc20Deployer = 0
)

// sentWEVMOS is the amount of WEVMOS sent to the WEVMOS contract during testing.
var sentWEVMOS = sdk.NewInt(1e18)

func TestDepositWEVMOS(t *testing.T) {
	t.Parallel()

	senderIdx := 0
	kr := testkeyring.New(1)

	testcases := []struct {
		name        string
		malleate    func(suite *STRV2WEVMOSHooksTestSuite) error
		expFound    bool
		errContains string
		chainID     string
	}{
		{
			name: "found - deposit WEVMOS",
			malleate: func(ts *STRV2WEVMOSHooksTestSuite) error {
				// Deploy WEVMOS contract
				wevmosAddr, err := ts.factory.DeployContract(
					ts.keyring.GetPrivKey(erc20Deployer),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
				)
				require.NoError(t, err, "failed to deploy WEVMOS contract")

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				require.False(t, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))

				return nil
			},
			expFound: true,
			chainID:  utils.TestingChainID + "-1",
		},
		{
			name: "found - should not fail if address already there while deposit WEVMOS",
			malleate: func(ts *STRV2WEVMOSHooksTestSuite) error {
				// Deploy WEVMOS contract
				wevmosAddr, err := ts.factory.DeployContract(
					ts.keyring.GetPrivKey(erc20Deployer),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
				)
				require.NoError(t, err, "failed to deploy WEVMOS contract")

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				ts.network.App.Erc20Keeper.SetSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0))

				return nil
			},
			expFound: true,
			chainID:  utils.TestingChainID + "-1",
		},
		{
			name: "not found - should not register since its not expected contract",
			malleate: func(ts *STRV2WEVMOSHooksTestSuite) error {
				// Deploy WEVMOS contract
				wevmosAddr, err := ts.factory.DeployContract(
					ts.keyring.GetPrivKey(erc20Deployer),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{Contract: contracts.WEVMOSContract},
				)
				require.NoError(t, err, "failed to deploy WEVMOS contract")

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				return nil
			},
			expFound: false,
			chainID:  utils.TestnetChainID + "-1",
		},
	}

	for _, tc := range testcases {
		tc := tc // capture range variable (for parallel testing)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Set up a new network
			nw := network.NewUnitTestNetwork(
				network.WithChainID(tc.chainID),
				network.WithPreFundedAccounts(kr.GetAllAccAddrs()...),
			)
			handler := grpchandler.NewIntegrationHandler(nw)
			txFactory := testfactory.New(nw, handler)

			ts := &STRV2WEVMOSHooksTestSuite{
				keyring: kr,
				network: nw,
				factory: txFactory,
			}

			err := tc.malleate(ts)
			require.NoError(t, err, "failed to malleate test suite")

			// Send WEVMOS to account
			_, err = ts.factory.ExecuteEthTx(
				ts.keyring.GetPrivKey(senderIdx),
				evmtypes.EvmTxArgs{
					To:     &ts.wevmosContract,
					Amount: sentWEVMOS.BigInt(),
					// FIXME: the gas simulation is not working correctly - otherwise results in out of gas
					GasLimit: 100_000,
				},
			)
			require.NoError(t, err, "failed to send WEVMOS to account")

			require.Equal(t, tc.expFound, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))
		})
	}
}

func TestWithdrawWEVMOS(t *testing.T) {
	t.Parallel()

	senderIdx := 0
	kr := testkeyring.New(1)

	testcases := []struct {
		name        string
		malleate    func(suite *STRV2WEVMOSHooksTestSuite) error
		expFound    bool
		errContains string
		chainID     string
	}{
		{
			name: "found - withdraw WEVMOS",
			malleate: func(ts *STRV2WEVMOSHooksTestSuite) error {
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

				// Address was added after deposit, delete the entry to test withdraw
				require.True(t, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))
				ts.network.App.Erc20Keeper.DeleteSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0))
				require.False(t, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				return nil
			},
			expFound: true,
			chainID:  utils.TestingChainID + "-1",
		},
		{ //nolint:dupl
			name: "found - with address already registered - withdraw WEVMOS",
			malleate: func(ts *STRV2WEVMOSHooksTestSuite) error {
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

				require.True(t, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				return nil
			},
			expFound: true,
			chainID:  utils.TestingChainID + "-1",
		},
		{ //nolint:dupl
			name: "not found - wrong contract - withdraw WEVMOS",
			malleate: func(ts *STRV2WEVMOSHooksTestSuite) error {
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

				require.True(t, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))

				ts.account = kr.GetAddr(senderIdx)
				ts.wevmosContract = wevmosAddr

				return nil
			},
			expFound: false,
			chainID:  utils.TestnetChainID + "-1",
		},
	}

	for _, tc := range testcases {
		tc := tc // capture range variable (for parallel testing)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Set up a new network
			nw := network.NewUnitTestNetwork(
				network.WithChainID(utils.TestingChainID+"-1"),
				network.WithPreFundedAccounts(kr.GetAllAccAddrs()...),
			)
			handler := grpchandler.NewIntegrationHandler(nw)
			txFactory := testfactory.New(nw, handler)

			ts := &STRV2WEVMOSHooksTestSuite{
				keyring: kr,
				network: nw,
				factory: txFactory,
			}

			err := tc.malleate(ts)
			require.NoError(t, err, "failed to malleate test suite")

			_, err = ts.factory.ExecuteContractCall(
				ts.keyring.GetPrivKey(0),
				evmtypes.EvmTxArgs{
					To: &ts.wevmosContract,
				},
				testfactory.CallArgs{
					ContractABI: contracts.WEVMOSContract.ABI,
					MethodName:  "withdraw",
					Args: []interface{}{
						transferAmount,
					},
				},
			)
			require.NoError(t, err)
			require.True(t, ts.network.App.Erc20Keeper.HasSTRv2Address(ts.network.GetContext(), ts.keyring.GetAccAddr(0)))
		})
	}
}
