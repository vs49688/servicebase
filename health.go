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
	"log/slog"
	"net/http"
)

type HealthStatus string

const (
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type GetHealthResponse struct {
	Status       HealthStatus                  `json:"status"`
	Message      string                        `json:"message"`
	Dependencies map[string]*GetHealthResponse `json:"dependencies"`
}

func healthStatusToHTTP(r *GetHealthResponse) HTTPHealthResponse {
	hr := HTTPHealthResponse{}
	switch r.Status {
	case HealthStatusUnhealthy:
		hr.Status = HTTPHealthStatusFail
	case HealthStatusDegraded, HealthStatusUnknown:
		hr.Status = HTTPHealthStatusWarn
	default:
		hr.Status = HTTPHealthStatusPass
	}

	if r.Message != "" {
		hr.Notes = []string{r.Message}
	}

	return hr
}

func checkHealth(impl Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var hr HTTPHealthResponse

		r, err := impl.GetHealth(req.Context())
		if err != nil || r == nil {
			logger.Error("health check failed", slog.Any("error", err))
			hr = HTTPHealthResponse{
				Status: HTTPHealthStatusFail,
				Notes:  []string{"health check failed"},
				Output: err.Error(), // ehhhhh
			}
		} else {
			hr = healthStatusToHTTP(r)
		}

		hr.ServeHTTP(w, req)
	}
}
