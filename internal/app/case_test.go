package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lacsar712/reefctl/internal/config"
	"github.com/lacsar712/reefctl/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	err = a.ValidateRefrigerantHigh(context.Background(), 400.0)
	if err == nil {
		t.Fatal("expected refrigerant high violation")
	}
	if !errors.Is(err, model.ErrRefrigerantHigh) {
		t.Fatalf("expected ErrRefrigerantHigh, got %v", err)
	}
}
