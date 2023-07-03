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
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type httpWrapper struct {
	srv      *http.Server
	lis      net.Listener
	logEntry *log.Entry
}

func (w *httpWrapper) Serve(ctx context.Context) error {
	serveChannel := make(chan error, 1)
	go func() {
		serveChannel <- w.srv.Serve(w.lis)
	}()

	var err error

	select {
	case <-ctx.Done():
		// Context cancelled, attempt to clean up gracefully.
		// Calling Shutdown() will cause Serve() to return immediately with http.ErrServerClosed.
		shutdownChannel := make(chan error, 1)
		go func() {
			sdCtx, sdCancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer sdCancel()
			shutdownChannel <- w.srv.Shutdown(sdCtx)
		}()
		err = <-shutdownChannel
	case err = <-serveChannel:
		break
	}

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (w *httpWrapper) Log() *log.Entry {
	return w.logEntry
}

func (w *httpWrapper) Close() error {
	return w.srv.Close()
}

func (l *MultiListener) ListenHTTP(cfg *ListenConfig, srv *http.Server) error {
	lis, err := Listen(cfg, l.logger)
	if err != nil {
		return err
	}

	l.servers = append(l.servers, &httpWrapper{
		srv:      srv,
		lis:      lis,
		logEntry: l.logger.WithField("index", len(l.servers)),
	})
	return nil
}
