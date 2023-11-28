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

package servicebase

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vs49688/servicebase/multilistener"
	"google.golang.org/grpc"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
)

type LogFormat string

const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

// FileMode is a wrapper for fs.FileMode that supports serialisation
type FileMode fs.FileMode

func (mode *FileMode) UnmarshalText(data []byte) error {
	perms, err := strconv.ParseInt(string(data), 8, 32)
	if err != nil {
		return fmt.Errorf("invalid socket permissions: %v", string(data))
	}

	*mode = FileMode(perms)
	return nil
}

func (mode FileMode) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(mode), 8)), nil
}

type HealthCheckable interface {
	GetHealth(ctx context.Context) (*GetHealthResponse, error)
}

type Service interface {
	HealthCheckable

	Close(ctx context.Context) error
}

type ServiceParameters struct {
	Logger  *slog.Logger
	Metrics Metrics

	// ServiceRouter is the top-level HTTP router, without the path prefix applied.
	ServiceRouter *mux.Router

	// ApplicationRouter is the application-level HTTP router, with the path prefix applied.
	ApplicationRouter *mux.Router

	// GRPCRegistrar is the GRPC service registrar.
	GRPCRegistrar grpc.ServiceRegistrar
}

type ServiceFactory func(ctx context.Context, params ServiceParameters) (Service, error)

type serviceBase struct {
	logger            *slog.Logger
	multiListener     *multilistener.MultiListener
	metrics           Metrics
	serviceRouter     *mux.Router
	applicationRouter *mux.Router
	httpServer        *http.Server
	grpcServer        *grpc.Server
	svc               Service
}
