package registry

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// MockRegistryStore is a thread-safe mock implementation of RegistryStore.
type MockRegistryStore struct {
	mu             sync.Mutex
	store          map[string][]byte
	allowOverwrite bool
}

func NewMockRegistryStore() *MockRegistryStore {
	return &MockRegistryStore{
		store: make(map[string][]byte),
	}
}

func (m *MockRegistryStore) Exists(ref string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.store[ref]
	return exists, nil
}

func (m *MockRegistryStore) Put(ref string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.store[ref]; exists && !m.allowOverwrite {
		return errors.New("tag already exists (conflict)")
	}
	m.store[ref] = data
	return nil
}

func TestPush_AlreadyExists(t *testing.T) {
	store := NewMockRegistryStore()
	client, err := NewClient(WithRegistryStore(store))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ref := "localhost:5001/helm-charts/my-chart:1.0.0"
	data := []byte("chart-data")

	// First push should succeed
	err = client.Push(context.Background(), ref, data)
	if err != nil {
		t.Fatalf("expected first push to succeed, got: %v", err)
	}

	// Second push should fail with duplicate error
	err = client.Push(context.Background(), ref, data)
	if err == nil {
		t.Fatal("expected second push to fail, but it succeeded")
	}

	expectedErr := "chart version already exists in the registry"
	if err.Error() != "chart version already exists in the registry: "+ref {
		t.Errorf("expected error containing %q, got %q", expectedErr, err.Error())
	}
}

func TestPush_AllowOverwrite(t *testing.T) {
	store := NewMockRegistryStore()
	store.allowOverwrite = true
	client, err := NewClient(WithRegistryStore(store))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ref := "localhost:5001/helm-charts/my-chart:1.0.0"
	data := []byte("chart-data")

	// First push
	err = client.Push(context.Background(), ref, data)
	if err != nil {
		t.Fatalf("expected first push to succeed, got: %v", err)
	}

	// Second push with allow overwrite should succeed
	err = client.Push(context.Background(), ref, data, WithAllowOverwrite(true))
	if err != nil {
		t.Fatalf("expected second push with overwrite to succeed, got: %v", err)
	}
}

func TestPush_Concurrent(t *testing.T) {
	store := NewMockRegistryStore()
	client, err := NewClient(WithRegistryStore(store))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ref := "localhost:5001/helm-charts/my-chart:1.0.0"
	data := []byte("chart-data")

	const numWorkers = 10
	var wg sync.WaitGroup
	errChan := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- client.Push(context.Background(), ref, data)
		}()
	}

	wg.Wait()
	close(errChan)

	successCount := 0
	failCount := 0

	for err := range errChan {
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful push, got %d", successCount)
	}
	if failCount != numWorkers-1 {
		t.Errorf("expected %d failed pushes, got %d", numWorkers-1, failCount)
	}
}
