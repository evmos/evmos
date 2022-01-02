package app

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"
	"github.com/tharsis/ethermint/encoding"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// Evmos testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   -1, // no limit
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// Setup initializes a new Evmos. A Nop logger is set in Evmos.
func Setup(isCheckTx bool, feemarketGenesis *feemarkettypes.GenesisState) *Evmos {
	db := dbm.NewMemDB()
	app := NewEvmos(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, encoding.MakeConfig(ModuleBasics), simapp.EmptyAppOptions{})
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()

		// Verify feeMarket genesis
		if feemarketGenesis != nil {
			if err := feemarketGenesis.Validate(); err != nil {
				panic(err)
			}

			genesisState[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				ChainId:         "evmos_9000-1",
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// SetupTestingApp initializes the IBC-go testing application
func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	cfg := encoding.MakeConfig(ModuleBasics)
	app := NewEvmos(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, cfg, simapp.EmptyAppOptions{})
	return app, NewDefaultGenesisState()
}
