package ve

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abcitypes "github.com/evmos/evmos/v16/abci/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"
)

// VoteExtensionHandler is a handler that extends a vote with the oracle's
// current price feed. In the case where oracle data is unable to be fetched
// or correctly marshalled, the handler will return an empty vote extension to
// ensure liveliness.
type VoteExtensionHandler struct {
	logger log.Logger

	txDecoder sdk.TxDecoder

	evmClient       evmtypes.QueryClient
	feeMarketClient feemarkettypes.QueryClient

	// timeout is the maximum amount of time to wait for the client to respond
	timeout time.Duration

	// preBlocker is utilized to retrieve the latest on-chain information.
	preBlocker sdk.PreBlocker
}

// type VoteExtension struct {
// 	Bloom     *big.Int
// 	GasUsed   uint64
// 	BaseFee   *big.Int
// 	Miner     common.Address
// 	extraData []byte
// }

// NewVoteExtensionHandler returns a new VoteExtensionHandler.
func NewVoteExtensionHandler(
	logger log.Logger,
	txDecoder sdk.TxDecoder,
	evmClient evmtypes.QueryClient,
	feeMarketClient feemarkettypes.QueryClient,
	timeout time.Duration,
	preBlocker sdk.PreBlocker,
) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		logger:          logger,
		txDecoder:       txDecoder,
		feeMarketClient: feeMarketClient,
		timeout:         timeout,
		preBlocker:      preBlocker,
	}
}

func (h *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *cometabci.RequestExtendVote) (resp *cometabci.ResponseExtendVote, err error) {
		start := time.Now()
		// measure latencies from invocation to return, catch panics first
		defer func() {
			// catch panics if possible
			if r := recover(); r != nil {
				h.logger.Error(
					"recovered from panic in ExtendVoteHandler",
					"err", r,
				)

				resp, err = &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, fmt.Errorf("%v", r)
			}

			// measure latency
			latency := time.Since(start)
			h.logger.Info(
				"extend vote handler",
				"duration (seconds)", latency.Seconds(),
				"err", err,
			)

			// ignore all non-panic errors
			// var p ErrPanic
			// if !errors.As(err, &p) {
			// 	err = nil
			// }
		}()

		if req == nil {
			h.logger.Error("extend vote handler received a nil request")
			// err = slinkyabci.NilRequestError{
			// 	Handler: servicemetrics.ExtendVote,
			// }
			return nil, err
		}

		// Create a context with a timeout to ensure we do not wait forever for the oracle
		// to respond.
		reqCtx, cancel := context.WithTimeout(ctx.Context(), h.timeout)
		defer cancel()

		baseFee, err := abcitypes.GetBaseFee(reqCtx, h.evmClient)
		if err != nil {
			h.logger.Error(
				"failed to get base fee for vote extension; returning empty vote extension",
				"height", req.Height,
				"ctx_err", reqCtx.Err(),
				"err", err,
			)
			// return an empty vote extension to ensure liveliness
			return &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, err
		}

		miner, err := abcitypes.GetProposerAddres(reqCtx, h.evmClient, req.ProposerAddress)
		if err != nil {
			h.logger.Error(
				"failed to miner for vote extension; returning empty vote extension",
				"height", req.Height,
				"ctx_err", reqCtx.Err(),
				"err", err,
			)
			// return an empty vote extension to ensure liveliness
			return &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, err
		}

		// h.feeMarketClient.BlockGas(ctx, )

		ve := abcitypes.EVMVoteExtension{
			BaseFee:      *baseFee,
			BlockGasUsed: 0,
			Bloom:        nil,
			Miner:        miner.String(),
		}

		if err := ve.Validate(); err != nil {
			h.logger.Error(
				"failed to validate vote extension; returning empty vote extension",
				"height", req.Height,
				"err", err,
			)
			// return an empty vote extension to ensure liveliness
			return &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, err
		}

		// TODO: use codec?
		veBz, err := ve.Marshal()
		if err != nil {
			h.logger.Error(
				"failed to marshal vote extension; returning empty vote extension",
				"height", req.Height,
				"ctx_err", reqCtx.Err(),
				"err", err,
			)
			// return an empty vote extension to ensure liveliness
			return &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, err
		}

		return &cometabci.ResponseExtendVote{VoteExtension: veBz}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(_ sdk.Context, req *cometabci.RequestVerifyVoteExtension) (_ *cometabci.ResponseVerifyVoteExtension, err error) {
		var ve abcitypes.EVMVoteExtension
		if err := ve.Unmarshal(req.VoteExtension); err != nil {
			return nil, err
		}

		if ve.Height != req.Height {
			return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, ve.Height)
		}

		if err := ve.Validate(); err != nil {
			h.logger.Error(
				"failed to validate vote extension; returning empty vote extension",
				"height", req.Height,
				"err", err,
			)
			return nil, err
		}

		return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func ValidateVoteExtension(ctx sdk.Context, voteExt abcitypes.EVMVoteExtension) error {
	return voteExt.Validate()
}
