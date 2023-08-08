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

package api

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

// GetProviders returns all providers.
func (r *Routes) getProviders(c *gin.Context) {
	providers, err := r.pl.ListProviders(c.Request.Context())
	if checkError(c, err) {
		return
	}
	c.JSON(http.StatusOK, providers)
}

// getProviderFiles returns an index of all files the user has access to on a provider.
func (r *Routes) getProviderFiles(c *gin.Context) {
	provider := r.getProviderByID(c)
	if provider == (types.Provider{}) {
		return
	}
	files, err := r.dc.ListProviderFiles(c.Request.Context(), provider.ID)
	if checkError(c, err) {
		return
	}
	c.JSON(http.StatusOK, files)
}

// getProviderFile returns the file with the given ID hosted by the given provider.
func (r *Routes) getProviderFile(c *gin.Context) {
	// This might be a bit expensive, as we really only need to verify the provider ID.
	// Might be best if the backend would cache provider listings in general, then this
	// wouldn't be much more expensive than just checking the ID.
	provider := r.getProviderByID(c)
	if provider == (types.Provider{}) {
		return
	}
	// Extract logger after getProvider, as it injects the provider name into the logger.
	i := c.Param("file_id")
	if i == "" {
		return
	}

	fileInfo, err := r.dc.GetProviderFileInfo(c.Request.Context(), provider.ID, i)
	if checkError(c, err) {
		return
	}

	fileContents, err := r.dc.GetProviderFile(c.Request.Context(), provider.ID, i)
	if checkError(c, err) {
		return
	}
	mimeTypeData := fileInfo.MimeType
	if mimeTypeData == "" {
		mimeTypeData = mimetype.Detect(fileContents).String()
	}

	c.DataFromReader(
		http.StatusOK, int64(len(fileContents)), mimeTypeData, bytes.NewReader(fileContents),
		map[string]string{"Content-Disposition": fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name)},
	)
}

func (r *Routes) getDownloadCredentials(c *gin.Context) {
	// This might be a bit expensive, as we really only need to verify the provider ID.
	// Might be best if the backend would cache provider listings in general, then this
	// wouldn't be much more expensive than just checking the ID.
	provider := r.getProviderByID(c)
	if provider == (types.Provider{}) {
		return
	}
	// Extract logger after getProvider, as it injects the provider name into the logger.
	i := c.Param("file_id")
	if i == "" {
		return
	}

	downloadCredentials, err := r.dc.GetDownloadCredentials(c.Request.Context(), provider.ID, i)
	if checkError(c, err) {
		return
	}

	c.JSON(http.StatusOK, downloadCredentials)
}

func (r *Routes) getProviderByID(c *gin.Context) types.Provider {
	logger := logging.Extract(c)
	id := c.Param("provider_id")
	if len(id) == 0 {
		return types.Provider{}
	}
	provider, err := r.pl.GetProvider(c.Request.Context(), id)
	if checkError(c, err) {
		return types.Provider{}
	}
	logger = logger.With("provider", provider.Name)
	logging.InjectGin(c, logger)
	return provider
}
