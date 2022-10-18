package v4_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	tmclient "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/evmos/v9/app"
	v4 "github.com/evmos/evmos/v9/app/upgrades/v4"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx                    sdk.Context
	app                    *app.Evmos
	consAddress            sdk.ConsAddress
	expiredOsmoClient      *tmclient.ClientState
	activeOsmoClient       *tmclient.ClientState
	expiredCosmosHubClient *tmclient.ClientState
	activeCosmosHubClient  *tmclient.ClientState
	osmoConsState          *tmclient.ConsensusState
	cosmosHubConsState     *tmclient.ConsensusState
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
		Time:            time.Date(2022, 5, 9, 8, 0, 0, 0, time.UTC),
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

	// set expired clients
	suite.expiredOsmoClient = &tmclient.ClientState{
		ChainId:                      "osmosis-1",
		TrustLevel:                   tmclient.DefaultTrustLevel,
		TrustingPeriod:               10 * 24 * time.Hour,
		UnbondingPeriod:              14 * 24 * time.Hour,
		MaxClockDrift:                25 * time.Second,
		FrozenHeight:                 clienttypes.NewHeight(0, 0),
		LatestHeight:                 clienttypes.NewHeight(1, 3484087),
		UpgradePath:                  []string{"upgrade", "upgradedIBCState"},
		AllowUpdateAfterExpiry:       true,
		AllowUpdateAfterMisbehaviour: true,
	}

	// set active clients
	suite.activeOsmoClient = &tmclient.ClientState{
		ChainId:                      "osmosis-1",
		TrustLevel:                   tmclient.DefaultTrustLevel,
		TrustingPeriod:               10 * 24 * time.Hour,
		UnbondingPeriod:              14 * 24 * time.Hour,
		MaxClockDrift:                25 * time.Second,
		FrozenHeight:                 clienttypes.NewHeight(0, 0),
		LatestHeight:                 clienttypes.NewHeight(1, 4264373),
		UpgradePath:                  []string{"upgrade", "upgradedIBCState"},
		AllowUpdateAfterExpiry:       true,
		AllowUpdateAfterMisbehaviour: true,
	}

	suite.expiredCosmosHubClient = &tmclient.ClientState{
		ChainId:                      "cosmoshub-4",
		TrustLevel:                   tmclient.DefaultTrustLevel,
		TrustingPeriod:               10 * 24 * time.Hour,
		UnbondingPeriod:              21 * 24 * time.Hour,
		MaxClockDrift:                20 * time.Second,
		FrozenHeight:                 clienttypes.NewHeight(0, 0),
		LatestHeight:                 clienttypes.NewHeight(4, 9659547),
		UpgradePath:                  []string{"upgrade", "upgradedIBCState"},
		AllowUpdateAfterExpiry:       true,
		AllowUpdateAfterMisbehaviour: true,
	}

	suite.activeCosmosHubClient = &tmclient.ClientState{
		ChainId:                      "cosmoshub-4",
		TrustLevel:                   tmclient.DefaultTrustLevel,
		TrustingPeriod:               10 * 24 * time.Hour,
		UnbondingPeriod:              21 * 24 * time.Hour,
		MaxClockDrift:                20 * time.Second,
		FrozenHeight:                 clienttypes.NewHeight(0, 0),
		LatestHeight:                 clienttypes.NewHeight(4, 10409568),
		UpgradePath:                  []string{"upgrade", "upgradedIBCState"},
		AllowUpdateAfterExpiry:       true,
		AllowUpdateAfterMisbehaviour: true,
	}

	suite.osmoConsState = &tmclient.ConsensusState{
		Timestamp: time.Date(2022, 5, 4, 23, 41, 9, 152600097, time.UTC),
		// Root:               types.NewMerkleRoot([]byte("q3R1c0MR0+nJqcIMJzcEZYmB0Q1wNammv+yU6IVowro=")),
		// NextValidatorsHash: tmbytes.HexBytes("195FC0E44AC0AA591F72AF0FEADD8F6279B2174353C05965850360532FE1E5A8"),
	}

	suite.cosmosHubConsState = &tmclient.ConsensusState{
		Timestamp: time.Date(2022, 4, 29, 11, 9, 59, 595932461, time.UTC),
		// Root:               types.NewMerkleRoot([]byte("q3lQF/LzwRElBj7V1LW9O/yote+/yfzOfOE/o94phsNQ=")),
		// NextValidatorsHash: tmbytes.HexBytes("FADA047AE843608F6C3639E7C82199E5F36B791CB1EB1B2CC034B00C68FC6B6D"),
	}

	// FIXME: unmarshal client and consensus state
	// var expiredOsmoClientRes *clienttypes.QueryClientStateResponse
	// expiredOsmoClientJSON := `{"client_state":{"@type":"/ibc.lightclients.tendermint.v1.ClientState","chain_id":"osmosis-1","trust_level":{"numerator":"1","denominator":"3"},"trusting_period":"864000s","unbonding_period":"1209600s","max_clock_drift":"25s","frozen_height":{"revision_number":"0","revision_height":"0"},"latest_height":{"revision_number":"1","revision_height":"3484087"},"proof_specs":[{"leaf_spec":{"hash":"SHA256","prehash_key":"NO_HASH","prehash_value":"SHA256","length":"VAR_PROTO","prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":33,"min_prefix_length":4,"max_prefix_length":12,"empty_child":null,"hash":"SHA256"},"max_depth":0,"min_depth":0},{"leaf_spec":{"hash":"SHA256","prehash_key":"NO_HASH","prehash_value":"SHA256","length":"VAR_PROTO","prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":32,"min_prefix_length":1,"max_prefix_length":1,"empty_child":null,"hash":"SHA256"},"max_depth":0,"min_depth":0}],"upgrade_path":["upgrade","upgradedIBCState"],"allow_update_after_expiry":true,"allow_update_after_misbehaviour":true},"proof":"CoQICoEICiNjbGllbnRzLzA3LXRlbmRlcm1pbnQtMC9jbGllbnRTdGF0ZRKxAQorL2liYy5saWdodGNsaWVudHMudGVuZGVybWludC52MS5DbGllbnRTdGF0ZRKBAQoJb3Ntb3Npcy0xEgQIARADGgQIgN40IgQIgOpJKgIIGTIAOgcIARC309QBQhkKCQgBGAEgASoBABIMCgIAARAhGAQgDDABQhkKCQgBGAEgASoBABIMCgIAARAgGAEgATABSgd1cGdyYWRlShB1cGdyYWRlZElCQ1N0YXRlUAFYARoLCAEYASABKgMAAgIiKwgBEicCBOaVCSBH/a+xq/sEchIMBlVVduKqd80PT7HKbeFclfSm65RxRSAiKwgBEicECOaVCSD14U6gEfcqGwF9t+s5rKrC5ZJ2Zw72/r1/RqxL00YwnSAiKwgBEicGEOaVCSBS0lI8puPYR3MJsJ0xHnlRF3j8OunVuEIxh+CFGDHb3iAiLQgBEgYIIOaVCSAaISAk473pEXrAVabOcOfQ6yRJxaYUDaeLpOovHQa6txU6XSItCAESBgxg0KcMIBohIKpAXKjN+Ih4SteLVhxex1lO6Aa0p7XF05hCUOSdcR4HIiwIARIoDsIB8PQOIH2KsT5xoCdteclIZ/bjQe0/o3dMPr6KGoA7cbUJN8rWICIuCAESBxLCBciRFyAaISAg//KZy8113irmLWjXNiFMQaJCYtU2ekq/PI70zLOMfSIsCAESKBScCuDJGSBVblw7HX+R9QdfiW6Q4Mmp8wOMHzxdRw5JxbdlTh8+BiAiLAgBEigW1A/gyRkgb1GW9IvpEeRjt3oQrKDE+zSItj7Ia64hwvfiyOXrmicgIiwIARIoGNYn4MkZIH+ht3W2ht98F8+77ICUuwIZ/bUp/H1QKF9pOwR21AvcICIsCAESKBrWR+DJGSBPJlKw4/jXwcI28GtCDtB/GdjRzHlVY2rCYOCgbHQr+iAiLQgBEikc1ocB4MkZIKThntJ6+2nP9aB5JgY27O/STo7t6wge917188FdW+XTICItCAESKR7WhwLgyRkg4xxvGOIOpQFxHeoCx+fe0A1+OCR4edGGR3puI2wtqaogIi8IARIIINaHBODJGSAaISDCUjwNqYBIsOEG9X5mOnJEVsWsjxvKwzQ49a8Rb5XpsSItCAESKSLWhwjgyRkgQZ2ZXwM33mMiBafzvfA07pzCAjkfG6sAKKgC+cYQG8ogIi8IARIIJNaHEODJGSAaISBfPk8QONuPW2L93O+LBpo8CNUqkoXGl3RCpTPSOrFo9SIvCAESCCaQyCDq0BkgGiEgjgdsb/ZSFOXHy9vvDibUPwrgHcYd/+ab8q1cA7cUp0IK/AEK+QEKA2liYxIgp4EvKBkIIHILneBBL6M4apBIx5dgH84zHgXQwsW5AZ8aCQgBGAEgASoBACIlCAESIQHbc5uia4ZUGiJzE2JxCRNaxqkStE/NVbiKohR1BkM93CInCAESAQEaIFBZeb1cCLEvt5jj4oZbR3QtFGOncSCAFObe6Vi8lTnNIiUIARIhAaZj320mJUhEB3tSCaSyv38zf5kMjOLCsebFOdyDwgokIiUIARIhAfkCH7Rm146Z6uTX0tPrXdGHKu0XkJ4W9efa+mzqz3rAIicIARIBARoga4zywF/cEvKIov7jysbh37sH2KgVT+EfG687MriD/DU=","proof_height":{"revision_number":"2","revision_height":"209977"}}`

	// err = suite.app.AppCodec().UnmarshalJSON([]byte(expiredOsmoClientJSON), expiredOsmoClientRes)
	// suite.Require().NoError(err)
	// suite.Require().NotNil(expiredOsmoClientRes)
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
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredOsmosisClient, suite.expiredOsmoClient)
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredCosmosHubClient, suite.expiredCosmosHubClient)

				// set active clients
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ActiveOsmosisClient, suite.activeOsmoClient)
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ActiveCosmosHubClient, suite.activeCosmosHubClient)

				// set active consensus states
				suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, v4.ActiveOsmosisClient, suite.activeOsmoClient.GetLatestHeight(), suite.osmoConsState)
				suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, v4.ActiveCosmosHubClient, suite.activeCosmosHubClient.GetLatestHeight(), suite.cosmosHubConsState)

				activeOsmoStore := suite.app.IBCKeeper.ClientKeeper.ClientStore(suite.ctx, v4.ActiveOsmosisClient)
				activeHubStore := suite.app.IBCKeeper.ClientKeeper.ClientStore(suite.ctx, v4.ActiveCosmosHubClient)

				// set processing time and height
				tmclient.SetProcessedHeight(activeOsmoStore, suite.activeOsmoClient.LatestHeight, suite.activeOsmoClient.LatestHeight)
				tmclient.SetProcessedHeight(activeHubStore, suite.activeCosmosHubClient.LatestHeight, suite.activeCosmosHubClient.LatestHeight)

				tmclient.SetProcessedTime(activeOsmoStore, suite.activeOsmoClient.LatestHeight, suite.osmoConsState.GetTimestamp())
				tmclient.SetProcessedTime(activeHubStore, suite.activeCosmosHubClient.LatestHeight, suite.cosmosHubConsState.GetTimestamp())
			},
			false,
		},
		{
			"Osmosis IBC client update failed",
			func() {
				// set expired clients
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredOsmosisClient, suite.expiredOsmoClient)

				// set active clients
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ActiveCosmosHubClient, suite.activeCosmosHubClient)
			},
			true,
		},
		{
			"Cosmos Hub IBC client update failed",
			func() {
				// set expired clients
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredOsmosisClient, suite.expiredOsmoClient)
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ExpiredCosmosHubClient, suite.expiredCosmosHubClient)

				// set active clients
				suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, v4.ActiveOsmosisClient, suite.activeOsmoClient)

				// set active consensus states
				suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, v4.ActiveOsmosisClient, suite.activeOsmoClient.GetLatestHeight(), suite.osmoConsState)

				activeOsmoStore := suite.app.IBCKeeper.ClientKeeper.ClientStore(suite.ctx, v4.ActiveOsmosisClient)

				// set processing time and height
				tmclient.SetProcessedHeight(activeOsmoStore, suite.activeOsmoClient.LatestHeight, suite.activeOsmoClient.LatestHeight)
				tmclient.SetProcessedTime(activeOsmoStore, suite.activeOsmoClient.LatestHeight, suite.osmoConsState.GetTimestamp())
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// test if we can update the clients
			canUpdateClient := tmclient.IsMatchingClientState(*suite.expiredOsmoClient, *suite.activeOsmoClient)
			suite.Require().True(canUpdateClient, "cannot update non-matching osmosis client")

			canUpdateClient = tmclient.IsMatchingClientState(*suite.expiredCosmosHubClient, *suite.activeCosmosHubClient)
			suite.Require().True(canUpdateClient, "cannot update non-matching cosmos hub client")

			err := v4.UpdateIBCClients(suite.ctx, suite.app.IBCKeeper.ClientKeeper)
			if tc.expError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
