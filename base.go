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
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sebest/xff"
	log "github.com/sirupsen/logrus"

	"github.com/vs49688/servicebase/internal/middleware/combinedlog"
	"github.com/vs49688/servicebase/multilistener"
)

func closeService(ctx context.Context, impl Service, timeout time.Duration, logger *log.Logger) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := impl.Close(ctx); err != nil {
		logger.WithError(err).Error("service cleanup failed")
	}
}

func RunService(ctx context.Context, cfg ServiceConfig, factory ServiceFactory) error {
	sw := &serviceBase{}

	sw.logger = log.New()
	if lvl, err := log.ParseLevel(cfg.LogLevel); err == nil {
		sw.logger.SetLevel(lvl)
	}

	sw.multiListener = multilistener.New(sw.logger)
	defer func() {
		if err := sw.multiListener.Close(); err != nil {
			sw.logger.WithError(err).Error("error_closing_listeners")
		}
	}()

	if cfg.LogFormat == LogFormatJSON {
		sw.logger.SetFormatter(&log.JSONFormatter{})
	} else {
		sw.logger.SetFormatter(&log.TextFormatter{})
	}

	// Create our internal, top-level router
	sw.serviceRouter = mux.NewRouter()
	sw.serviceRouter.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	sw.serviceRouter.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedHandler)

	// Register /metrics
	metrics, metricsHandler, err := configureMetrics(sw.logger)
	if err != nil {
		return err
	}

	sw.metrics = metrics

	// Create the default handler chain, in reverse order
	// 1. XFF handling
	// 2. Logging
	// 3. Metrics
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sw.metrics.RecordHTTPRequest(req)
		sw.serviceRouter.ServeHTTP(w, req)
	})

	handler = combinedlog.NewHandler(handler, sw.logger)

	if !cfg.HTTP.DisableXFF {
		xfff, err := xff.New(xff.Options{AllowedSubnets: nil, Debug: false})
		if err != nil {
			sw.logger.WithError(err).Error("xff_creation_failed")
			return err
		}

		handler = xfff.Handler(handler)
	}

	sw.httpServer = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
	}

	// Create the application-level router
	if cfg.HTTP.PathPrefix == "" {
		cfg.HTTP.PathPrefix = "/"
	}

	pathPrefix := path.Clean(cfg.HTTP.PathPrefix)
	if pathPrefix == "/" {
		sw.applicationRouter = sw.serviceRouter // On your own head be it
	} else {
		sw.applicationRouter = sw.serviceRouter.PathPrefix(pathPrefix).Subrouter()
		sw.applicationRouter.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
		sw.applicationRouter.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedHandler)
	}

	// Register the /health and /metrics endponts. This must be done before the
	// service factory is called, so they can't override them. checkHealth() requires
	// a Service instance, so reserve just the route.
	if !cfg.HTTP.DisableMetrics {
		sw.serviceRouter.Handle("/metrics", metricsHandler).Methods(http.MethodGet)
	}

	var healthRouter *mux.Route
	if !cfg.HTTP.DisableHealth {
		healthRouter = sw.serviceRouter.Path("/health").Methods(http.MethodGet)
	}

	// Create the GRPC server
	sw.grpcServer, err = createGRPCServer(&cfg.GRPC, sw.metrics.Registry)
	if err != nil {
		log.WithError(err).Error("error creating grpc server")
		return err
	}

	// Finally, create the service itself
	svc, err := factory(ctx, ServiceParameters{
		Logger:            sw.logger,
		Metrics:           sw.metrics,
		ServiceRouter:     sw.serviceRouter,
		ApplicationRouter: sw.applicationRouter,
		GRPCRegistrar:     sw.grpcServer,
	})
	if err != nil {
		return err
	}

	sw.svc = svc
	defer closeService(ctx, svc, cfg.ShutdownTimeout, sw.logger)

	if healthRouter != nil {
		healthRouter.HandlerFunc(checkHealth(svc, sw.logger))
	}

	// Finally, handle enables.
	// We still create the actual server objects because initialisation code may rely on it.
	if cfg.HTTP.Enabled {
		err = sw.multiListener.ListenHTTP(&multilistener.ListenConfig{
			BindAddress:       cfg.HTTP.BindAddress,
			BindNetwork:       cfg.HTTP.BindNetwork,
			SocketPermissions: fs.FileMode(cfg.HTTP.SocketPermissions),
		}, sw.httpServer)
		if err != nil {
			return err
		}
	}

	if cfg.GRPC.Enabled {
		err = sw.multiListener.ListenGRPC(&multilistener.ListenConfig{
			BindAddress:       cfg.GRPC.BindAddress,
			BindNetwork:       cfg.GRPC.BindNetwork,
			SocketPermissions: fs.FileMode(cfg.GRPC.SocketPermissions),
		}, sw.grpcServer)
		if err != nil {
			return err
		}
	}

	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	doneChan := make(chan error, 1)
	go func() {
		doneChan <- sw.multiListener.Serve(runCtx)
	}()

	for {
		select {
		case sig := <-sigChan:
			sw.logger.WithField("signal", sig.String()).Info("caught_signal")
			cancelRun()

		case err := <-doneChan:
			if err != nil {
				sw.logger.WithError(err).Error("server_termination_error")
			}

			return nil
		}
	}
}
