package streaming

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	abci "github.com/cometbft/cometbft/abci/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v15/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// ABCIListener interface used to hook into the ABCI message processing of the BaseApp.
// the error results are propagated to consensus state machine,
// if you don't want to affect consensus, handle the errors internally and always return `nil` in these APIs.
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error
	// ListenCommit updates the steaming service with the latest Commit event
	ListenCommit(ctx context.Context, res abci.ResponseCommit) error
}

// StreamingService interface for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService interface {
	// Stream is the streaming service loop, awaits kv pairs and writes them to some destination stream or file
	Stream(wg *sync.WaitGroup) error
	// Listeners returns the streaming service's listeners for the BaseApp to register
	Listeners() map[storetypes.StoreKey][]storetypes.WriteListener
	// ABCIListener interface for hooking into the ABCI messages from inside the BaseApp
	ABCIListener
	// Closer interface
	io.Closer
}

var _ StreamingService = &Streamer{}

type Streamer struct {
	TxDecoder sdk.TxDecoder
}

func (s Streamer) Stream(wg *sync.WaitGroup) error {
	return errors.New("not supported")
}

// Listeners returns the streaming service's listeners for the BaseApp to register
func (s Streamer) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener {
	return nil
}

func (s Streamer) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	return errors.New("not supported")
}

func (s Streamer) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	return errors.New("not supported")
}

func (s Streamer) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	tx, err := s.TxDecoder(req.Tx)
	if err != nil {
		return err
	}

	// update tx counter

	if res.IsErr() {
		// TODO: add error tx counter
	}

	if evmostypes.IsEthTx(tx) {
	}

	return errors.New("not supported")
}

func (s Streamer) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	return errors.New("not supported")
}

func (s Streamer) Close() error {
	return nil
}

func (s Streamer) ProcessEthereumTx(tx sdk.Tx) error {
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return fmt.Errorf("invalid tx: %T", tx)
	}

	ethMsg, ok := msgs[0].(*evmtypes.MsgEthereumTx)
	if !ok {
		return fmt.Errorf("invalid tx type: %T", tx)
	}

	// from := common.HexToAddress(ethMsg.From)

	ethTx := ethMsg.AsTransaction()
	if ethMsg.Hash == "" {
		ethMsg.Hash = ethTx.Hash().Hex()
	}

	return nil
}
