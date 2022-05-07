package v4_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/tharsis/evmos/v4/app"
	v4 "github.com/tharsis/evmos/v4/app/upgrades/v4"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Evmos
	consAddress sdk.ConsAddress
}

func (suite *UpgradeTestSuite) SetupTest() {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

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
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (suite *UpgradeTestSuite) TestUpdateIBCClients() {
	testCases := []struct {
		name     string
		malleate func()
		expError bool
	}{
		{
			"IBC clients updated successfully",
			func() {
				// set expired clients
				var expiredOsmoClient exported.ClientState = &tmclient.ClientState{
					ChainId:         "osmosis-1",
					TrustLevel:      tmclient.DefaultTrustLevel,
					TrustingPeriod:  10 * 24 * time.Hour,
					UnbondingPeriod: 14 * 24 * time.Hour,
					MaxClockDrift:   25 * time.Second,
					FrozenHeight:    clienttypes.NewHeight(0, 0),
					LatestHeight:    clienttypes.NewHeight(1, 3484087),
					UpgradePath:     []string{"upgrade", "upgradedIBCState"},
				}

				var expiredCosmosHubClient exported.ClientState = &tmclient.ClientState{
					ChainId:         "cosmoshub-4",
					TrustLevel:      tmclient.DefaultTrustLevel,
					TrustingPeriod:  10 * 24 * time.Hour,
					UnbondingPeriod: 21 * 24 * time.Hour,
					MaxClockDrift:   20 * time.Second,
					FrozenHeight:    clienttypes.NewHeight(0, 0),
					LatestHeight:    clienttypes.NewHeight(4, 9659547),
					UpgradePath:     []string{"upgrade", "upgradedIBCState"},
				}

				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredOsmosisClient, expiredOsmoClient)
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredCosmosHubClient, expiredCosmosHubClient)

				// set active clients
				var activeOsmoClient exported.ClientState = &tmclient.ClientState{
					ChainId:         "osmosis-1",
					TrustLevel:      tmclient.DefaultTrustLevel,
					TrustingPeriod:  9 * 24 * time.Hour,
					UnbondingPeriod: 14 * 24 * time.Hour,
					MaxClockDrift:   1035 * time.Second,
					FrozenHeight:    clienttypes.NewHeight(0, 0),
					LatestHeight:    clienttypes.NewHeight(1, 4264373),
					UpgradePath:     []string{"upgrade", "upgradedIBCState"},
				}

				var osmoConsState exported.ConsensusState = &tmclient.ConsensusState{
					Timestamp:          time.Date(2022, 0o5, 4, 23, 41, 9, 152600097, time.UTC),
					Root:               types.NewMerkleRoot([]byte("q3R1c0MR0+nJqcIMJzcEZYmB0Q1wNammv+yU6IVowro=")),
					NextValidatorsHash: tmbytes.HexBytes("195FC0E44AC0AA591F72AF0FEADD8F6279B2174353C05965850360532FE1E5A8"),
				}

				var activeCosmosHubClient exported.ClientState = &tmclient.ClientState{
					ChainId:         "cosmoshub-4",
					TrustLevel:      tmclient.DefaultTrustLevel,
					TrustingPeriod:  14 * 24 * time.Hour,
					UnbondingPeriod: 21 * 24 * time.Hour,
					MaxClockDrift:   2530 * time.Second,
					FrozenHeight:    clienttypes.NewHeight(0, 0),
					LatestHeight:    clienttypes.NewHeight(4, 10293839),
					UpgradePath:     []string{"upgrade", "upgradedIBCState"},
				}

				suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, v4.ExpiredCosmosHubClient, activeOsmoClient.GetLatestHeight(), osmoConsState)
				suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, v4.ActiveOsmosisClient, activeOsmoClient.GetLatestHeight(), osmoConsState)
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ActiveOsmosisClient, activeOsmoClient)
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ActiveCosmosHubClient, activeCosmosHubClient)
			},
			false,
		},
		{
			"Osmosis IBC client update failed",
			func() {},
			false,
		},
		{
			"Cosmos Hub IBC client update failed",
			func() {},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := v4.UpdateIBCClients(suite.ctx, suite.app.IBCKeeper.ClientKeeper)
			if tc.expError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
