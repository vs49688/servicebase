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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestListen(t *testing.T) {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.TraceLevel)
	xx := New(logger)
	defer func() {
		err := xx.Close()
		assert.NoError(t, err)
	}()

	srv1 := &http.Server{}
	cfg1 := &ListenConfig{
		BindAddress:       "127.0.0.1:50052",
		BindNetwork:       "tcp",
		SocketPermissions: 0600,
	}

	srv2 := &http.Server{}
	cfg2 := &ListenConfig{
		BindAddress:       "127.0.0.1:50053",
		BindNetwork:       "tcp",
		SocketPermissions: 0600,
	}

	srv3 := grpc.NewServer()
	cfg3 := &ListenConfig{
		BindAddress:       "127.0.0.1:50054",
		BindNetwork:       "tcp",
		SocketPermissions: 0600,
	}

	var err error

	err = xx.ListenHTTP(cfg1, srv1)
	require.NoError(t, err)

	err = xx.ListenHTTP(cfg2, srv2)
	require.NoError(t, err)

	err = xx.ListenGRPC(cfg3, srv3)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		t.Log("serve started")
		err := xx.Serve(ctx)
		t.Logf("serve stopped, error: %v", err)
		ch <- err
	}()

	err = <-ch

	require.NoError(t, err)
}

func TestListenWithoutServe(t *testing.T) {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.TraceLevel)
	xx := New(logger)
	defer func() {
		err := xx.Close()
		assert.NoError(t, err)
	}()

	srv1 := &http.Server{}
	cfg1 := &ListenConfig{
		BindAddress:       "127.0.0.1:50052",
		BindNetwork:       "tcp",
		SocketPermissions: 0600,
	}

	err := xx.ListenHTTP(cfg1, srv1)
	require.NoError(t, err)
}
