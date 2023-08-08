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

package types

import (
	"context"

	"github.com/google/uuid"
)

// ProviderLister is an interface for looking up providers.
type ProviderLister interface {
	ListProviders(ctx context.Context) ([]Provider, error)
	GetProvider(ctx context.Context, providerID string) (Provider, error)
	GetProviderURL(ctx context.Context, providerID string) (string, error)
}

// StudyLister is an interface for looking up studies.
type StudyLister interface {
	ListStudies(ctx context.Context) ([]Study, error)
	GetStudy(ctx context.Context, studyID uuid.UUID) (Study, error)
	ListStudyFiles(ctx context.Context, studyID uuid.UUID) ([]ProviderFile, error)
}

// DataspaceConnector is an interface for listing, and receiving data.
type DataspaceConnector interface {
	ListProviderFiles(ctx context.Context, providerID string) ([]ProviderFile, error)
	GetProviderFileInfo(ctx context.Context, providerID string, fileID string) (ProviderFile, error)
	// TODO: This should return a reader, not bytes, as it might be a large file.
	GetProviderFile(ctx context.Context, providerID string, fileID string) ([]byte, error)
	GetDownloadCredentials(ctx context.Context, providerID string, fileID string) (DownloadCredentials, error)
}

// AccessManager is an interface for managing access policies between entities and data.
type AccessManager interface {
	ListPolicies(ctx context.Context) ([]Policy, error)
	SubmitPolicy(ctx context.Context, policy Policy) error
	DeletePolicy(ctx context.Context, policyID uuid.UUID) error
}

// ShareManager is an interfacd for managing sharing between non-dataspace entities and data.
type ShareManager interface {
	SubmitShare(ctx context.Context, share ShareRequest) (ShareResponse, error)
}

// Validator is the interface all validator implementations must implement.
type Validator interface {
	Validate() error
}
