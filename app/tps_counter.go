package app

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	tagKeyStatus     = tag.MustNewKey("status")
	mTransactions    = stats.Int64("transactions", "the number of transactions after .EndBlocker", "1")
	viewTransactions = &view.View{
		Name:        "transactions_processed",
		Measure:     mTransactions,
		Description: "The transactions processed",
		TagKeys:     []tag.Key{tagKeyStatus},
		Aggregation: view.Count(),
	}
)

func ObservabilityViews() (views []*view.View) {
	views = append(views, viewTransactions)
	return views
}

type tpsCounter struct {
	nSuccessful, NFailed uint64
	reportPeriod         time.Duration
	logger               log.Logger
	doneCloseOnce        sync.Once
	doneCh               chan bool
}

func newTPSCounter(logger log.Logger) *tpsCounter {
	return &tpsCounter{logger: logger, doneCh: make(chan bool, 1)}
}

func (tpc *tpsCounter) incrementSuccess() { atomic.AddUint64(&tpc.nSuccessful, 1) }
func (tpc *tpsCounter) incrementFailure() { atomic.AddUint64(&tpc.NFailed, 1) }

const defaultTPSReportPeriod = 10 * time.Second

func (tpc *tpsCounter) start(ctx context.Context) error {
	tpsReportPeriod := defaultTPSReportPeriod
	if tpc.reportPeriod > 0 {
		tpsReportPeriod = tpc.reportPeriod
	}
	ticker := time.NewTicker(tpsReportPeriod)
	defer ticker.Stop()
	defer tpc.doneCloseOnce.Do(func() {
		close(tpc.doneCh)
	})

	var lastNSuccessful, lastNFailed uint64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			// Report the number of transactions seen in the designated period of time.
			latestNSuccessful := atomic.LoadUint64(&tpc.nSuccessful)
			latestNFailed := atomic.LoadUint64(&tpc.NFailed)

			var nTxn int64
			nSuccess, err := tpc.recordValue(ctx, latestNSuccessful, lastNSuccessful, statusSuccess)
			if err == nil {
				nTxn += nSuccess
			} else {
				panic(err)
			}
			nFailed, err := tpc.recordValue(ctx, latestNFailed, lastNFailed, statusFailure)
			if err == nil {
				nTxn += nFailed
			} else {
				panic(err)
			}

			if nTxn != 0 {
				// Record to our logger for easy examination in the logs.
				secs := float64(tpsReportPeriod) / float64(time.Second)
				tpc.logger.Info("Transactions per second", "tps", float64(nTxn)/secs)
			}

			lastNFailed = latestNFailed
			lastNSuccessful = latestNSuccessful
		}
	}
}

type status string

const (
	statusSuccess = "success"
	statusFailure = "failure"
)

func (tpc *tpsCounter) recordValue(ctx context.Context, latest, previous uint64, status status) (int64, error) {
	if latest < previous {
		return 0, nil
	}

	n := int64(latest - previous)
	if n < 0 {
		// Perhaps we exceeded the uint64 limits then wrapped around, for the latest value.
		// TODO: Perhaps log this?
		return 0, nil
	}

	statusValue := "OK"
	if status == statusFailure {
		statusValue = "ERR"
	}
	ctx, err := tag.New(ctx, tag.Upsert(tagKeyStatus, statusValue))
	if err != nil {
		return 0, err
	}

	stats.Record(ctx, mTransactions.M(n))
	return n, nil
}
