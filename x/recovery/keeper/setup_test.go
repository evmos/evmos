package keeper_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	ibctesting "github.com/evmos/evmos/v12/ibc/testing"
	"github.com/evmos/evmos/v12/testutil"
	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/utils"
	feemarkettypes "github.com/evmos/evmos/v12/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcgotesting "github.com/cosmos/ibc-go/v6/testing"

	"github.com/evmos/evmos/v12/app"
	claimstypes "github.com/evmos/evmos/v12/x/claims/types"
	"github.com/evmos/evmos/v12/x/recovery/types"
)

var (
	ibcAtomDenom = "ibc/A4DB47A9D3CF9A068D454513891B526702455D3EF08FB9EB558C561F9DC2B701"
	ibcOsmoDenom = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	erc20Denom   = "erc20/0xdac17f958d2ee523a2206206994597c13d831ec7"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app         *app.Evmos
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(utiltx.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	header := testutil.NewHeader(
		1, time.Now().UTC(), "evmos_9001-1", consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.RecoveryKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	claimsParams := claimstypes.DefaultParams()
	claimsParams.AirdropStartTime = suite.ctx.BlockTime()
	err := suite.app.ClaimsKeeper.SetParams(suite.ctx, claimsParams)
	suite.Require().NoError(err)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	EvmosChain      *ibcgotesting.TestChain
	IBCOsmosisChain *ibcgotesting.TestChain
	IBCCosmosChain  *ibcgotesting.TestChain

	pathOsmosisEvmos  *ibctesting.Path
	pathCosmosEvmos   *ibctesting.Path
	pathOsmosisCosmos *ibctesting.Path
}

var s *IBCTestingSuite

func TestIBCTestingSuite(t *testing.T) {
	s = new(IBCTestingSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}
