package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"
)

func TestTPSCounter(t *testing.T) {
	t.Skip("FIXME: non deterministic")

	buf := new(bytes.Buffer)
	wlog := &writerLogger{w: buf}
	tpc := newTPSCounter(wlog)
	tpc.reportPeriod = 5 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	go tpc.start(ctx) //nolint:errcheck

	// Concurrently increment the counter.
	n := 50
	repeat := 5
	go func() {
		defer cancel()
		for i := 0; i < repeat; i++ {
			for j := 0; j < n; j++ {
				if j&1 == 0 {
					tpc.incrementSuccess()
				} else {
					tpc.incrementFailure()
				}
			}
			<-time.After(tpc.reportPeriod)
		}
	}()

	<-ctx.Done()
	<-tpc.doneCh

	// We expect that the TPS reported will be:
	// 100 / 5ms => 100 / 0.005s = 20,000 TPS
	lines := strings.Split(buf.String(), "\n")
	require.True(t, len(lines) > 1, "Expected at least 1 line")
	wantReg := regexp.MustCompile(`Transactions per second tps \d+\.\d+`)
	matches := wantReg.FindAllString(buf.String(), -1)
	require.Equal(t, 5, len(matches))
	wantTotalTPS := float64(len(matches)) * float64(n) / (float64(tpc.reportPeriod) / float64(time.Second))
	require.Equal(t, wantTotalTPS, wlog.nTotalTPS, "Mismatched total TPS")
}

type writerLogger struct {
	nTotalTPS float64
	mu        sync.Mutex
	w         io.Writer
	log.Logger
}

var _ log.Logger = (*writerLogger)(nil)

func (wl *writerLogger) Info(msg string, keyVals ...interface{}) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	wl.nTotalTPS += keyVals[1].(float64)
	fmt.Fprintf(wl.w, msg+" "+fmt.Sprintf("%s %.2f\n", keyVals[0], keyVals[1]))
}
