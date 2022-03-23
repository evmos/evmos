package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	"github.com/tharsis/ethermint/tests"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/tharsis/evmos/v3/app"
	claimstypes "github.com/tharsis/evmos/v3/x/claims/types"
	"github.com/tharsis/evmos/v3/x/withdraw/types"
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
	consAddress := sdk.ConsAddress(tests.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9000-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.WithdrawKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	claimsParams := claimstypes.DefaultParams()
	claimsParams.AirdropStartTime = suite.ctx.BlockTime()
	suite.app.ClaimsKeeper.SetParams(suite.ctx, claimsParams)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = claimsParams.GetClaimsDenom()
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestGetIBCDenomDestinationIdentifiers() {
	address := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()

	testCases := []struct {
		name                                      string
		denom                                     string
		malleate                                  func()
		expError                                  bool
		expDestinationPort, expDestinationChannel string
	}{
		{
			"invalid native denom",
			"aevmos",
			func() {},
			true,
			"", "",
		},
		{
			"invalid IBC denom hash",
			"ibc/aevmos",
			func() {},
			true,
			"", "",
		},
		{
			"denom trace not found",
			"ibc/A4DB47A9D3CF9A068D454513891B526702455D3EF08FB9EB558C561F9DC2B701",
			func() {},
			true,
			"", "",
		},
		{
			"channel not found",
			"ibc/A4DB47A9D3CF9A068D454513891B526702455D3EF08FB9EB558C561F9DC2B701",
			func() {
				denomTrace := transfertypes.DenomTrace{
					Path:      "transfer/channel-3",
					BaseDenom: "uatom",
				}
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, denomTrace)
			},
			true,
			"", "",
		},
		{
			"success - ATOM",
			ibcAtomDenom,
			func() {
				denomTrace := transfertypes.DenomTrace{
					Path:      "transfer/channel-3",
					BaseDenom: "uatom",
				}
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, denomTrace)

				channel := channeltypes.Channel{
					Counterparty: channeltypes.NewCounterparty("transfer", "channel-292"),
				}
				suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, "transfer", "channel-3", channel)
			},
			false,
			"transfer", "channel-3",
		},
		{
			"success - OSMO",
			ibcOsmoDenom,
			func() {
				denomTrace := transfertypes.DenomTrace{
					Path:      "transfer/channel-0",
					BaseDenom: "uosmo",
				}
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, denomTrace)

				channel := channeltypes.Channel{
					Counterparty: channeltypes.NewCounterparty("transfer", "channel-204"),
				}
				suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, "transfer", "channel-0", channel)
			},
			false,
			"transfer", "channel-0",
		},
		{
			"success - ibcATOM (via Osmosis)",
			"ibc/6CDD4663F2F09CD62285E2D45891FC149A3568E316CE3EBBE201A71A78A69388",
			func() {
				denomTrace := transfertypes.DenomTrace{
					Path:      "transfer/channel-0/transfer/channel-0",
					BaseDenom: "uatom",
				}

				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, denomTrace)

				channel := channeltypes.Channel{
					Counterparty: channeltypes.NewCounterparty("transfer", "channel-204"),
				}
				suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, "transfer", "channel-0", channel)
			},
			false,
			"transfer", "channel-0",
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			destinationPort, destinationChannel, err := suite.app.WithdrawKeeper.GetIBCDenomDestinationIdentifiers(suite.ctx, tc.denom, address)
			if tc.expError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expDestinationPort, destinationPort)
				suite.Require().Equal(tc.expDestinationChannel, destinationChannel)
			}
		})
	}
}
