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

package authforwarder

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type contextKeyType string

const (
	contextKey contextKeyType = "authheader"
)

// HTTPMiddleware injects the authorization header into the context.
func HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authContents := c.Request.Header.Get("Authorization")
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), contextKey, authContents))
		c.Next()
	}
}

// AuthRoundTripper is a http client "middleware" that extracts the auth middleware out of the
// context and injects it into the request.
type AuthRoundTripper struct {
	Proxied http.RoundTripper
}

// Roundtrip does the actual injection.
func (art AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	authVal := ExtractAuthorization(req.Context())
	req.Header.Add("Authorization", authVal)
	return art.Proxied.RoundTrip(req)
}
