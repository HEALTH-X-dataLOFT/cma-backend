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

// Package waitgroup contains some waitgroup utility functions to make it possible
// to inject/extract one from a context.
package waitgroup

import (
	"context"
	"sync"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/gin-gonic/gin"
)

type contextKeyType string

const contextKey contextKeyType = "waitgroup"

// Inject will inject a waitgroup into the context.
func Inject(ctx context.Context, wg *sync.WaitGroup) context.Context {
	return context.WithValue(ctx, contextKey, wg)
}

// InjectGin is a helper function to inject a waitgroup into a gin request context.
func InjectGin(c *gin.Context, wg *sync.WaitGroup) {
	c.Request = c.Request.WithContext(Inject(c.Request.Context(), wg))
}

// Extract will extract a waitgroup from the context. If no waitgroup is found, a new one
// will be returned.
func Extract(ctx context.Context) *sync.WaitGroup {
	logger := logging.Extract(ctx)
	ctxVal := ctx.Value(contextKey)
	if c, ok := ctx.(*gin.Context); ok {
		ctxVal = c.Request.Context().Value(contextKey)
	}
	if ctxVal == nil {
		logger.Warn("waitgroup not found in context - creating new one")
		return &sync.WaitGroup{}
	}
	wg, ok := ctxVal.(*sync.WaitGroup)
	if !ok {
		panic("waitgroup in context is not of type *sync.WaitGroup")
	}
	return wg
}
