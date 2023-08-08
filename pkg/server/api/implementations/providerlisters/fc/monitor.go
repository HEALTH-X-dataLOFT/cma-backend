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

package fc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
)

const (
	participantsPath     = "participants"
	selfDescriptionsPath = "self-descriptions"
	queryPath            = "query"
	participantQuery     = "MATCH (provider:LegalParticipant) <-[:providedBy]- (offer:ServiceOffering) -[:aggregationOf]-> (s:SoftwareResource) <-[:instanceOf]- (rh:InstantiatedVirtualResource) -[:serviceAccessPoint]-> (access:ServiceAccessPoint) WHERE access.name = 'ids' return {provider: provider.legalName, protocol: access.protocol, port: access.port, host: access.host}" //nolint:lll
)

func (pl *ProviderLister) getProviderQueryData(ctx context.Context) ([]FCProviderInfo, error) {
	selfDescriptionsUrl := fmt.Sprintf("%s/%s", pl.catalogURL, selfDescriptionsPath)
	queryUrl := fmt.Sprintf("%s/%s", pl.catalogURL, queryPath)
	logger := logging.Extract(ctx).With("self-descriptions-url", selfDescriptionsUrl, "query-url", queryUrl)
	ctx, span := tracer.Start(ctx, "fcProviderLister.getProviderQueryData")
	defer span.End()

	query := FCQuery{
		Statement: participantQuery,
	}
	jsonBody, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", queryUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	logger.Info("Retrieving provider info")
	client := &http.Client{}
	piResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if piResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error when getting providers: %d", piResp.StatusCode)
	}
	defer piResp.Body.Close()
	body, err := io.ReadAll(piResp.Body)
	if err != nil {
		return nil, err
	}

	pr := &FCProviderResponse{}
	err = json.Unmarshal(body, pr)
	if err != nil {
		return nil, err
	}

	var providerInfoList []FCProviderInfo
	for _, entry := range pr.Items {
		for _, pi := range entry {
			providerInfoList = append(providerInfoList, pi)
		}
	}

	return providerInfoList, nil
}

func (pl *ProviderLister) getParticipants(ctx context.Context) ([]ParticipantInfoWithVP, error) {
	participantsUrl := fmt.Sprintf("%s/%s", pl.catalogURL, participantsPath)
	logger := logging.Extract(ctx).With("participants-url", participantsUrl)
	ctx, span := tracer.Start(ctx, "fcProviderLister.getParticipants")
	defer span.End()
	req, err := http.NewRequestWithContext(ctx, "GET", participantsUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	logger.Info("Retrieving participant info")
	client := &http.Client{}
	piResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if piResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error when getting participants: %d", piResp.StatusCode)
	}
	defer piResp.Body.Close()
	body, err := io.ReadAll(piResp.Body)
	if err != nil {
		return nil, err
	}

	participantInfoResponse := &FCParticipantsResponse{}
	err = json.Unmarshal(body, participantInfoResponse)
	if err != nil {
		return nil, err
	}

	var participants []ParticipantInfoWithVP
	for _, pi := range participantInfoResponse.Items {
		vp := &VerifiablePresentation{}
		err = json.Unmarshal([]byte(pi.SelfDescription), vp)
		if err != nil {
			return nil, err
		}
		participants = append(participants, ParticipantInfoWithVP{
			ID:                     pi.ID,
			Name:                   pi.Name,
			PublicKey:              pi.PublicKey,
			SelfDescription:        pi.SelfDescription,
			VerifiablePresentation: *vp,
		})
	}
	return participants, nil
}

func (pl *ProviderLister) monitorParticipants(ctx context.Context, t *time.Ticker) {
	logger := logging.Extract(ctx).With("monitor_type", "federated catalog provider lister")
	logger.Info("Starting monitor")
	pl.updateParticipants(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Info("Context done, stopping monitor")
			t.Stop()
			logger.Info("Timer stopped")
			return
		case <-t.C:
			pl.updateParticipants(ctx)
		}
	}
}

func (pl *ProviderLister) updateParticipants(ctx context.Context) {
	selfDescriptionsUrl := fmt.Sprintf("%s/%s", pl.catalogURL, selfDescriptionsPath)
	queryUrl := fmt.Sprintf("%s/%s", pl.catalogURL, queryPath)
	logger := logging.Extract(ctx).With("self-descriptions-url", selfDescriptionsUrl, "query-url", queryUrl)
	ctx, span := tracer.Start(ctx, "fcProviderLister.updateParticipants")
	defer span.End()

	participants, err := pl.getParticipants(ctx)
	if err != nil {
		logger.Error("Failed to get participants", "error", err)
		return
	}

	providerInfoList, err := pl.getProviderQueryData(ctx)
	if err != nil {
		logger.Error("Failed to get provider info", "error", err)
	}
	logger.Info("provider info", "provider-info", providerInfoList)

	receivedProviders, err := normaliseProviders(ctx, participants, providerInfoList)
	if err != nil {
		logger.Error("Error normalising providers", "error", err)
		return
	}
	if err := pl.saveProviders(ctx, receivedProviders); err != nil {
		logger.Error("Error saving providers", "error", err)
	}
}

func (pl *ProviderLister) saveProviders(ctx context.Context, providers []ProviderInfo) error {
	logger := logging.Extract(ctx)
	ctx, span := tracer.Start(ctx, "fcProviderLister.saveProviders")
	defer span.End()

	logger.Info("Saving providers")
	if err := pl.r.Del(ctx, storageKey).Err(); err != nil {
		logger.Error("couldn't clear providers", "error", err)
	}
	for _, p := range providers {
		jd, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("couldn't marshal provider %s: %w", p.ID, err)
		}
		if err := pl.r.HSet(ctx, storageKey, p.ID, jd).Err(); err != nil {
			return fmt.Errorf("couldn't save provider %s: %w", p.ID, err)
		}
	}
	return nil
}

func normaliseProviders(
	ctx context.Context,
	p []ParticipantInfoWithVP,
	piList []FCProviderInfo,
) ([]ProviderInfo, error) {
	logger := logging.Extract(ctx)

	providerInfo := make(map[string]FCProviderInfo)
	for _, data := range piList {
		providerInfo[data.Provider] = data
	}

	n := make([]ProviderInfo, 0)
	for _, rp := range p {
		vc := rp.VerifiablePresentation.VerifiableCredential[0]

		pi, exists := providerInfo[vc.CredentialSubject.LegalName]
		if !exists {
			logger.Info("Could not find matching FCProviderInfo for provider", "provider", vc.CredentialSubject.LegalName)
			continue
		}
		h := sha256.New()
		_, err := h.Write([]byte(vc.CredentialSubject.ID))
		if err != nil {
			return nil, fmt.Errorf("Failed to generate hash from ID %s", vc.CredentialSubject.ID)
		}

		n = append(n, ProviderInfo{
			ID:                   fmt.Sprintf("%x", h.Sum(nil)),
			Name:                 vc.CredentialSubject.LegalName,
			PublicKey:            rp.PublicKey,
			SelfDescription:      rp.SelfDescription,
			VerifiableCredential: vc,
			Host:                 pi.Host,
			Protocol:             pi.Protocol,
			Provider:             pi.Provider,
			Port:                 pi.Port,
		})
	}
	return n, nil
}
