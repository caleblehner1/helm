package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/caleblehner1/helm/pkg/registry"
)

// PushConfiguration holds the configuration for the push action.
type PushConfiguration struct {
	RegistryClient *registry.Client
}

// Push is the action for pushing a chart.
type Push struct {
	cfg            *PushConfiguration
	AllowOverwrite bool
}

// NewPush creates a new Push action.
func NewPush(cfg *PushConfiguration) *Push {
	return &Push{
		cfg: cfg,
	}
}

// Run executes the push action.
func (p *Push) Run(ctx context.Context, ref string, data []byte) (string, error) {
	if p.cfg == nil || p.cfg.RegistryClient == nil {
		return "", fmt.Errorf("registry client not configured")
	}

	var opts []registry.PushOpt
	if p.AllowOverwrite {
		opts = append(opts, registry.WithAllowOverwrite(true))
	}

	err := p.cfg.RegistryClient.Push(ctx, ref, data, opts...)
	if err != nil {
		if strings.Contains(err.Error(), "chart version already exists in the registry") {
			// Format cleanly for CLI output
			return "", fmt.Errorf("conflict: %w", err)
		}
		return "", err
	}

	return fmt.Sprintf("Successfully pushed %s to registry", ref), nil
}
