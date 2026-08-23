package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/config"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.RunCompressorStart(ctx)
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()
	<-done
	if a.CompressorFreqPct() > 10 {
		t.Fatalf("compressor start continued after cancel, freq=%.1f", a.CompressorFreqPct())
	}
}
