package e2e

import (
	"fmt"
	"io"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/tharsis/evmos/v4/tests/e2e/util"
)

func (s *IntegrationTestSuite) TestUpgrade() {
	s.initUpgrade()

	s.stopAllNodeContainers()
	s.replaceContainers()
	newGenesis := s.migrateGenesis(s.valResources[s.chains[0].ChainMeta.ID][0].Container.ID)
	s.replaceGenesis(newGenesis)
	s.stopAllNodeContainers()
	s.upgrade()
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chains[0].ChainMeta.ID][0].GetHostPort("1317/tcp"))
	balancesA, err := queryBalances(chainAAPIEndpoint, s.chains[0].Validators[0].PublicAddress)
	s.Require().NoError(err)
	s.Require().NotNil(balancesA)
	s.Require().Equal(2, len(balancesA))
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
