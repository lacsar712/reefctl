package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	anchor := a.clk.Now().Add(-3 * time.Minute)
	err = a.ConfirmZoneHold(context.Background(), anchor)
	if err == nil {
		t.Fatal("expected zone hold error")
	}
	if !errors.Is(err, model.ErrZoneHold) {
		t.Fatalf("expected ErrZoneHold, got %v", err)
	}
}
