package servicebase

import (
	"context"

	"github.com/vs49688/servicebase/internal/middleware/requestid"
)

func GetRequestID(ctx context.Context) string {
	return requestid.FromContext(ctx)
}

func AsPtr[T any](val T) *T {
	return &val
}
