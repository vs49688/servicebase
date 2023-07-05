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
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	HeaderName      = "X-Request-ID"
	ContextKey      = "servicebase_request_id"
	GRPCMetadataKey = ContextKey
)

type requestIDHandler struct {
	handler http.Handler
}

func NewHandler(handler http.Handler) http.Handler {
	return &requestIDHandler{
		handler: handler,
	}
}

func newID() string {
	return uuid.New().String()
}

func (h *requestIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the user's supplied one use it, otherwise generate one.
	requestID := r.Header.Get(HeaderName)
	if requestID == "" {
		r.Header.Clone()
		requestID = newID()
	}

	ctx := context.WithValue(r.Context(), ContextKey, requestID)

	ww := requestIDWriter{
		ResponseWriter: w,
		requestID:      requestID,
	}
	h.handler.ServeHTTP(&ww, r.WithContext(ctx))
}

func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKey).(string); ok {
		return id
	}
	return ""
}

type requestIDWriter struct {
	http.ResponseWriter
	requestID string
}

func (r *requestIDWriter) WriteHeader(status int) {
	r.Header().Set(HeaderName, r.requestID)
	r.ResponseWriter.WriteHeader(status)
}

func injectRequestIDGRPC(ctx context.Context) context.Context {
	inMeta, _ := metadata.FromIncomingContext(ctx)
	if inMeta == nil {
		inMeta = metadata.MD{}
		ctx = metadata.NewIncomingContext(ctx, inMeta)
	}

	var requestID string
	vals := inMeta.Get(GRPCMetadataKey)
	if len(vals) > 0 {
		requestID = vals[0]
	} else {
		requestID = newID()
		inMeta.Set(GRPCMetadataKey, requestID)
	}

	ctx = context.WithValue(ctx, ContextKey, requestID)

	return metadata.AppendToOutgoingContext(ctx, GRPCMetadataKey, requestID)
}

func UnaryServerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(injectRequestIDGRPC(ctx), req)
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedStream) Context() context.Context {
	return s.ctx
}

func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, &wrappedStream{ServerStream: ss, ctx: injectRequestIDGRPC(ss.Context())})
}
