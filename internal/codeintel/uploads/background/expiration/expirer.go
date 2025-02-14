package expiration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type expirer struct {
	dbStore       DBStore
	metrics       *metrics
	policyMatcher PolicyMatcher
}

var (
	_ goroutine.Handler      = &expirer{}
	_ goroutine.ErrorHandler = &expirer{}
)

func (r *expirer) Handle(ctx context.Context) error {
	if err := r.HandleUploadExpirer(ctx); err != nil {
		return err
	}

	return nil
}

func (r *expirer) HandleError(err error) {
}
