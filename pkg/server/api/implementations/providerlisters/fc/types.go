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

//nolint:tagliatelle
package fc

import (
	"time"
)

// SelfDescription is the self-description of a provider.
type SelfDescription struct {
	SelfDescriptionCredential SelfDescriptionCredential `json:"selfDescriptionCredential"`
}

// SelfDescriptionCredential contains a credential subject.
type SelfDescriptionCredential struct {
	CredentialSubject CredentialSubject `json:"credentialSubject"`
}

// CredentialSubject contains the service data and service ids endpoints.
type CredentialSubject struct {
	ServiceDataEndPoint string `json:"gx-service-offering:serviceDataEndPoint"`
	ServiceIdsEndPoint  string `json:"gx-service-offering:serviceIdsEndPoint"`
}

// ProviderInfo as we'll save it for our own reference, without the json in string.
type ProviderInfo struct {
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	PublicKey            string               `json:"publicKey"`
	ProtocolVersion      string               `json:"protocolVersion"`
	SelfDescription      string               `json:"selfDescription"`
	VerifiableCredential VerifiableCredential `json:"verifiableCredential"`
	Host                 string               `json:"host,omitempty"`
	Protocol             string               `json:"protocol,omitempty"`
	Provider             string               `json:"provider,omitempty"`
	Port                 string               `json:"port,omitempty"`
}

type FCQuery struct {
	Statement string `json:"statement"`
}

type VerifiablePresentation struct {
	Context              []string               `json:"@context"`
	ID                   string                 `json:"id"`
	Type                 []string               `json:"type"`
	VerifiableCredential []VerifiableCredential `json:"verifiableCredential"`
	Proof                Proof                  `json:"proof"`
}

type Proof struct {
	Type               string    `json:"type"`
	Created            time.Time `json:"created"`
	ProofPurpose       string    `json:"proofPurpose"`
	ProofValue         string    `json:"proofValue"`
	VerificationMethod string    `json:"verificationMethod"`
	JWS                string    `json:"jws"`
}

type VerifiableCredential struct {
	Context           []string             `json:"@context,omitempty"`
	Type              []string             `json:"type,omitempty"`
	ID                string               `json:"id,omitempty"`
	Issuer            string               `json:"issuer,omitempty"`
	IssuanceDate      time.Time            `json:"issuanceDate,omitempty"`
	ExpirationDate    time.Time            `json:"expirationDate,omitempty"`
	CredentialSubject CredentialSubjectNew `json:"credentialSubject,omitempty"`
	Proof             Proof                `json:"proof,omitempty"`
}

type Address struct {
	CountryCode   string `json:"gx:addressCountryCode,omitempty"`
	Code          string `json:"gx:addressCode,omitempty"`
	StreetAddress string `json:"gx:streetAddress,omitempty"`
	PostalCode    string `json:"gx:postalCode,omitempty"`
	Locality      string `json:"gx:locality,omitempty"`
}

type RegistrationNumber struct {
	ID string `json:"id,omitempty"`
}

type CredentialSubjectNew struct {
	ID                      string             `json:"id,omitempty"`
	Type                    string             `json:"type,omitempty"`
	LegalName               string             `json:"gx:legalName,omitempty"`
	LegalRegistrationNumber RegistrationNumber `json:"gx:legalRegistrationNumber,omitempty"`
	HeadQuarterAddress      Address            `json:"gx:headquarterAddress,omitempty"`
	LegalAddress            Address            `json:"gx:legalAddress,omitempty"`
	TermsAndConditions      string             `json:"gx-terms-and-conditions:gaiaxTermsAndConditions,omitempty"`
}

type ParticipantInfo struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
	// This is actually a verifiable presentation, but it is included as an escaped jsons string.
	SelfDescription string `json:"selfDescription,omitempty"`
}

type ParticipantInfoWithVP struct {
	ID                     string                 `json:"id,omitempty"`
	Name                   string                 `json:"name,omitempty"`
	PublicKey              string                 `json:"publicKey,omitempty"`
	SelfDescription        string                 `json:"selfDescription,omitempty"`
	VerifiablePresentation VerifiablePresentation `json:"verifiableCredential,omitempty"`
}

type FCParticipantsResponse struct {
	TotalCount int64 `json:"totalCount,omitempty"`
	// Items      []FCProviderInfo `json:"items,omitempty"`
	Items []ParticipantInfo `json:"items,omitempty"`
}

type FCProviderResponse struct {
	TotalCount int64 `json:"totalCount,omitempty"`
	// Items      []FCProviderInfo `json:"items,omitempty"`
	Items []map[string]FCProviderInfo `json:"items,omitempty"`
}

type FCSelfDescriptionsPart struct {
	ExpirationDate time.Time `json:"expirationTime,omitempty"`
	Content        string    `json:"content,omitempty"`
	Validators     []string  `json:"validators,omitempty"`
	SDHash         string    `json:"sdHash,omitempty"`
	ID             string    `json:"id,omitempty"`
	Status         string    `json:"status,omitempty"`
	Issuer         string    `json:"issuer,omitempty"`
	ValidatorDids  []string  `json:"validatorDids,omitempty"`
	UploadDateTime time.Time `json:"uploadDatetime,omitempty"`
	StatusDateTime time.Time `json:"statusDatetime,omitempty"`
}

type FCSelfDescriptionsEntry struct {
	Meta    FCSelfDescriptionsPart `json:"meta,omitempty"`
	Content string                 `json:"content,omitempty"`
}

type FCSelfDescriptionResponse struct {
	TotalCount int64                     `json:"totalCount,omitempty"`
	Items      []FCSelfDescriptionsEntry `json:"items,omitempty"`
}

type FCProviderInfo struct {
	Host     string `json:"host,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Provider string `json:"provider,omitempty"`
	Port     string `json:"port,omitempty"`
}
