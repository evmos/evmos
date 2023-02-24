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

	"github.com/evmos/evmos/v11/app"
	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	evm "github.com/evmos/evmos/v11/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v11/x/feemarket/types"
	"github.com/evmos/evmos/v11/x/revenue/v1/types"

	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app            *app.Evmos
	queryClient    types.QueryClient
	queryClientEvm evm.QueryClient
	address        common.Address
	signer         keyring.Signer
	ethSigner      ethtypes.Signer
	consAddress    sdk.ConsAddress
	validator      stakingtypes.Validator
	denom          string
}

var s *KeeperTestSuite

var (
	contract = utiltx.GenerateAddress()
	deployer = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	withdraw = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
)

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.SetupApp()
}
