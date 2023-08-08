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

// Package fc contains the interface to the federated catalog.
package fc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	storageKey   = "catalogue:fc"
	pollInterval = 1
)

var tracer trace.Tracer

var ErrProviderExcluded = errors.New("provider excluded via filter")

func init() {
	tracer = otel.Tracer(
		"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/providerlister/fc",
	)
}

type ProviderLister struct {
	r                  *redis.Client
	catalogURL         string
	providerPublicKeys map[string]string
}

// New creates a new federated catalogue provider lister.
func New(
	ctx context.Context,
	redisClient *redis.Client,
	catalogURL string,
	publicKeys map[string]string,
) (*ProviderLister, error) {
	pl := &ProviderLister{
		r:                  redisClient,
		catalogURL:         catalogURL,
		providerPublicKeys: publicKeys,
	}
	t := time.NewTicker(pollInterval * time.Minute)
	go pl.monitorParticipants(ctx, t)
	return pl, nil
}

// ListProviders returns all providers from the cache.
func (pl *ProviderLister) ListProviders(ctx context.Context) ([]types.Provider, error) {
	logger := logging.Extract(ctx)
	logger.Info("Listing providers")
	ctx, span := tracer.Start(ctx, "fcProviderLister.ListProviders")
	defer span.End()
	prov, err := pl.r.HVals(ctx, storageKey).Result()
	if err != nil {
		return nil, fmt.Errorf("couldn't get providers: %w", err)
	}
	providers := make([]types.Provider, 0)
	for _, p := range prov {
		pr, err := pl.convertProvider(p)
		if err != nil {
			if errors.Is(err, ErrProviderExcluded) {
				continue
			}
			return nil, fmt.Errorf("couldn't convert provider: %w", err)
		}
		pr.PublicKey = pl.providerPublicKeys[pr.ProviderUrl]
		if pr.PublicKey == "" {
			pr.PublicKey = "No key found"
		}
		providers = append(providers, pr)
	}
	return providers, nil
}

// GetProvider returns the provider with the given ID.
func (pl *ProviderLister) GetProvider(ctx context.Context, providerID string) (types.Provider, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting single provider")
	ctx, span := tracer.Start(ctx, "fcProviderLister.GetProvider")
	defer span.End()
	p, err := pl.r.HGet(ctx, storageKey, providerID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return types.Provider{}, fmt.Errorf("%w: provider not found", types.ErrNotFound)
		}
		return types.Provider{}, fmt.Errorf("couldn't get provider: %w", err)
	}
	return pl.convertProvider(p)
}

// GetProviderURL returns the provider URL for the given provider ID.
func (pl *ProviderLister) GetProviderURL(ctx context.Context, providerID string) (string, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting provider URL")
	ctx, span := tracer.Start(ctx, "fcProviderLister.GetProviderURL")
	defer span.End()
	p, err := pl.r.HGet(ctx, storageKey, providerID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("%w: provider not found", types.ErrNotFound)
		}
		return "", fmt.Errorf("couldn't get provider: %w", err)
	}
	var prov ProviderInfo
	err = json.Unmarshal([]byte(p), &prov)
	if err != nil {
		return "", fmt.Errorf("couldn't unmarshal provider: %w", err)
	}

	return prov.Host, nil
}

func (pl *ProviderLister) convertProvider(p string) (types.Provider, error) {
	var prov ProviderInfo
	err := json.Unmarshal([]byte(p), &prov)
	if err != nil {
		return types.Provider{}, err
	}
	vc, err := json.Marshal(prov.VerifiableCredential)

	return types.Provider{
		ID:                   prov.ID,
		Name:                 prov.Name,
		Description:          "No description available.",
		LogoURI:              "",
		ContactInformation:   "No contact information available.",
		VerifiableCredential: string(vc),
		ProviderUrl:          fmt.Sprintf("%s://%s", prov.Protocol, prov.Host),
	}, err
}
