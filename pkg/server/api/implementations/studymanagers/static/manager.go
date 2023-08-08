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

// Package static contains a static Study lister implementation, made for basic testing.
package static

import (
	"context"
	"fmt"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	staticStudies = []types.Study{
		{
			ID: uuid.MustParse("842b90d4-4007-4f67-87ae-301317d728b6"),
			Organization: types.Organization{
				ID:   uuid.MustParse("bdcfb2c3-d787-4915-a990-91559c1685b2"),
				Name: "Example Research Institute",
			},

			Name:               "Example Study",
			Description:        "This is an example study and this is the longer description..",
			DescriptionSummary: "This is an example study.",
			StudyUri:           "https://researchinstitute.edu/ExampleStudy",
			StudyStart:         1706742000000,
			StudyEnd:           1717192800000,
			ResearchData: []types.ResearchData{
				{
					Name:        "Caffeine level",
					Description: "The level of caffeine in the blood of the participants.",
					DataType:    "C8H10N402",
					AccessType:  types.AccessTypeAnonymized,
				},
				{
					Name:        "Heart rate",
					Description: "The heart rate of the participants.",
					DataType:    "STRSS180",
					AccessType:  types.AccessTypeAnonymized,
				},
			},
		},
		{
			ID: uuid.MustParse("333f20fb-c323-46fa-bed8-1656bb2ef613"),
			Organization: types.Organization{
				ID:   uuid.MustParse("4be32995-11f7-4264-b47d-048fbd078e0b"),
				Name: "Another Research Institute",
			},
			Name:               "Another study",
			Description:        "Different study",
			DescriptionSummary: "This is a completely different study than the other one.",
			StudyUri:           "https://researchinstitute.edu/AnotherStudy",
			StudyStart:         1709593200000,
			StudyEnd:           1733353200000,
			ResearchData: []types.ResearchData{
				{
					Name:        "Height",
					Description: "How tall the person is.",
					DataType:    "TALL210",
					AccessType:  types.AccessTypeAnonymized,
				},
				{
					Name:        "Vision",
					Description: "How good the person can see.",
					DataType:    "HINDSIGHT2020",
					AccessType:  types.AccessTypeAnonymized,
				},
			},
		},
	}
	tracer trace.Tracer
)

func init() {
	tracer = otel.Tracer(
		"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/studiesearcher/static",
	)
}

// StudyManager doesn't do anything, it just returns static data.
type StudyManager struct{}

func New() *StudyManager {
	return &StudyManager{}
}

// ListStudies returns all studies, in this case a static list.
func (pl *StudyManager) ListStudies(ctx context.Context) ([]types.Study, error) {
	logger := logging.Extract(ctx)
	logger.Info("Listing studies")
	_, span := tracer.Start(ctx, "StaticConnector.ListStudies")
	defer span.End()
	return staticStudies, nil
}

// GetStudy returns the Study with the given ID.
func (sl *StudyManager) GetStudy(ctx context.Context, studyId uuid.UUID) (types.Study, error) {
	logger := logging.Extract(ctx)
	logger.Info("Getting single Study")
	ctx, span := tracer.Start(ctx, "StaticConnector.GetStudy")
	defer span.End()
	studies, err := sl.ListStudies(ctx)
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
func (pl *StudyManager) ListStudyFiles(ctx context.Context, studyID uuid.UUID) ([]types.ProviderFile, error) {
	logger := logging.Extract(ctx)
	logger.Info("Listing study files")
	_, span := tracer.Start(ctx, "StaticConnector.ListStudyFiles")
	defer span.End()

	mimetype := "application/json"
	starbucks := types.Provider{
		ID:                 "f4d55ac8-e0b6-47c0-81e6-0920594a858f",
		Name:               "Starbucks",
		Description:        "The local caffeine provider.",
		LogoURI:            "https://logo.link/maybe.jpg",
		ContactInformation: "We're at the corner serving expensive 'coffee",
	}

	pizzaHut := types.Provider{
		ID:                 "290d34d7-dc03-4ec5-a052-65f685856350",
		Name:               "Pizza Hut",
		Description:        "The local pizza provider.",
		LogoURI:            "https://logo.link/maybe.jpg",
		ContactInformation: "We're easy to find, just google it",
	}

	return []types.ProviderFile{
		{
			ID:          "842b90d4-4007-4f67-87ae-301317d728b6",
			Name:        "caffeine_levels.json",
			Description: "The level of caffeine in the blood of the patient.",
			CreatedAt:   1706742000000,
			MimeType:    mimetype,
			Size:        1024,
			Provider:    starbucks,
		},
		{
			ID:          "159ef473-3afd-40a4-a657-77b401dee68e",
			Name:        "heart_rate.json",
			Description: "The heart rate of the patient.",
			CreatedAt:   1706742000000,
			MimeType:    mimetype,
			Size:        1024,
			Provider:    starbucks,
		},
		{
			ID:          "333f20fb-c323-46fa-bed8-1656bb2ef613",
			Name:        "colesterol_levels.json",
			Description: "The colesterol levels of the patient.",
			CreatedAt:   1709593200000,
			MimeType:    mimetype,
			Size:        1024,
			Provider:    pizzaHut,
		},
	}, nil
}
