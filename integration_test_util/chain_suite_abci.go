package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/gogoproto/proto"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"github.com/pkg/errors"
	"time"
)

// commitAndCreateNewCtx commits a block at a given time, creating and return a new ctx for the next block
func (suite *ChainIntegrationTestSuite) commitAndCreateNewCtx(ctx sdk.Context, t time.Duration, vs *tmtypes.ValidatorSet) (sdk.Context, *tmtypes.ValidatorSet, error) {
	header, nextVs, err := suite.commit(ctx, t, vs)
	if err != nil {
		return ctx, nil, err
	}

	newCtx := suite.createNewContext(ctx, header)

	return newCtx, nextVs, nil
}

// createNewContext returns a new sdk.Context with the same settings as the old one
func (suite *ChainIntegrationTestSuite) createNewContext(oldCtx sdk.Context, header tmproto.Header) sdk.Context {
	// NewContext function keeps the multistore
	// but resets other context fields
	// GasMeter is set as InfiniteGasMeter
	var newCtx sdk.Context
	if suite.HasTendermint() {
		newCtx = sdk.NewContext(suite.BaseApp().CommitMultiStore(), header, false, suite.BaseApp().Logger())
	} else {
		newCtx = suite.BaseApp().NewContext(false, header)
	}
	// set the reset-ted fields to keep the current ctx settings
	newCtx = newCtx.WithMinGasPrices(oldCtx.MinGasPrices())
	newCtx = newCtx.WithEventManager(oldCtx.EventManager())
	newCtx = newCtx.WithKVGasConfig(oldCtx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(oldCtx.TransientKVGasConfig())

	return newCtx
}

// DeliverTx delivers a Cosmos tx for a given set of msgs.
// The delivery mode is SYNC
func (suite *ChainIntegrationTestSuite) DeliverTx(
	ctx sdk.Context,
	signer *itutiltypes.TestAccount,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (authsigning.Tx, abci.ResponseDeliverTx, error) {
	suite.Require().NotNil(signer)

	tx, err := suite.PrepareCosmosTx(
		ctx,
		signer,
		CosmosTxArgs{
			Gas:      10_000_000,
			GasPrice: gasPrice,
			Msgs:     msgs,
		},
	)
	if err != nil {
		return nil, abci.ResponseDeliverTx{}, err
	}
	resDeliverTx, err := suite.BroadcastTx(tx)
	return tx, resDeliverTx, err
}

// DeliverTxAsync is the same as DeliverTx but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) DeliverTxAsync(
	ctx sdk.Context,
	signer *itutiltypes.TestAccount,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (*coretypes.ResultBroadcastTx, error) {
	suite.Require().NotNil(signer)

	tx, err := suite.PrepareCosmosTx(
		ctx,
		signer,
		CosmosTxArgs{
			Gas:      10_000_000,
			GasPrice: gasPrice,
			Msgs:     msgs,
		},
	)
	if err != nil {
		return nil, err
	}
	return suite.BroadcastTxAsync(tx)
}

// DeliverEthTx generates and broadcasts MsgEthereumTx message populated within a Cosmos tx.
// The delivery mode is SYNC
func (suite *ChainIntegrationTestSuite) DeliverEthTx(
	signer *itutiltypes.TestAccount,
	ethMsg *evmtypes.MsgEthereumTx,
) (*itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(signer)

	tx, err := suite.PrepareEthTx(signer, ethMsg)
	if err != nil {
		return nil, err
	}
	responseDeliverTx, err := suite.BroadcastTx(tx)
	if err != nil {
		return nil, err
	}

	res := itutiltypes.NewResponseDeliverEthTx(&responseDeliverTx)

	if _, err := checkEthTxResponse(responseDeliverTx, suite.EncodingConfig.Codec); err != nil {
		return res, err
	}
	return res, nil
}

// DeliverEthTxAsync is the same as DeliverEthTx but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) DeliverEthTxAsync(
	account *itutiltypes.TestAccount,
	ethMsg *evmtypes.MsgEthereumTx,
) error {
	suite.Require().NotNil(account)

	tx, err := suite.PrepareEthTx(account, ethMsg)
	if err != nil {
		return err
	}
	_, err = suite.BroadcastTxAsync(tx)
	return err
}

// BroadcastTx does broadcast a tx over the network and returns the response
// The delivery mode is SYNC
func (suite *ChainIntegrationTestSuite) BroadcastTx(tx sdk.Tx) (responseDeliverTx abci.ResponseDeliverTx, err error) {
	// bz are bytes to be broadcast over the network
	var bz []byte
	bz, err = suite.EncodingConfig.TxConfig.TxEncoder()(tx)

	if err == nil {
		if suite.HasTendermint() {
			res, err := suite.QueryClients.TendermintRpcHttpClient.BroadcastTxCommit(context.Background(), bz)
			suite.Require().NoError(err)
			responseDeliverTx = res.DeliverTx
		} else {
			responseDeliverTx = suite.BaseApp().DeliverTx(
				abci.RequestDeliverTx{
					Tx: bz,
				},
			)
		}

		if responseDeliverTx.Code != 0 {
			err = errorsmod.Wrapf(errortypes.ErrInvalidRequest, responseDeliverTx.Log)
			responseDeliverTx = abci.ResponseDeliverTx{} // purge
		}
	}

	return
}

// BroadcastTxAsync is the same as BroadcastTx but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) BroadcastTxAsync(tx sdk.Tx) (resultBroadcastTx *coretypes.ResultBroadcastTx, err error) {
	suite.EnsureTendermint()
	// bz are bytes to be broadcast over the network
	var bz []byte
	bz, err = suite.EncodingConfig.TxConfig.TxEncoder()(tx)

	if err == nil {
		res, err := suite.QueryClients.TendermintRpcHttpClient.BroadcastTxAsync(context.Background(), bz)
		suite.Require().NoError(err)
		resultBroadcastTx = res
	}

	return
}

