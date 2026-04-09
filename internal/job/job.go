package job

import "context"

// Job defines the contract for any task that can be processed by the pool
type Job interface {
	ID() string
	Type() string
	Execute(ctx context.Context) error
}
