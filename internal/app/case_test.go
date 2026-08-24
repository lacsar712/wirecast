package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lacsar712/wirecast/internal/config"
	"github.com/lacsar712/wirecast/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	err = a.ValidateShellDrift(context.Background(), 25.0)
	if err == nil {
		t.Fatal("expected moisture drift violation")
	}
	if !errors.Is(err, model.ErrShellDrift) {
		t.Fatalf("expected ErrShellDrift, got %v", err)
	}
}
