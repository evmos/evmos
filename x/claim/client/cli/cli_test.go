package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/app/params"
	"github.com/tharsis/evmos/x/claim/client/cli"
	"github.com/tharsis/evmos/x/claim/types"
	claimtypes "github.com/tharsis/evmos/x/claim/types"
)

var (
	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
)

func init() {
	params.SetAddressPrefixes()
	addr1 = ed25519.GenPrivKey().PubKey().Address().Bytes()
	addr2 = ed25519.GenPrivKey().PubKey().Address().Bytes()
}

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()

	genState := app.ModuleBasics.DefaultGenesis(s.cfg.Codec)
	claimGenState := claimtypes.DefaultGenesis()
	claimGenState.ModuleAccountBalance = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(30))
	claimGenState.ClaimRecords = []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)),
			ActionCompleted:        []bool{false, false, false, false},
		},
		{
			Address:                addr2.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 20)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	claimGenStateBz := s.cfg.Codec.MustMarshalJSON(claimGenState)
	genState[claimtypes.ModuleName] = claimGenStateBz

	s.cfg.GenesisState = genState
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// TODO: Figure out how to get genesis time from IntegrationTestSuite
// Because right now, verifying the correctness of the airdrop_start_time
// isn't possible.
// Other than that, this works

// func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name           string
// 		args           []string
// 		expectedOutput string
// 	}{
// 		{
// 			"json output",
// 			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
// 			`{"airdrop_start_time":"1970-01-01T00:00:00Z","duration_until_decay":"3600s","duration_of_decay":"18000s"}`,
// 		},
// 		{
// 			"text output",
// 			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
// 			`airdrop_start_time: "1970-01-01T00:00:00Z"
// duration_of_decay: 18000s
// duration_until_decay: 3600s`,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdQueryParams()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)
// 			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
// 		})
// 	}
// }

func (s *IntegrationTestSuite) TestCmdQueryClaimRecord() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
	}{
		{
			"query claim record",
			[]string{
				addr1.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClaimRecord()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.QueryClaimRecordResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
		})
	}
}

func (s *IntegrationTestSuite) TestCmdQueryClaimableForAction() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query claimable-for-action amount",
			[]string{
				addr2.String(),
				types.ActionAddLiquidity.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClaimableForAction()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.QueryClaimableForActionResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(result.Coins.String(), tc.coins.String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
