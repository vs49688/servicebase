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

package multilistener

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type grpcWrapper struct {
	srv    *grpc.Server
	lis    net.Listener
	logger *slog.Logger
}

func (w *grpcWrapper) Serve(ctx context.Context) error {
	serveChannel := make(chan error, 1)
	go func() {
		serveChannel <- w.srv.Serve(w.lis)
	}()

	var err error

	select {
	case <-ctx.Done():
		w.srv.GracefulStop()

		// If the context was cancelled due to a deadline (timeout), forcefully stop the server
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			w.srv.Stop()
		}
	case err = <-serveChannel:
		break
	}

	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}

	return err

}

func (w *grpcWrapper) Log() *slog.Logger {
	return w.logger
}

func (w *grpcWrapper) Close() error {
	w.srv.Stop()
	return nil
}

func (l *MultiListener) ListenGRPC(cfg *ListenConfig, srv *grpc.Server) error {
	lis, err := Listen(cfg, l.logger)
	if err != nil {
		return err
	}

	l.servers = append(l.servers, &grpcWrapper{
		srv:    srv,
		lis:    lis,
		logger: l.logger.With(slog.Int("index", len(l.servers))),
	})
	return nil
}
