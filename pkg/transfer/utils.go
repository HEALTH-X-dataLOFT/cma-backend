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

// Copyright 2024 excds
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transfer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	dspclient "github.com/go-dataspace/run-dsrpc/gen/go/dsp/v1alpha1"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
)

func SendHTTPRequest(
	ctx context.Context, method string, url *url.URL, reqBody []byte,
) ([]byte, error) {
	logger := logging.Extract(ctx).With("method", method, "target_url", url)
	logger.Debug("Doing HTTP request")
	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		logger.Error("Failed to create request", "err", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request", "err", err)
		return nil, err
	}
	defer resp.Body.Close()
	// In the future we might want to return the reader to handle big bodies.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read body", "err", err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.Error("Received non-200 status code", "status_code", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	return respBody, nil
}

func RetrieveDSPFile(ctx context.Context, publishInfo *dspclient.PublishInfo) ([]byte, error) {
	logger := logging.Extract(ctx).With("method", "GET", "target_url", publishInfo.Url)
	logger.Debug("Doing HTTP request")
	req, err := http.NewRequestWithContext(ctx, "GET", publishInfo.Url, nil)
	if err != nil {
		logger.Error("Failed to create request", "err", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	switch auth := publishInfo.AuthenticationType; auth {
	case dspclient.AuthenticationType_AUTHENTICATION_TYPE_BASIC:
		req.SetBasicAuth(publishInfo.Username, publishInfo.Password)
	case dspclient.AuthenticationType_AUTHENTICATION_TYPE_BEARER:
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", publishInfo.Password))
	case dspclient.AuthenticationType_AUTHENTICATION_TYPE_UNSPECIFIED:
	default:
		panic(fmt.Sprintf("unexpected dspv1alpha1.AuthenticationType: %#v", auth))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request", "err", err)
		return nil, err
	}
	defer resp.Body.Close()
	// In the future we might want to return the reader to handle big bodies.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read body", "err", err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.Error("Received non-200 status code", "status_code", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	return respBody, nil
}
