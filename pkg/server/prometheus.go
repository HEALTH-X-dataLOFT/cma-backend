// Copyright 2025 HEALTH-X dataLOFT
//
// Licensed under the European Union Public Licence, Version 1.2 (the
// "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://eupl.eu/1.2/en/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	sloggin "github.com/samber/slog-gin"
)

func runPrometheus(
	ctx context.Context, appRouter *gin.Engine, listenAddr string, port int,
) *http.Server {
	logger := logging.Extract(ctx).With("service", "prometheus", "listen_addr", listenAddr, "port", port)

	metricRouter := gin.New()
	metricRouter.Use(sloggin.New(logging.Extract(ctx)))

	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/metrics")
	m.SetSlowTime(10)
	// use metric middleware without expose metric path
	m.UseWithoutExposingEndpoint(appRouter)
	// set metric path expose to metric router
	m.Expose(metricRouter)
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", listenAddr, port),
		Handler:           metricRouter,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		logger.Info("Starting prometheus server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start prometheus server", "error", err)
		}
	}()

	return srv
}
