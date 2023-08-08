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
	"net/http"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getStudies returns all studies.
func (r *Routes) getStudies(c *gin.Context) {
	studies, err := r.sl.ListStudies(c.Request.Context())
	if checkError(c, err) {
		return
	}
	c.JSON(http.StatusOK, studies)
}

// getStudyById returns the study with the given ID.
func (r *Routes) getStudyById(c *gin.Context) {
	logger := logging.Extract(c)
	i := c.Param("study_id")
	studyID := parseID(c, i, "study")
	if studyID == (uuid.UUID{}) {
		return
	}
	study, err := r.sl.GetStudy(c.Request.Context(), studyID)
	if checkError(c, err) {
		return
	}
	logger = logger.With("study", study.Name)
	logging.InjectGin(c, logger)
	c.JSON(http.StatusOK, study)
}

// getStudyFiles returns a list of all files the user has that matches the data
// wanted by the specific study.
func (r *Routes) getStudyFiles(c *gin.Context) {
	logger := logging.Extract(c)
	i := c.Param("study_id")
	studyID := parseID(c, i, "study")
	if studyID == (uuid.UUID{}) {
		return
	}
	providerFiles, err := r.sl.ListStudyFiles(c.Request.Context(), studyID)
	if checkError(c, err) {
		return
	}
	logger.Info("Found files", "count", len(providerFiles))
	c.JSON(http.StatusOK, providerFiles)
}
