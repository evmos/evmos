package keeper_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/x/claims/types"
	evm "github.com/evmos/evmos/v12/x/evm/types"
	"github.com/stretchr/testify/suite"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app            *app.Evmos
	queryClient    types.QueryClient
	queryClientEvm evm.QueryClient
	address        common.Address
	signer         keyring.Signer
	ethSigner      ethtypes.Signer
	validator      stakingtypes.Validator
}

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}
