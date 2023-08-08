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
package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Type Routes contains all the routes for the API.
type Routes struct {
	am types.AccessManager
	dc types.DataspaceConnector
	pl types.ProviderLister
	sl types.StudyLister
}

// New returns a new Routes instance with the appropriate connectors.
func New(
	ps types.ProviderLister,
	dc types.DataspaceConnector,
	sl types.StudyLister,
) *Routes {
	return &Routes{
		pl: ps,
		dc: dc,
		sl: sl,
	}
}

// AddRoutes adds all routes to the given router group.
func (r *Routes) AddRoutes(rg *gin.RouterGroup) {
	rg.GET("/policies", r.getPolicies)
	rg.POST("/policies", r.postPolicy)
	rg.DELETE("/policies/:policy_id", r.deletePolicy)
	rg.GET("/providers", r.getProviders)
	rg.GET("/providers/:provider_id/files", r.getProviderFiles)
	rg.GET("/providers/:provider_id/files/:file_id", r.getProviderFile)
	rg.GET("/providers/:provider_id/files/:file_id/credentials", r.getDownloadCredentials)
	rg.GET("/studies", r.getStudies)
	rg.GET("/studies/:study_id", r.getStudyById)
	rg.GET("/studies/:study_id/files", r.getStudyFiles)
}

func checkError(c *gin.Context, err error) bool {
	logger := logging.Extract(c)
	if err != nil {
		logger.Error("Backend error", "error", err)
		switch {
		case errors.Is(err, types.ErrNotFound):
			// As this is a defined error, we return the error message to the client.
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Status: "Not found",
				Error:  err.Error(),
			})
		case errors.Is(err, types.ErrInvalid):
			// As this is a defined error, we return the error message to the client.
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Status: "Invalid request",
				Error:  err.Error(),
			})
		case errors.Is(err, types.ErrInvalidCredentials):
			// As this is a defined error, we return the error message to the client.
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Status: "Invalid credentials",
				Error:  err.Error(),
			})
		case errors.Is(err, types.ErrBadGateway):
			// As this is a defined error, we return the error message to the client.
			c.JSON(http.StatusBadGateway, types.ErrorResponse{
				Status: "Upstream service unavailable",
				Error:  err.Error(),
			})
		default:
			// Do we want to return the error to the client?
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Status: "Internal error",
				Error:  "The backend encountered an error",
			})
		}
		return true
	}
	return false
}

func parseID(c *gin.Context, id string, entityName string) uuid.UUID {
	logger := logging.Extract(c)
	u, err := uuid.Parse(id)
	if err != nil {
		logger.Error("Invalid ID", "error", err, "entity_name", entityName)
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Status: "Invalid ID",
			Error:  fmt.Sprintf("The given %s ID is invalid", entityName),
		})
		return uuid.UUID{}
	}
	return u
}
