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

// Package api contains the API implementation.
package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	mtypes "github.com/HEALTH-X-dataLOFT/cma-backend/mocks/github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/alecthomas/assert/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

//nolint:funlen,lll
func TestAPIRoutes(t *testing.T) {
	type mockParams struct {
		method    string
		arguments []any
		returns   []any
	}
	type mocks struct {
		providerListerParams     []mockParams
		dataspaceConnectorParams []mockParams
		studyListerParams        []mockParams
	}
	type request struct {
		method string
		path   string
		body   []byte
	}
	type expect struct {
		status int
		body   string
	}
	tests := []struct {
		name    string
		request request
		expect  expect
		mocks   mocks
	}{
		{
			name: "TestListProviders",
			request: request{
				method: http.MethodGet,
				path:   "/api/providers",
				body:   nil,
			},
			expect: expect{
				status: http.StatusOK,
				body:   `[{"id":"37737548-2926-4bd9-b2e6-48fa669e31aa","name":"TestProvider","description":"TestDescription","logo_uri":"some_logo","contact_information":"TestContactInformation","verifiable_credential":"very-verifiable-credential","MetadataKey":"","provider_url":"","public_key":""}]`,
			},
			mocks: mocks{
				providerListerParams: []mockParams{
					{
						method:    "ListProviders",
						arguments: []any{mock.Anything},
						returns: []any{
							[]types.Provider{
								{
									ID:                   "37737548-2926-4bd9-b2e6-48fa669e31aa",
									Name:                 "TestProvider",
									Description:          "TestDescription",
									LogoURI:              "some_logo",
									ContactInformation:   "TestContactInformation",
									VerifiableCredential: "very-verifiable-credential",
								},
							},
							nil,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := mtypes.NewMockProviderLister(t)
			ds := mtypes.NewMockDataspaceConnector(t)
			sl := mtypes.NewMockStudyLister(t)
			router := gin.New()
			routes := api.New(pl, ds, sl)
			routes.AddRoutes(router.Group("/api"))

			for _, p := range tt.mocks.providerListerParams {
				pl.On(p.method, p.arguments...).Return(p.returns...)
			}
			for _, p := range tt.mocks.dataspaceConnectorParams {
				ds.On(p.method, p.arguments...).Return(p.returns...)
			}
			for _, p := range tt.mocks.studyListerParams {
				sl.On(p.method, p.arguments...).Return(p.returns...)
			}
			body := bytes.NewReader(tt.request.body)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.request.method, tt.request.path, body)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expect.status, w.Code)
			assert.Equal(t, tt.expect.body, w.Body.String())
			pl.AssertExpectations(t)
		})
	}
}
