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

package dspconnector

import (
	"context"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/transfer"
	dspclient "github.com/go-dataspace/run-dsrpc/gen/go/dsp/v1alpha1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer(
		"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/dsconnectors/dspconnector",
	)
}

type DataspaceConnector struct {
	dsp dspclient.ClientServiceClient
	pl  types.ProviderLister
}

func New(client dspclient.ClientServiceClient, pl types.ProviderLister) *DataspaceConnector {
	return &DataspaceConnector{
		dsp: client,
		pl:  pl,
	}
}

// ListProviderFiles returns files.
func (dc *DataspaceConnector) ListProviderFiles(
	ctx context.Context, providerID string,
) ([]types.ProviderFile, error) {
	logger := logging.Extract(ctx)
	ctx, span := tracer.Start(ctx, "dspconnector.DataspaceConnector.ListProviderFiles")
	defer span.End()

	provider, err := dc.pl.GetProvider(ctx, providerID)
	if err != nil {
		return nil, convertError(ctx, err)
	}
	logger.Info("Listing files at provider", "provider", provider.Name)

	catalogue, err := dc.dsp.GetProviderCatalogue(ctx, &dspclient.GetProviderCatalogueRequest{
		ProviderUri: provider.ProviderUrl,
	})
	if err != nil {
		return nil, convertError(ctx, err)
	}

	providerFiles := make([]types.ProviderFile, len(catalogue.Datasets))
	for i, ds := range catalogue.Datasets {
		var issued int64
		if is := ds.GetIssued(); is != nil {
			issued = is.AsTime().Unix()
		}
		f := types.ProviderFile{
			ID:   ds.Id,
			Name: ds.Title,
			// Description: ds.Description,
			CreatedAt: issued,
			MimeType:  ds.GetMediaType(),
			Provider:  provider,
		}
		providerFiles[i] = f
	}

	return providerFiles, nil
}

// GetProviderFileInfo returns file info.
func (dc *DataspaceConnector) GetProviderFileInfo(
	ctx context.Context, providerID string, fileID string,
) (types.ProviderFile, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting file info")
	ctx, span := tracer.Start(ctx, "dspconnector.DataspaceConnector.GetProviderFileInfo")
	defer span.End()
	files, err := dc.ListProviderFiles(ctx, providerID)
	if err != nil {
		return types.ProviderFile{}, convertError(ctx, err)
	}
	for _, f := range files {
		if f.ID == fileID {
			return f, nil
		}
	}
	return types.ProviderFile{}, types.ErrNotFound
}

// GetProviderFile returns file with the given ID hosted by the given provider.
func (dc *DataspaceConnector) GetProviderFile(ctx context.Context, providerID string, fileID string,
) ([]byte, error) {
	logger := logging.Extract(ctx)
	logger.Info("Downloading file")
	ctx, span := tracer.Start(ctx, "dspconnector.DataspaceConnector.GetProviderFile")
	defer span.End()

	provider, err := dc.pl.GetProvider(ctx, providerID)
	if err != nil {
		return nil, convertError(ctx, err)
	}

	dlInfo, err := dc.dsp.GetProviderDatasetDownloadInformation(
		ctx,
		&dspclient.GetProviderDatasetDownloadInformationRequest{
			ProviderUrl: provider.ProviderUrl,
			DatasetId:   fileID,
		},
	)
	if err != nil {
		logger.Error("Seems file for download could not be found", "error", err)
		return nil, types.ErrNotFound
	}
	defer dc.dsp.SignalTransferComplete(ctx, &dspclient.SignalTransferCompleteRequest{ //nolint:errcheck
		TransferId: dlInfo.TransferId,
	})

	logger.Info("Got download information", "auth_type", dlInfo.PublishInfo.AuthenticationType)

	return transfer.RetrieveDSPFile(ctx, dlInfo.PublishInfo)
}

func (dc *DataspaceConnector) GetDownloadCredentials(
	ctx context.Context, providerID string, fileID string,
) (types.DownloadCredentials, error) {
	logger := logging.Extract(ctx)
	logger.Info("Retrieving download credentials")
	ctx, span := tracer.Start(ctx, "dspconnector.DataspaceConnector.GetProviderFile")
	defer span.End()

	provider, err := dc.pl.GetProvider(ctx, providerID)
	if err != nil {
		return types.DownloadCredentials{}, convertError(ctx, err)
	}

	dlInfo, err := dc.dsp.GetProviderDatasetDownloadInformation(
		ctx,
		&dspclient.GetProviderDatasetDownloadInformationRequest{
			ProviderUrl: provider.ProviderUrl,
			DatasetId:   fileID,
		},
	)
	if err != nil {
		logger.Error("Seems file for download could not be found", "error", err)
		return types.DownloadCredentials{}, types.ErrNotFound
	}

	logger.Info("Got download information", "auth_type", dlInfo.PublishInfo.AuthenticationType)

	return types.DownloadCredentials{
		AuthenticationType: int64(dlInfo.PublishInfo.AuthenticationType),
		URL:                dlInfo.PublishInfo.Url,
		Username:           dlInfo.PublishInfo.Username,
		Password:           dlInfo.PublishInfo.Password,
	}, nil
}

func convertError(ctx context.Context, err error) error {
	logger := logging.Extract(ctx)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			code := e.Code()
			logger.Info("got status code", "code", code)
			switch code { //nolint
			case codes.Unavailable, codes.Unauthenticated, codes.PermissionDenied:
				return types.ErrInvalidCredentials
			case codes.NotFound:
				return types.ErrNotFound
			default:
				return err
			}
		} else {
			logger.Info("not able to parse error returned", "error", err)
		}
	}
	return err
}
