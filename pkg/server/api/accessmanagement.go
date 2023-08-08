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
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getPolicies returns all policies.
func (r *Routes) getPolicies(c *gin.Context) {
	policies, err := r.am.ListPolicies(c.Request.Context())
	if checkError(c, err) {
		return
	}
	c.JSON(http.StatusOK, policies)
}

// postPolicy posts a new policy.
func (r *Routes) postPolicy(c *gin.Context) {
	var policy types.Policy
	logger := logging.Extract(c)
	if err := c.ShouldBindJSON(&policy); err != nil {
		logger.Error("Could not parse request", "error", err)
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Status: "Could not parse request",
			Error:  "The request could not be parsed",
		})
		return
	}
	if checkError(c, policy.Validate()) {
		return
	}
	if checkError(c, r.am.SubmitPolicy(c.Request.Context(), policy)) {
		return
	}
	c.Status(http.StatusNoContent)
}

// DeletePolicy deletes a policy given its ID.
func (r *Routes) deletePolicy(c *gin.Context) {
	i := c.Param("policy_id")
	policyID := parseID(c, i, "policy")
	if policyID == (uuid.UUID{}) {
		return
	}
	if checkError(c, r.am.DeletePolicy(c.Request.Context(), policyID)) {
		return
	}
	c.Status(http.StatusNoContent)
}
