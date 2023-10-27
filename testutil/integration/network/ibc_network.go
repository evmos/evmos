package network

import (
	"fmt"
	"strconv"
	"testing"

	tmtypes "github.com/cometbft/cometbft/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/core/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/evmos/evmos/v15/utils"
)

type MultiZoneSetup interface {
	GetMainNetwork() EvmosIBCNetwork
	GetNetworkByIndex(index int) EvmosIBCNetwork
	GetNetworkByChainID(chainID string) EvmosIBCNetwork

	GetCoordinator() Coordinator
	GetEVMNetworks() []EvmosIBCNetwork
}

type EvmosIBCNetwork interface {
	EvmosNetwork
	GetIBCQueryServer() ibctypes.QueryServer
	GetSenderAccount() ibctesting.SenderAccount
	GetSenderAccounts() []ibctesting.SenderAccount

	getIBCChain(coord *ibctesting.Coordinator) *ibctesting.TestChain
}

type IntegrationEvmosIBCNetwork struct {
	IntegrationNetwork
	prefundedAccounts []ibctesting.SenderAccount
	valSigners        map[string]tmtypes.PrivValidator
}

var _ EvmosIBCNetwork = (*IntegrationEvmosIBCNetwork)(nil)

func NewEvmosIBCNetwork(opts ...ConfigOption) *IntegrationEvmosIBCNetwork {
	network := New(opts...)
	return &IntegrationEvmosIBCNetwork{
		IntegrationNetwork: *network,
	}
}

func (n *IntegrationEvmosIBCNetwork) GetIBCQueryServer() ibctypes.QueryServer {
	return nil
}

func (n *IntegrationEvmosIBCNetwork) GetSenderAccount() ibctesting.SenderAccount {
	return ibctesting.SenderAccount{}
}

func (n *IntegrationEvmosIBCNetwork) GetSenderAccounts() []ibctesting.SenderAccount {
	return nil
}

func (n *IntegrationEvmosIBCNetwork) getIBCChain(coord *ibctesting.Coordinator) *ibctesting.TestChain {
	// create an account to send transactions from
	t := &testing.T{}
	defaultAcct := n.prefundedAccounts[0]
	return &ibctesting.TestChain{
		T:              t,
		Coordinator:    coord,
		ChainID:        n.GetChainID(),
		App:            n.app,
		CurrentHeader:  n.ctx.BlockHeader(),
		QueryServer:    n.GetIBCQueryServer(),
		TxConfig:       n.app.GetTxConfig(),
		Codec:          n.app.AppCodec(),
		Vals:           n.GetValSet(),
		NextVals:       n.GetValSet(),
		Signers:        n.valSigners,
		SenderPrivKey:  defaultAcct.SenderPrivKey,
		SenderAccount:  defaultAcct.SenderAccount,
		SenderAccounts: n.prefundedAccounts,
	}
}

type IntegrationIBCNetwork struct {
	cfg                 IBCNetworkConfig
	defaultEvmosNetwork EvmosIBCNetwork
	coordinator         Coordinator
}

func NewIBCNetwork() *IntegrationIBCNetwork {
	cfg := DefaultIBCNetworkConfig()
	evmosNetwork := NewEvmosIBCNetwork()

	ibcNetwork := &IntegrationIBCNetwork{
		cfg:                 cfg,
		defaultEvmosNetwork: evmosNetwork,
	}

	err := ibcNetwork.configureAndInitChains()
	if err != nil {
		panic(err)
	}
	return ibcNetwork
}

func (n *IntegrationIBCNetwork) configureAndInitChains() error {
	// Create evmos chains
	evmosChains := createEvmosIBCChains(n.cfg)
	// Create Coordinator
	coordinator := newCoordinator(n.cfg)
	// Set evmos chains within the coordinator
	n.coordinator.SetEvmosChains(evmosChains)

	n.coordinator = coordinator
	n.defaultEvmosNetwork = evmosChains[0]
	return nil
}

func createEvmosIBCChains(cfg IBCNetworkConfig) []EvmosIBCNetwork {
	chains := make([]EvmosIBCNetwork, cfg.numberOfChains)
	for i := 1; i <= cfg.numberOfChains; i++ {
		chainID := generateEVMChainID(i)
		chain := NewEvmosIBCNetwork(
			WithChainID(chainID),
		)
		chains[i] = chain
	}
	return chains
}

func generateEVMChainID(index int) string {
	return fmt.Sprintf("%v-%v", utils.TestnetChainID, strconv.Itoa(index))
}
