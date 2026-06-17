package action

import (
	"context"
	"strings"
	"testing"

	"github.com/caleblehner1/helm/pkg/registry"
)

func TestPushAction_ConflictError(t *testing.T) {
	store := registry.NewMockRegistryStore()
	regClient, err := registry.NewClient(registry.WithRegistryStore(store))
	if err != nil {
		t.Fatalf("failed to create registry client: %v", err)
	}

	cfg := &PushConfiguration{
		RegistryClient: regClient,
	}
	pushAct := NewPush(cfg)

	ref := "localhost:5001/helm-charts/my-chart:1.0.0"
	data := []byte("chart-data")

	// First push should succeed
	msg, err := pushAct.Run(context.Background(), ref, data)
	if err != nil {
		t.Fatalf("expected push to succeed, got: %v", err)
	}
	if !strings.Contains(msg, "Successfully pushed") {
		t.Errorf("expected success message, got: %q", msg)
	}

	// Second push should fail with conflict error
	_, err = pushAct.Run(context.Background(), ref, data)
	if err == nil {
		t.Fatal("expected push to fail with conflict, but it succeeded")
	}

	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("expected error to contain 'conflict', got: %v", err)
	}
}
