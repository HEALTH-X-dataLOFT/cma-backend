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

// Package logging provides logging utilities.
package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

// create a new type for the context key, as context doesn't allow string as the key for
// collision reasons.
type contextKeyType string

const contextKey contextKeyType = "logger"

// New will initialise a new structured logger with JSON output, logging at the desired level.
// If the requested level doesn't exist, it panics.
func NewJSON(requestedLevel string, humanReadable bool) *slog.Logger {
	var level slog.Level
	switch requestedLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		panic("unknown log level")
	}
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, &opts)
	if humanReadable {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource: true,
			Level:     level,
		})
	}
	return slog.New(handler)
}

// Inject will inject a logger into the context.
func Inject(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

// InjectGin is a helper function to inject a logger into a gin request context.
func InjectGin(c *gin.Context, logger *slog.Logger) {
	c.Request = c.Request.WithContext(Inject(c.Request.Context(), logger))
}

// Extract will extract a logger from the context. If no logger is found, a default logger
// with level info is returned.
func Extract(ctx context.Context) *slog.Logger {
	ctxVal := ctx.Value(contextKey)
	if c, ok := ctx.(*gin.Context); ok {
		ctxVal = c.Request.Context().Value(contextKey)
	}
	if ctxVal == nil {
		logger := NewJSON("info", false)
		logger.Warn("logger not found in context, returning default logger with level info")
		return logger
	}
	logger, ok := ctxVal.(*slog.Logger)
	if !ok {
		panic("logger in context is not of type *slog.Logger")
	}
	return logger
}
