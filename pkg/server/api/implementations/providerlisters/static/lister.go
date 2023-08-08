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

// Package static contains a static provider lister implementation, made for basic testing.
package static

import (
	"context"
	"fmt"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer(
		"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/providersearcher/static",
	)
}

// ProviderLister doesn't do anything, it just returns static data.
type ProviderLister struct{}

func New() *ProviderLister {
	return &ProviderLister{}
}

// ListProviders returns all providers, in this case a static list.
func (pl *ProviderLister) ListProviders(ctx context.Context) ([]types.Provider, error) {
	logger := logging.Extract(ctx)
	logger.Info("Listing providers")
	_, span := tracer.Start(ctx, "StaticConnector.ListProviders")
	defer span.End()
	return []types.Provider{
		{
			ID:                 "0E1EE0FB-9F9D-45E1-9C22-3F32FA24E0AA",
			Name:               "Example Provider",
			Description:        "This is an example provider.",
			LogoURI:            "https://example.org/logo.png",
			ContactInformation: "Example Street 1, 12345 Example City",
			ProviderUrl:        "http://localhost:8080",
			PublicKey:          "An RSA public key for the provider",
			MetadataKey:        "A metadata key",
		},
	}, nil
}

// GetProvider returns the provider with the given ID.
func (pl *ProviderLister) GetProvider(ctx context.Context, providerID string) (types.Provider, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting single provider")
	ctx, span := tracer.Start(ctx, "StaticConnector.GetProvider")
	defer span.End()
	providers, err := pl.ListProviders(ctx)
	if err != nil {
		return types.Provider{}, err
	}
	for _, p := range providers {
		if p.ID == providerID {
			return p, nil
		}
	}
	return types.Provider{}, fmt.Errorf("%w: provider %s not found", types.ErrNotFound, providerID)
}

func (pl *ProviderLister) GetProviderURL(ctx context.Context, providerID string) (string, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting provider URL")
	_, span := tracer.Start(ctx, "SimpleProviderLister.GetProviderURL")
	defer span.End()
	return "https://example.org/provider/", nil
}
