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
	"fmt"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type combinedLogRecord struct {
	http.ResponseWriter
	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	httpReferrer          string
	httpUserAgent         string
}

func (r *combinedLogRecord) Log(logger *log.Logger) {
	var httpReferrer string
	if r.httpReferrer == "" {
		httpReferrer = "-"
	} else {
		httpReferrer = r.httpReferrer
	}

	var httpUserAgent string
	if r.httpUserAgent == "" {
		httpUserAgent = "-"
	} else {
		httpUserAgent = r.httpUserAgent
	}

	_, tzsec := r.time.Zone()
	tzmin := tzsec / 60
	tzhour := tzmin / 60

	tzmin = tzmin % 60
	tzhour = tzhour % 60

	var tzsign string
	if tzhour < 0 {
		tzsign = "-"
	} else {
		tzsign = "+"
	}

	timeString := fmt.Sprintf("%04d/%02d/%02d:%02d:%02d:%02d %s%02d%02d",
		r.time.Year(), r.time.Month(), r.time.Day(),
		r.time.Hour(), r.time.Minute(), r.time.Second(),
		tzsign, tzhour, tzmin,
	)

	logger.Infof("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
		r.ip, timeString,
		r.method, r.uri, r.protocol,
		r.status, r.responseBytes,
		httpReferrer,
		httpUserAgent,
	)
}

func (r *combinedLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *combinedLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type combinedLoggingHandler struct {
	handler http.Handler
	logger  *log.Logger
}

func NewCombinedLoggingHandler(handler http.Handler, logger *log.Logger) http.Handler {
	return &combinedLoggingHandler{
		handler: handler,
		logger:  logger,
	}
}

func (h *combinedLoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	record := &combinedLogRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		httpReferrer:   r.Header.Get("Referer"),
		httpUserAgent:  r.Header.Get("User-Agent"),
	}

	h.handler.ServeHTTP(record, r)
	record.time = time.Now()
	record.Log(h.logger)
}
