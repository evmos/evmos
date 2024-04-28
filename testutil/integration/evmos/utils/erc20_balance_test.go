package utils_test

import (
	"math/big"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/contracts"
	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	testfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	testutils "github.com/evmos/evmos/v18/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestGetERC20Balance(t *testing.T) {
	keyring := testkeyring.New(2)
	network := testnetwork.New(
		testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	handler := grpc.NewIntegrationHandler(network)
	factory := testfactory.New(network, handler)

	mintedBalance := big.NewInt(1000)
	keyWithoutFunds, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err, "failed to generate dummy key")

	// Deploy ERC20 contract
	deployer := keyring.GetKey(0)
	erc20Addr, err := factory.DeployContract(
		deployer.Priv,
		evmtypes.EvmTxArgs{}, // NOTE: using default values by passing empty struct
		testfactory.ContractDeploymentData{
			Contract: contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{
				"TestToken", "TT", uint8(6),
			},
		},
	)
	require.NoError(t, err, "failed to deploy ERC20 contract")

	testcases := []struct {
		name         string
		priv         cryptotypes.PrivKey
		contractAddr common.Address
		malleate     func() error
		expBalance   *big.Int
		expPass      bool
		errContains  string
	}{
		{
			name:         "pass - empty balance",
			priv:         keyring.GetPrivKey(1),
			contractAddr: erc20Addr,
			expBalance:   big.NewInt(0),
			expPass:      true,
		},
		{
			name:         "pass - non-empty balance",
			priv:         keyring.GetPrivKey(0),
			contractAddr: erc20Addr,
			malleate: func() error {
				_, err := factory.ExecuteContractCall(
					deployer.Priv,
					evmtypes.EvmTxArgs{
						To: &erc20Addr,
					},
					testfactory.CallArgs{
						ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
						MethodName:  "mint",
						Args: []interface{}{
							keyring.GetAddr(0),
							mintedBalance,
						},
					},
				)
				return err
			},
			expBalance: mintedBalance,
			expPass:    true,
		},
		{
			name:         "fail - wrong contract",
			priv:         keyring.GetPrivKey(0),
			contractAddr: utiltx.GenerateAddress(),
			errContains:  "got empty return value from contract call",
		},
		{
			name:         "fail - sender has no balance",
			priv:         keyWithoutFunds,
			contractAddr: erc20Addr,
			errContains:  "failed to execute contract call",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.malleate != nil {
				err = tc.malleate()
				require.NoError(t, err, "failed to malleate")
			}

			balance, err := testutils.GetERC20Balance(
				factory,
				tc.priv,
				tc.contractAddr,
			)
			if tc.expPass {
				require.NoError(t, err, "failed to get ERC20 balance")
				require.Equal(t, tc.expBalance.String(), balance.String(), "unexpected balance")
			} else {
				// NOTE: this is to be able to test errors potentially going forward.
				// With the existing integration test setup, there are no errors to be forced.
				require.Error(t, err, "expected error")
				require.ErrorContains(t, err, tc.errContains, "unexpected error")
			}
		})
	}
}
