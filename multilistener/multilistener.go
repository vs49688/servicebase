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
	"reflect"

	log "github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

func New(logger *log.Logger) *MultiListener {
	l := &MultiListener{
		logger: logger,
	}

	return l
}

func (l *MultiListener) Serve(ctx context.Context) error {
	// Create a sub-context so we can cancel ourselves.
	ourCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fire up all the servers.
	chans := make([]chan error, len(l.servers))
	allCases := make([]reflect.SelectCase, len(l.servers)+1)
	for i, srv := range l.servers {
		chans[i] = make(chan error, 1)
		allCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(chans[i]),
		}

		go func(i int, w wrapper) {
			chans[i] <- w.Serve(ourCtx)
		}(i, srv)
	}

	contextIndex := len(allCases) - 1
	allCases[contextIndex] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ourCtx.Done()),
	}

	serveCases := allCases[:contextIndex]

	serverErrors := make([]error, len(l.servers))

	numActive := len(l.servers)

	// Pass 1: Wait for a context cancellation, or for one of the servers to die.
	idx, value, _ := reflect.Select(allCases)
	if idx == contextIndex {
		// Context closed, flag and wait for our servers to close.
		l.logger.Trace("context closed")
	} else {
		// One of our servers terminated. Capture its error and kill the rest.
		serverErrors[idx] = l.handleServerTermination(idx, value)
		numActive--
	}

	// Cancel all the servers upon a server error
	cancel()

	l.logger.Trace("waiting for death")

	// Pass 2: Wait for the rest to die.
	for numActive > 0 {
		idx, value, _ := reflect.Select(serveCases)
		serverErrors[idx] = l.handleServerTermination(idx, value)
		numActive--
	}

	return multierr.Combine(serverErrors...)
}

func (l *MultiListener) handleServerTermination(idx int, value reflect.Value) error {
	srv := l.servers[idx]

	if value.IsNil() {
		srv.Log().Info("server terminated")
		return nil
	}

	err := value.Interface().(error)
	srv.Log().Error("server terminated")
	return err
}

func (l *MultiListener) Close() error {
	errs := make([]error, 0, len(l.servers))
	for _, srv := range l.servers {
		err := srv.Close()
		if err != nil {
			srv.Log().WithError(err).Error("error closing server")
		}

		errs = append(errs, err)
	}

	return multierr.Combine(errs...)
}
