package v5_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	"github.com/evmos/evmos/v9/app"
	v5 "github.com/evmos/evmos/v9/app/upgrades/v5"
	evmostypes "github.com/evmos/evmos/v9/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Evmos
	consAddress sdk.ConsAddress
}

func (suite *UpgradeTestSuite) SetupTest(chainID string) {
	feemarkettypes.DefaultMinGasPrice = v5.MainnetMinGasPrices
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         chainID,
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

	cp := suite.app.BaseApp.GetConsensusParams(suite.ctx)
	suite.ctx = suite.ctx.WithConsensusParams(cp)
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (suite *UpgradeTestSuite) TestUpdateConsensusParams() {
	testCases := []struct {
		name              string
		malleate          func()
		expEvidenceParams *tmproto.EvidenceParams
	}{
		{
			"empty evidence params",
			func() {
				subspace, found := suite.app.ParamsKeeper.GetSubspace(baseapp.Paramspace)
				suite.Require().True(found)

				ep := &tmproto.EvidenceParams{}
				subspace.Set(suite.ctx, baseapp.ParamStoreKeyEvidenceParams, ep)
			},
			&tmproto.EvidenceParams{},
		},
		{
			"success",
			func() {
				subspace, found := suite.app.ParamsKeeper.GetSubspace(baseapp.Paramspace)
				suite.Require().True(found)

				ep := &tmproto.EvidenceParams{
					MaxAgeDuration:  2 * 24 * time.Hour,
					MaxAgeNumBlocks: 100000,
					MaxBytes:        suite.ctx.ConsensusParams().Evidence.MaxBytes,
				}
				subspace.Set(suite.ctx, baseapp.ParamStoreKeyEvidenceParams, ep)
			},
			&tmproto.EvidenceParams{
				MaxAgeDuration:  stakingtypes.DefaultUnbondingTime,
				MaxAgeNumBlocks: int64(stakingtypes.DefaultUnbondingTime / (2 * time.Second)),
				MaxBytes:        suite.ctx.ConsensusParams().Evidence.MaxBytes,
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(evmostypes.TestnetChainID + "-2") // reset

			tc.malleate()

			suite.Require().NotPanics(func() {
				v5.UpdateConsensusParams(suite.ctx, suite.app.StakingKeeper, suite.app.ParamsKeeper)
				suite.app.Commit()
			})

			cp := suite.app.BaseApp.GetConsensusParams(suite.ctx)
			suite.Require().NotNil(cp)
			suite.Require().NotNil(cp.Evidence)
			suite.Require().Equal(tc.expEvidenceParams.String(), cp.Evidence.String())
		})
	}
}

func (suite *UpgradeTestSuite) TestUpdateIBCDenomTraces() {
	testCases := []struct {
		name           string
		originalTraces ibctransfertypes.Traces
		expDenomTraces ibctransfertypes.Traces
	}{
		{
			"no traces",
			ibctransfertypes.Traces{},
			ibctransfertypes.Traces{},
		},
		{
			"native IBC tokens",
			ibctransfertypes.Traces{
				{
					BaseDenom: "aevmos",
					Path:      "",
				},
				{
					BaseDenom: "uosmo",
					Path:      "transfer/channel-0",
				},
				{
					BaseDenom: "uatom",
					Path:      "transfer/channel-0/transfer/channel-0",
				},
				{
					BaseDenom: "uatom",
					Path:      "transfer/channel-3",
				},
				{
					BaseDenom: "gravity0x6B175474E89094C44Da98b954EedeAC495271d0F",
					Path:      "transfer/channel-8",
				},
			},
			ibctransfertypes.Traces{
				{
					BaseDenom: "aevmos",
					Path:      "",
				},
				{
					BaseDenom: "uatom",
					Path:      "transfer/channel-0/transfer/channel-0",
				},
				{
					BaseDenom: "uosmo",
					Path:      "transfer/channel-0",
				},
				{
					BaseDenom: "uatom",
					Path:      "transfer/channel-3",
				},
				{
					BaseDenom: "gravity0x6B175474E89094C44Da98b954EedeAC495271d0F",
					Path:      "transfer/channel-8",
				},
			},
		},
		{
			"with invalid tokens",
			ibctransfertypes.Traces{
				{
					BaseDenom: "aevmos",
					Path:      "",
				},
				{
					BaseDenom: "uatom",
					Path:      "transfer/channel-3",
				},
				{
					BaseDenom: "uosmo",
					Path:      "transfer/channel-0/transfer/channel-0",
				},
				{
					BaseDenom: "1",
					Path:      "transfer/channel-0/gamm/pool",
				},
				{
					BaseDenom: "0x85bcBCd7e79Ec36f4fBBDc54F90C643d921151AA",
					Path:      "transfer/channel-20/erc20",
				},
			},
			ibctransfertypes.Traces{
				{
					BaseDenom: "aevmos",
					Path:      "",
				},
				{
					BaseDenom: "gamm/pool/1",
					Path:      "transfer/channel-0",
				},
				{
					BaseDenom: "uosmo",
					Path:      "transfer/channel-0/transfer/channel-0",
				},
				{
					BaseDenom: "erc20/0x85bcBCd7e79Ec36f4fBBDc54F90C643d921151AA",
					Path:      "transfer/channel-20",
				},
				{
					BaseDenom: "uatom",
					Path:      "transfer/channel-3",
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(evmostypes.TestnetChainID + "-2") // reset

			for _, dt := range tc.originalTraces {
				suite.app.TransferKeeper.SetDenomTrace(suite.ctx, dt)
			}

			v5.UpdateIBCDenomTraces(suite.ctx, suite.app.TransferKeeper)

			traces := suite.app.TransferKeeper.GetAllDenomTraces(suite.ctx)
			suite.Require().Equal(tc.expDenomTraces, traces)
		})
	}
}
