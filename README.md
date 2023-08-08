# Consent Management App backend

## What is this and what is it for?

The [Consent Management App](https://github.com/HEALTH-X-dataLOFT/consent-managemnet-app) is
the mobile app for interacting with the HEALTH-X dataLOFT dataspace connected
services. Since the [IDSA dataspace protocol](https://docs.internationaldataspaces.org/ids-knowledgebase/dataspace-protocol)
requires both participants in communication to have publicly accessible
endpoints where API callbacks are done, it is not feasible for the mobile
application itself to speak that protocol. And here is where the cma-backend
comes into the picture as an in-between system to simplify the mobile app and
abstract away the listing of participants in the dataspace, studies available
and data that the user owns in the dataspace.

## Dataspace connector service

The dataspace connector used is [RUN-DSP](https://github.com/go-dataspace/run-dsp), which
is a lightweight connector written in go.

## Parts of the system

### Overview

The backend has a client API defined in [OpenAPI](cma_backend_api.yml) and this currently has three main uses:

* Listing providers available in the dataspace
* Interacting with selected providers
* Listing available studies to share information with

It also passes on authentication from the client so that a user can be verified
on the provider side. The client retrieves a JWT from keycloak, wraps that as
a JWE using the provider's public key and passes that to the backend as a
bearer token in the authentication header.

**Note:** There is a static and hardcoded version of both study manager and
provider lister for testing purposes.

### Study manager

The role of the study manager component is to retrieve available studies and to
return this information to the client. Study information is retrieved via a
dataspace connection using RUN-DSP.

### Provider lister

Providers in the dataspace host the data that users can access, and as a user
of the dataspace you need to find which providers are available. Finding the
available providers is done by talking to the federated catalogue that contains
all registered providers.

The provider lister will also provide the public keys for the client to use
for creating the JWE when using authentication. Provided is the bash script
[generate_jwk.sh](bin/generate_jwk.sh) that will create the necessary public/private
keys needed by the provider and client.

### Dataspace connector

This is the "glue" that handles the requests for file listings and transfers
from the client. Communication is done using RUN-DSP and authentication is
passed on using authentication headers.

## Running the backend

### Requirements

* go - For building the application (currently version 1.23.6)
* redis - For storing results from listing studies etc
* public key for the available providers  (required for clients to use end-to-end encryption)
* an instance of [RUN-DSP](https://github.com/go-dataspace/run-dsp)

### Running

See [launch.json.example](.vscode/launch.json.example) as an example on parameters
when running in vscode.

```
$ cma-backend server --help
Usage: cma-backend server

Run server

Flags:
  -h, --help                              Show context-sensitive help.
      --debug
      --log-level="info"                  Set log level ($LOGLEVEL).

      --listen-addr="0.0.0.0"             Listen address ($LISTEN_ADDR)
      --port=8080                         Listen port ($PORT)
      --prometheus-port=8081              Listen port ($PORT)
      --tracing-enabled                   Enable tracing ($TRACING_ENABLED)
      --tracing-endpoint=STRING           Tracing endpoint as <host>:<port> ($TRACING_ENDPOINT)
      --provider-lister="static"          Provider lister to use ($PROVIDER_LISTER)
      --provider-catalog-url=""           Link to the federated catalog ($PROVIDER_CATALOG_URL)
      --provider-public-key-file=""       JSON file with map of provider_url -> base64 JWK public key ($PROVIDER_PUBLIC_KEY_FILE)
      --study-manager="static"            Study manager to use ($STUDY_MANAGER).
      --study-catalog-base-uri="https://study.dev-dataloft-ionos.de/api"
                                          Study catalog base URI ($STUDY_CATALOG_BASE_URI).
      --redis-host="localhost"            Redis host ($REDIS_HOST)
      --redis-port=6379                   Redis port ($REDIS_PORT)
      --redis-password=""                 Redis password ($REDIS_PASSWORD)
      --redis-db=0                        Redis DB ($REDIS_DB)
      --redis-tls                         Redis enable TLS ($REDIS_TLS)
      --redis-tls-insecure-skip-verify    Redis skip TLS verification ($REDIS_TLS_INSECURE_SKIP_VERIFY)
      --redis-cache-timeout=10            Redis cache timeout in minutes ($REDIS_CACHE_TIMEOUT)
      --run-dsp-address=""                Address of run-dsp GRPC endpoint ($RUNDSP_URL)
      --run-dsp-insecure                  RunDsp connection does not use TLS ($RUNDSP_INSECURE)
      --run-dsp-ca-cert=STRING            Custom CA certificate for rundsp's TLS certificate ($RUNDSP_CA)
      --run-dsp-client-cert=STRING        Client certificate to use to authenticate with rundsp ($RUNDSP_CLIENT_CERT)
      --run-dsp-client-cert-key=STRING    Key to the client certificate ($RUNDSP_CLIENT_CERT_KEY)
```

```
$ cma-backend server --provider-lister=fc \
    --provider-catalog-url=<federated catalog url> \
    --provider-public-key-file=<path to public keys json> \
    --study-manager=dsp \
    --study-catalog-base-uri=<study provider url> \
    --run-dsp-address=<run-dsp host>:<port> \
    --run-dsp-ca-cert=<path to CA certificate> \
    --run-dsp-client-cert=<path to certificate> \
    --run-dsp-client-cert-key=<path to certificate key> \
    --redis-host=<redis host>
```

## Generating mocks for tests

When you have updated any of the interfaces you will need to update the mock
implementations of these for the tests to work properly.

Requires [mockery](https://github.com/vektra/mockery) to be installed.

You can (re-)generate mocks (placed in `mocks/*`) by running the following command:

```bash
$ mockery
```
