package servicebase

import (
	"context"

	"git.vs49688.net/zane/servicebase/internal/middleware/requestid"
)

func GetRequestID(ctx context.Context) string {
	return requestid.FromContext(ctx)
}
