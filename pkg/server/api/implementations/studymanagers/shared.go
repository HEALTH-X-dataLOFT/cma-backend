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

package studymanagers

import "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"

func GetOrganizations(study Study) []Organization {
	var organizations []Organization
	for _, sci := range study.Contained {
		maybeOrganization, err := sci.AsOrganization()
		if err == nil && maybeOrganization.ResourceType == "Organization" {
			organizations = append(organizations, maybeOrganization)
		}
	}

	return organizations
}

func FindMatchingOrganization(study Study, organizations []Organization) *Organization {
	for _, ap := range study.AssociatedParty {
		for _, c := range ap.Role.Coding {
			if c.Code == "primary-investigator" {
				for _, o := range organizations {
					if o.Name == ap.Name {
						return &o
					}
				}
			}
		}
	}

	return nil
}

func ExtractResearchData(study Study) []types.ResearchData {
	research := make([]types.ResearchData, 0)
	for _, sci := range study.Contained {
		maybePlanDefinition, err := sci.AsPlanDefinition()
		if err == nil && maybePlanDefinition.ResourceType == "PlanDefinition" {
			for _, action := range maybePlanDefinition.Action {
				isCollectInformation := false
				for _, code := range action.Code.Coding {
					if code.Code == "collect-information" {
						isCollectInformation = true
					}
				}
				if !isCollectInformation {
					continue
				}
				for _, output := range action.Output {
					research = append(research, types.ResearchData{
						Name:       output.Title,
						AccessType: types.AccessTypePseudonymized,
					})
				}
			}
		}
	}
	return research
}
