package cmd

import (
	"context"

	"github.com/terraconstructs/grid/pkg/sdk"
)

// gridAPIClient abstracts the subset of the SDK used by sync so tests can
// substitute a fake implementation.
type gridAPIClient interface {
	GetStateInfo(ctx context.Context, ref sdk.StateReference) (*sdk.StateInfo, error)
	CreateState(ctx context.Context, input sdk.CreateStateInput) (*sdk.State, error)
	UpdateStateLabels(ctx context.Context, input sdk.UpdateStateLabelsInput) (*sdk.UpdateStateLabelsResult, error)
	AddDependency(ctx context.Context, input sdk.AddDependencyInput) (*sdk.AddDependencyResult, error)
	RemoveDependency(ctx context.Context, edgeID int64) error
}

var gridClientFactory = func(ctx context.Context, cfg sessionConfig) (gridAPIClient, error) {
	return newGridClient(ctx, cfg)
}