// commit is helper function, it:
//
// - Runs the EndBlocker logic.
//
// - Commits the changes.
//
// - Updates the header.
//
// - Runs the BeginBlocker logic.
//
// - Finally, returns the updated header.
func (suite *ChainIntegrationTestSuite) commit(ctx sdk.Context, t time.Duration, vs *tmtypes.ValidatorSet) (tmproto.Header, *tmtypes.ValidatorSet, error) {
	var nextVals *tmtypes.ValidatorSet

	baseApp := suite.BaseApp()

	header := ctx.BlockHeader()

	res := baseApp.EndBlock(abci.RequestEndBlock{
		Height: header.Height,
	})

	if vs != nil {
		var err error
		nextVals, err = applyValSetChanges(vs, res.ValidatorUpdates)
		if err != nil {
			return header, nil, err
		}
		header.ValidatorsHash = vs.Hash()
		header.NextValidatorsHash = nextVals.Hash()
	}

	_ = baseApp.Commit()

	header.Height++
	header.Time = header.Time.Add(t)
	header.AppHash = baseApp.LastCommitID().Hash

	baseApp.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	return header, nextVals, nil
}

// applyValSetChanges applies the validator set changes to the given validator set
func applyValSetChanges(valSet *tmtypes.ValidatorSet, valUpdates []abci.ValidatorUpdate) (*tmtypes.ValidatorSet, error) {
	updates, err := tmtypes.PB2TM.ValidatorUpdates(valUpdates)
	if err != nil {
		return nil, err
	}

	// must copy since validator set will mutate with UpdateWithChangeSet
	newVals := valSet.Copy()
	err = newVals.UpdateWithChangeSet(updates)
	if err != nil {
		return nil, err
	}

	return newVals, nil
}

func checkEthTxResponse(r abci.ResponseDeliverTx, cdc codec.Codec) ([]*evmtypes.MsgEthereumTxResponse, error) {
	if !r.IsOK() {
		return nil, fmt.Errorf("tx failed. Code: %d, Logs: %s", r.Code, r.Log)
	}

	var txData sdk.TxMsgData
	if err := cdc.Unmarshal(r.Data, &txData); err != nil {
		return nil, err
	}

	if len(txData.MsgResponses) == 0 {
		return nil, fmt.Errorf("no message responses found")
	}

	responses := make([]*evmtypes.MsgEthereumTxResponse, 0, len(txData.MsgResponses))
	for i := range txData.MsgResponses {
		var res evmtypes.MsgEthereumTxResponse
		if err := proto.Unmarshal(txData.MsgResponses[i].Value, &res); err != nil {
			// TODO use corresponding proto for each chain
			return nil, errors.Wrap(err, "failed to unmarshal proto")
		}

		if res.Failed() {
			return nil, fmt.Errorf("tx failed. VmError: %s", res.VmError)
		}
		responses = append(responses, &res)
	}

	return responses, nil
}
