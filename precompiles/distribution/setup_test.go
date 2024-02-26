package distribution_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	infltypes "github.com/evmos/evmos/v16/x/inflation/v1/types"

	"github.com/evmos/evmos/v16/precompiles/distribution"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile     *distribution.Precompile
	bondDenom      string
	validatorsKeys []testkeyring.Key
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	// TODO uncomment this
	// suite.Run(t, new(PrecompileTestSuite))

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Precompile Suite")
}

func (s *PrecompileTestSuite) SetupTest() {
	// enable inflation for staking rewards
	customGen := network.CustomGenesisState{}
	customGen[infltypes.ModuleName] = infltypes.DefaultGenesisState()

	keyring := testkeyring.New(2)
	s.validatorsKeys = generateKeys(3)

	operatorsAddr := make([]sdk.AccAddress, 3)
	for i, k := range s.validatorsKeys {
		operatorsAddr[i] = k.AccAddr
	}

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGen),
		network.WithValidatorOperators(operatorsAddr),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	ctx := nw.GetContext()
	sk := nw.App.StakingKeeper
	bondDenom, err := sk.BondDenom(ctx)
	if err != nil {
		panic(err)
	}

	s.bondDenom = bondDenom
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = nw
	s.precompile, err = distribution.NewPrecompile(
		s.network.App.DistrKeeper,
		s.network.App.StakingKeeper,
		s.network.App.AuthzKeeper,
	)
	if err != nil {
		panic(err)
	}
}
