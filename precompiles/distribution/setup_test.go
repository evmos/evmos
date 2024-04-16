package distribution_test

import (
	"testing"

	"github.com/evmos/evmos/v18/precompiles/distribution"
	"github.com/evmos/evmos/v18/x/evm/statedb"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmosapp "github.com/evmos/evmos/v18/app"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	app        *evmosapp.Evmos
	address    common.Address
	validators []stakingtypes.Validator
	valSet     *tmtypes.ValidatorSet
	ethSigner  ethtypes.Signer
	privKey    cryptotypes.PrivKey
	signer     keyring.Signer
	bondDenom  string

	precompile *distribution.Precompile
	stateDB    *statedb.StateDB

	queryClientEVM evmtypes.QueryClient
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Precompile Suite")
}

func (s *PrecompileTestSuite) SetupTest() {
	s.DoSetupTest()
}
