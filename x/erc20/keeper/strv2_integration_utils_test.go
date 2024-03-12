// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func SetupTest() *STRv2TrackingSuite {
	keyring := testkeyring.New(2)
	network := testnetwork.NewUnitTestNetwork(
		testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	handler := grpc.NewIntegrationHandler(network)
	factory := testfactory.New(network, handler)

	return &STRv2TrackingSuite{
		keyring: keyring,
		network: network,
		handler: handler,
		factory: factory,
	}
}

type ERC20ConstructorArgs struct {
	Name     string
	Symbol   string
	Decimals uint8
}

func (e ERC20ConstructorArgs) toInterface() []interface{} {
	return []interface{}{
		e.Name, e.Symbol, e.Decimals,
	}
}

func (s *STRv2TrackingSuite) DeployERC20Contract(
	deployer testkeyring.Key,
	cArgs ERC20ConstructorArgs,
) (common.Address, error) {
	addr, err := s.factory.DeployContract(
		deployer.Priv,
		evmtypes.EvmTxArgs{},
		testfactory.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: cArgs.toInterface(),
		},
	)
	if err != nil {
		return common.Address{}, errorsmod.Wrap(err, "failed to deploy ERC-20 contract")
	}

	return addr, nil
}

func (s *STRv2TrackingSuite) RegisterTokenPair(
	erc20Addr common.Address,
) error {
	_, err := s.network.App.Erc20Keeper.RegisterERC20(
		s.network.GetContext(),
		erc20Addr,
	)
	return err
}
