package keeper_test

import (
	"testing"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evm "github.com/evmos/evmos/v15/x/evm/types"

	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/x/incentives/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx              sdk.Context
	app              *app.Evmos
	queryClientEvm   evm.QueryClient
	queryClient      types.QueryClient
	address          common.Address
	consAddress      sdk.ConsAddress
	clientCtx        client.Context
	ethSigner        ethtypes.Signer
	priv             cryptotypes.PrivKey
	signer           keyring.Signer
	mintFeeCollector bool
}

var s *KeeperTestSuite

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
