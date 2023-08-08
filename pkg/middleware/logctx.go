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

// Package middleware provides custom middleware for gin.
package middleware

import (
	"log/slog"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/gin-gonic/gin"
)

// LogContext injects the logger into the gin context.
func LogContext(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqLogger := logger.With(
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
		)
		logging.InjectGin(c, reqLogger)
		c.Next()
	}
}
