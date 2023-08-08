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

// Package simple contains a simple Study lister implementation, made for basic testing.
package dsp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/studymanagers"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/transfer"
	dspclient "github.com/go-dataspace/run-dsrpc/gen/go/dsp/v1alpha1"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

const (
	pollInterval = 1
	storageKey   = "studies:dsp-studies"
)

func init() {
	tracer = otel.Tracer(
		"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/studiesearcher/simple",
	)
}

// StudyManager talks to the study catalog.
type StudyManager struct {
	dsp dspclient.ClientServiceClient
	uri string
	r   *redis.Client
}

func New(
	ctx context.Context,
	client dspclient.ClientServiceClient,
	studyCatalogBaseUri string,
	redisClient *redis.Client,
) *StudyManager {
	t := time.NewTicker(pollInterval * time.Minute)
	sm := &StudyManager{
		dsp: client,
		uri: studyCatalogBaseUri,
		r:   redisClient,
	}
	go sm.monitorStudies(ctx, t)
	return sm
}

// ListStudies returns all studies.
func (sm *StudyManager) ListStudies(ctx context.Context) ([]types.Study, error) {
	logger := logging.Extract(ctx)
	logger.Info("Listing studies")
	_, span := tracer.Start(ctx, "DspConnector.ListStudies")
	defer span.End()

	data, err := sm.r.Get(ctx, storageKey).Result()
	if err != nil {
		return nil, fmt.Errorf("couldn't get studies from redis: %w", err)
	}

	var returnedStudies []studymanagers.Study
	err = json.Unmarshal([]byte(data), &returnedStudies)
	if err != nil {
		return nil, err
	}
	studies := make([]types.Study, len(returnedStudies))
	for i, s := range returnedStudies {
		organizations := studymanagers.GetOrganizations(s)

		organization := studymanagers.FindMatchingOrganization(s, organizations)
		if organization == nil {
			return nil, fmt.Errorf("no matching organization found for study %s", s.Title)
		}

		studies[i] = types.Study{
			ID:   uuid.MustParse(*s.Id),
			Name: s.Title,
			Organization: types.Organization{
				ID:   uuid.MustParse(*organization.Id),
				Name: organization.Name,
			},
			Description:        s.Description,
			DescriptionSummary: s.DescriptionSummary,
			ResearchData:       studymanagers.ExtractResearchData(s),
		}
	}
	return studies, nil
}

// GetStudy returns the Study with the given ID.
func (sm *StudyManager) GetStudy(ctx context.Context, studyId uuid.UUID) (types.Study, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting single Study")
	ctx, span := tracer.Start(ctx, "SimpleConnector.GetStudy")
	defer span.End()
	studies, err := sm.ListStudies(ctx)
	if err != nil {
		return types.Study{}, err
	}
	for _, s := range studies {
		if s.ID == studyId {
			return s, nil
		}
	}
	return types.Study{}, fmt.Errorf("%w: Study %s not found", types.ErrNotFound, studyId)
}

// ListStudyFiles returns a list of all files the user has that matches the data
// wanted by the specific study.
func (sm *StudyManager) ListStudyFiles(ctx context.Context, studyID uuid.UUID) ([]types.ProviderFile, error) {
	logger := logging.Extract(ctx)
	logger.Info("Listing study files")
	_, span := tracer.Start(ctx, "SimpleConnector.ListStudyFiles")
	defer span.End()

	// NOTE: This is currently not used as the app for now is going to get the files via the provider API.
	return []types.ProviderFile{}, nil
}

func (sm *StudyManager) monitorStudies(ctx context.Context, t *time.Ticker) {
	logger := logging.Extract(ctx).With("monitor_type", "dsp study lister")
	logger.Info("Starting monitor")
	sm.updateStudies(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping monitor")
			t.Stop()
			logger.Info("Timer stopped")
			return
		case <-t.C:
			sm.updateStudies(logging.Inject(ctx, logger))
		}
	}
}

func (sm *StudyManager) updateStudies(ctx context.Context) {
	logger := logging.Extract(ctx)
	logger.Info("Listing studies")
	_, span := tracer.Start(ctx, "DspConnector.ListStudies")
	defer span.End()
	catalogue, err := sm.dsp.GetProviderCatalogue(ctx, &dspclient.GetProviderCatalogueRequest{
		ProviderUri: sm.uri,
	})
	if err != nil {
		logger.Error("failed to retrieve catalogue", "error", err)
	}
	if len(catalogue.Datasets) != 1 {
		logger.Error("catalogue does not contain single dataset", "num_items", len(catalogue.Datasets))
	}
	dlInfo, err := sm.dsp.GetProviderDatasetDownloadInformation(
		ctx,
		&dspclient.GetProviderDatasetDownloadInformationRequest{
			ProviderUrl: sm.uri,
			DatasetId:   catalogue.Datasets[0].Id,
		},
	)
	if err != nil {
		logger.Error("Seems file for download could not be found", "error", err)
	}
	defer sm.dsp.SignalTransferComplete(ctx, &dspclient.SignalTransferCompleteRequest{ //nolint:errcheck
		TransferId: dlInfo.TransferId,
	})

	logger.Info("Got download information", "auth_type", dlInfo.PublishInfo.AuthenticationType)
	body, err := transfer.RetrieveDSPFile(ctx, dlInfo.PublishInfo)
	if err != nil {
		logger.Error("failed to download study information from remote", "error", err)
	}

	if err := sm.r.Set(ctx, storageKey, body, 0).Err(); err != nil {
		panic(err)
	}
}
