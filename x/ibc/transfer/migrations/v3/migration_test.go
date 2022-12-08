package v3_test

import (
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/evmos/evmos/v10/app"
	v3 "github.com/evmos/evmos/v10/x/ibc/transfer/migrations/v3"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

func setupTestApp(t *testing.T) (*app.Evmos, sdk.Context) {
	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	// init app
	app := app.Setup(false, feemarkettypes.DefaultGenesisState())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
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
	return app, ctx
}

func TestMigrateEscrowAcc(t *testing.T) {
	app, ctx := setupTestApp(t)

	// Run migrations
	err := v3.MigrateEscrowAccounts(ctx, app.AccountKeeper)
	require.NoError(t, err)

	// check account types for channels 0 to 36
	for i := 0; i <= 36; i++ {
		ch := "channel-" + strconv.Itoa(i)
		addr := ibctypes.GetEscrowAddress(ibctypes.PortID, ch)
		acc := app.AccountKeeper.GetAccount(ctx, addr)

		_, isModuleAccount := acc.(authtypes.ModuleAccountI)
		require.True(t, isModuleAccount)
	}

}
