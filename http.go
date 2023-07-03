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
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"net"
	"net/http"
)

type HTTPHealthStatus string

const (
	HTTPHealthStatusPass = "pass"
	HTTPHealthStatusFail = "fail"
	HTTPHealthStatusWarn = "warn"
)

// https://www.ietf.org/archive/id/draft-inadarei-api-health-check-06.html
type HTTPHealthResponse struct {
	Status HTTPHealthStatus `json:"status"`
	Notes  []string         `json:"notes,omitempty"`
	Output string           `json:"output,omitempty"`
}

const ContentTypeTextPlainUTF8 = "text/plain; charset=utf-8"

func NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func MethodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func MakeStaticHandler(payload []byte, contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	})
}

func listen(cfg *ListenConfig, logger *log.Logger) (net.Listener, error) {
	ln, err := net.Listen(cfg.BindNetwork, cfg.BindAddress)
	if err != nil {
		return nil, err
	}

	if unix, ok := ln.(*net.UnixListener); ok {
		logger.WithField("permissions", fs.FileMode(cfg.SocketPermissions).String()).
			Debug("fixing socket permissions")

		socket, err := unix.File()
		if err != nil {
			logger.WithError(err).Error("unable to retrieve socket")
			return nil, err
		}

		if err := socket.Chmod(fs.FileMode(cfg.SocketPermissions)); err != nil {
			logger.WithError(err).Error("unable to chmod socket")
			return nil, err
		}
	}

	return ln, nil
}

func (r *HTTPHealthResponse) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	b, _ := json.Marshal(r)

	w.Header().Set("Content-Type", "application/health+json")
	w.Header().Set("Cache-Control", "max-age=60")

	switch r.Status {
	case HTTPHealthStatusFail:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusOK)
	}

	_, _ = w.Write(b)
}
