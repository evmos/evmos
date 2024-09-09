package distribution_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/evmos/evmos/v20/precompiles/distribution"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"

	"github.com/stretchr/testify/suite"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile           *distribution.Precompile
	bondDenom            string
	validatorsKeys       []testkeyring.Key
	withValidatorSlashes bool
}

func TestPrecompileUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	s.validatorsKeys = generateKeys(3)
	customGen := network.CustomGenesisState{}

	// set some slashing events for integration test
	distrGen := distrtypes.DefaultGenesisState()
	if s.withValidatorSlashes {
		distrGen.ValidatorSlashEvents = []distrtypes.ValidatorSlashEventRecord{
			{
				ValidatorAddress:    sdk.ValAddress(s.validatorsKeys[0].Addr.Bytes()).String(),
				Height:              0,
				Period:              1,
				ValidatorSlashEvent: distrtypes.NewValidatorSlashEvent(1, math.LegacyNewDecWithPrec(5, 2)),
			},
			{
				ValidatorAddress:    sdk.ValAddress(s.validatorsKeys[0].Addr.Bytes()).String(),
				Height:              1,
				Period:              1,
				ValidatorSlashEvent: distrtypes.NewValidatorSlashEvent(1, math.LegacyNewDecWithPrec(5, 2)),
			},
		}
	}
	customGen[distrtypes.ModuleName] = distrGen

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
