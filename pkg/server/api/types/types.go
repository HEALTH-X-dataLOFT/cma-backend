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

// Package types provides types and interfaces for the API and its backends.
package types

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// AccessType is the type of access to a resource.
type AccessType int

const (
	// AccessTypeFull means access to the full resource, including real name.
	AccessTypeFull AccessType = iota
	// AccessTypeAnonymized means access to the anonymized resource.
	AccessTypeAnonymized
	// AccessTypePseudonymized means access to the pseudononymised resource.
	AccessTypePseudonymized
)

// Provider represents a provider of a resource.
type Provider struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	LogoURI              string `json:"logo_uri"`
	ContactInformation   string `json:"contact_information"`
	VerifiableCredential string `json:"verifiable_credential"`
	MetadataKey          string // Only used as S3 object reference
	ProviderUrl          string `json:"provider_url"`
	PublicKey            string `json:"public_key"`
}

// Target represents the target of a policy permission request..
type Target struct {
	ProviderID string    `json:"provider_id"`
	FileID     uuid.UUID `json:"file_id"`
}

// ShareResponse represents the response to a share request.
type ShareResponse struct {
	TTL         int64  `json:"ttl"`
	DownloadURI string `json:"download_uri"`
	BearerToken string `json:"bearer_token"`
}

// ShareRequest represents a request to share a resource.
type ShareRequest struct {
	Target Target `json:"target"`
	Key    string `json:"key"`
}

// Validate checks the validity of the ShareRequest.
// TODO: actually validate the ShareRequest.
func (sr ShareRequest) Validate() error {
	return nil
}

// Policy represents a policy, describing the permission givven to a provider for accessing a
// resource.
type Policy struct {
	ID         uuid.UUID  `json:"id"`
	Target     Target     `json:"target"`
	Assignee   Provider   `json:"assignee"`
	AccessType AccessType `json:"access_type"`
}

// Validate checks the validity of the policy.
// TODO: actually validate the policy.
func (p Policy) Validate() error {
	return nil
}

// ProviderFile represents a file hosted by a provider.
type ProviderFile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CreatedAt   int64    `json:"created_at"`
	MimeType    string   `json:"mime_type"`
	Size        int64    `json:"size"`
	Provider    Provider `json:"provider"`
	Key         string
}

type DownloadCredentials struct {
	AuthenticationType int64  `json:"authentication_type"`
	URL                string `json:"url"`
	Username           string `json:"username"`
	Password           string `json:"password"`
}

// Validate checks the validity of the ProviderFile.
// TODO: actually validate the ProviderFile.
func (pf ProviderFile) Validate() error {
	return nil
}

type UserFile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mime_type"`
}

func (uf *UserFile) ToProviderFile(createdAt int64, size int64, provider Provider, key string) ProviderFile {
	return ProviderFile{
		ID:          uf.ID,
		Name:        uf.Name,
		Description: uf.Description,
		MimeType:    uf.MimeType,
		Size:        size,
		Provider:    provider,
		Key:         key,
	}
}

// ResearchData represents data and the type of access that a study wants to use for its research.
type ResearchData struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	DataType    string     `json:"data_type"`
	AccessType  AccessType `json:"access_type"`
}

// Study represents a study.
type Study struct {
	ID                 uuid.UUID      `json:"id"`
	Organization       Organization   `json:"organization"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	DescriptionSummary string         `json:"description_summary"`
	StudyUri           string         `json:"study_uri"`
	StudyStart         int64          `json:"study_start"`
	StudyEnd           int64          `json:"study_end"`
	ResearchData       []ResearchData `json:"research_data"`
}

type StudySharedFile struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Organization struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// SuccessResponse represents a response containing success information.
type SuccessResponse struct {
	Status string `json:"status"`
}

// ErrorResponse represents a response containing error information.
type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type Metadata struct {
	Provider  Provider   `json:"provider"`
	UserFiles []UserFile `json:"user_files"`
	Prefix    string
}

type S3Connection struct {
	S3Client *minio.Client
	Bucket   string
}

func (s3c *S3Connection) GetBytes(context context.Context, name string) ([]byte, error) {
	reader, err := s3c.S3Client.GetObject(context, s3c.Bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var fileData []byte
	fileData, err = io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return fileData, err
}

func (s3c *S3Connection) GetMetadataObject(context context.Context, name string) (*Metadata, error) {
	reader, err := s3c.S3Client.GetObject(context, s3c.Bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	fileData, err := s3c.GetBytes(context, name)
	if err != nil {
		return nil, err
	}

	var metadata Metadata
	err = json.Unmarshal(fileData, &metadata)

	// Set the provider metadata key to avoid extra lookup
	metadata.Provider.MetadataKey = name

	// Set prefix
	parts := strings.Split(name, "/")

	prefix := parts[0]
	if len(parts) > 1 {
		prefix = strings.Join(parts[:len(parts)-1], "/")
	}
	metadata.Prefix = prefix

	return &metadata, err
}

func (s3c *S3Connection) EnrichUserFile(
	context context.Context,
	userFile UserFile,
	metadata Metadata,
) (ProviderFile, error) {
	key := metadata.Prefix + "/" + userFile.ID
	reader, err := s3c.S3Client.GetObject(context, s3c.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return ProviderFile{}, err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return ProviderFile{}, err
	}

	return userFile.ToProviderFile(
		stat.LastModified.Unix(),
		stat.Size,
		metadata.Provider,
		key,
	), nil
}

func (s3c *S3Connection) GetMetadataObjects(context context.Context) ([]Metadata, error) {
	var object minio.ObjectInfo
	var metadataObjects []Metadata
	opts := minio.ListObjectsOptions{
		Recursive: true,
	}
	for object = range s3c.S3Client.ListObjects(context, s3c.Bucket, opts) {
		if object.Err != nil {
			return nil, object.Err
		}

		parts := strings.Split(object.Key, "/")
		if parts[len(parts)-1] != "metadata.json" {
			continue
		}

		metadata, err := s3c.GetMetadataObject(context, object.Key)
		if err != nil {
			return nil, err
		}

		metadataObjects = append(metadataObjects, *metadata)
	}

	return metadataObjects, nil
}
