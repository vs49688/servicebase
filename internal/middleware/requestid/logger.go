// Copyright 2023 Zane van Iperen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requestid

import (
	"context"
	"log/slog"
)

const (
	DefaultLoggerFieldName = "request_id"
)

type slogHandler struct {
	slog.Handler
	field string
}

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	handler := h.Handler

	if id := FromContext(ctx); id != "" {
		handler = h.WithAttrs([]slog.Attr{slog.String(h.field, id)})
	}

	return handler.Handle(ctx, r)
}

func NewLogHandler(field string, h slog.Handler) slog.Handler {
	return &slogHandler{
		Handler: h,
		field:   field,
	}
}
