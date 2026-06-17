package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrChartExists is returned when the chart version already exists in the registry.
var ErrChartExists = errors.New("chart version already exists in the registry")

// RegistryStore defines the interface to interact with the registry storage.
type RegistryStore interface {
	Exists(ref string) (bool, error)
	Put(ref string, data []byte) error
}

// Client is the registry client.
type Client struct {
	Registry RegistryStore
}

// ClientOpt is a function option for Client.
type ClientOpt func(*Client)

// WithRegistryStore sets the registry store.
func WithRegistryStore(store RegistryStore) ClientOpt {
	return func(c *Client) {
		c.Registry = store
	}
}

// NewClient creates a new Client.
func NewClient(opts ...ClientOpt) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

type pushOpts struct {
	allowOverwrite bool
}

// PushOpt is an option for Push.
type PushOpt func(*pushOpts)

// WithAllowOverwrite allows overwriting an existing tag.
func WithAllowOverwrite(allow bool) PushOpt {
	return func(o *pushOpts) {
		o.allowOverwrite = allow
	}
}

// Push uploads the chart to the registry.
func (c *Client) Push(ctx context.Context, ref string, data []byte, opts ...PushOpt) error {
	options := &pushOpts{}
	for _, opt := range opts {
		opt(options)
	}

	if c.Registry == nil {
		return errors.New("registry store not configured")
	}

	// 1. Pre-flight check: verify if the target chart version tag already exists.
	if !options.allowOverwrite {
		exists, err := c.Registry.Exists(ref)
		if err != nil {
			return fmt.Errorf("failed to check if chart exists: %w", err)
		}
		if exists {
			return fmt.Errorf("%w: %s", ErrChartExists, ref)
		}
	}

	// 2. Perform the push.
	if err := c.Registry.Put(ref, data); err != nil {
		// Handle potential race conditions where the tag is created between the check and the push
		if !options.allowOverwrite && (strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "conflict")) {
			return fmt.Errorf("%w: %s", ErrChartExists, ref)
		}
		return fmt.Errorf("failed to push chart: %w", err)
	}

	return nil
}
