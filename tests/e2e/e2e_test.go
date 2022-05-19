package e2e

import (
	"fmt"
	"io"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/tharsis/evmos/v3/tests/e2e/util"
)

// func (s *IntegrationTestSuite) Test001IBCTokenTransfer() {
// 	var ibcStakeDenom string

// 	s.Run("send_uosmo_to_chainB", func() {
// 		recipient := s.chains[1].Validators[0].PublicAddress
// 		token := sdk.NewInt64Coin(chain.OsmoDenom, chain.IbcSendAmount) // 3,300uosmo
// 		s.sendIBC(s.chains[0].ChainMeta.Id, s.chains[1].ChainMeta.Id, recipient, token)

// 		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chains[1].ChainMeta.Id][0].GetHostPort("1317/tcp"))

// 		// require the recipient account receives the IBC tokens (IBC packets ACKd)
// 		var (
// 			balances sdk.Coins
// 			err      error
// 		)
// 		s.Require().Eventually(
// 			func() bool {
// 				balances, err = queryBalances(chainBAPIEndpoint, recipient)
// 				s.Require().NoError(err)

// 				return balances.Len() == 3
// 			},
// 			time.Minute,
// 			5*time.Second,
// 		)

// 		for _, c := range balances {
// 			if strings.Contains(c.Denom, "ibc/") {
// 				ibcStakeDenom = c.Denom
// 				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
// 				break
// 			}
// 		}

// 		s.Require().NotEmpty(ibcStakeDenom)
// 	})
// }

func (s *IntegrationTestSuite) Test002QueryBalances() {
	// var (
	// 	expectedDenomsA   = []string{chain.OsmoDenom, chain.StakeDenom}
	// 	expectedDenomsB   = []string{chain.OsmoDenom, chain.StakeDenom, chain.IbcDenom}
	// 	expectedBalancesA = []uint64{chain.OsmoBalanceA - chain.IbcSendAmount, chain.StakeBalanceA - chain.StakeAmountA}
	// 	expectedBalancesB = []uint64{chain.OsmoBalanceB, chain.StakeBalanceB - chain.StakeAmountB, chain.IbcSendAmount}
	// )

	// chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chains[0].ChainMeta.Id][0].GetHostPort("1317/tcp"))
	// balancesA, err := queryBalances(chainAAPIEndpoint, s.chains[0].Validators[0].PublicAddress)
	// s.Require().NoError(err)
	// s.Require().NotNil(balancesA)
	// s.Require().Equal(2, len(balancesA))

	// chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chains[1].ChainMeta.Id][0].GetHostPort("1317/tcp"))
	// balancesB, err := queryBalances(chainBAPIEndpoint, s.chains[1].Validators[0].PublicAddress)
	// s.Require().NoError(err)
	// s.Require().NotNil(balancesB)
	// s.Require().Equal(3, len(balancesB))

	// actualDenomsA := make([]string, 0, 2)
	// actualBalancesA := make([]uint64, 0, 2)
	// actualDenomsB := make([]string, 0, 2)
	// actualBalancesB := make([]uint64, 0, 2)

	// for _, balanceA := range balancesA {
	// 	actualDenomsA = append(actualDenomsA, balanceA.GetDenom())
	// 	actualBalancesA = append(actualBalancesA, balanceA.Amount.Uint64())
	// }

	// for _, balanceB := range balancesB {
	// 	actualDenomsB = append(actualDenomsB, balanceB.GetDenom())
	// 	actualBalancesB = append(actualBalancesB, balanceB.Amount.Uint64())
	// }

	// s.Require().ElementsMatch(expectedDenomsA, actualDenomsA)
	// s.Require().ElementsMatch(expectedBalancesA, actualBalancesA)
	// s.Require().ElementsMatch(expectedDenomsB, actualDenomsB)
	// s.Require().ElementsMatch(expectedBalancesB, actualBalancesB)
}

func queryBalances(endpoint, addr string) (sdk.Coins, error) {
	path := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		endpoint, addr,
	)
	var err error
	var resp *http.Response
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return nil, fmt.Errorf("exceeded retry limit of %d with %d", retriesLeft, http.StatusServiceUnavailable)
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.GetBalances(), nil
}
