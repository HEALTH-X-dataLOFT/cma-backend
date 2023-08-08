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

// Package server provides the server subcommand.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	dspclient "github.com/go-dataspace/run-dsrpc/gen/go/dsp/v1alpha1"

	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/cli"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/logging"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/middleware"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/middleware/authforwarder"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api"
	dspconnector "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/dsconnectors/dsp"
	fc "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/providerlisters/fc"
	plstatic "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/providerlisters/static"
	sldsp "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/studymanagers/dsp"
	slstatic "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/implementations/studymanagers/static"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	"github.com/HEALTH-X-dataLOFT/cma-backend/pkg/waitgroup"
	"github.com/gin-gonic/gin"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const appName = "cma-backend"

// Command contains all the options for running the server.
type Command struct {
	ListenAddr      string `help:"Listen address" default:"0.0.0.0" env:"LISTEN_ADDR"`
	Port            int    `help:"Listen port" default:"8080" env:"PORT"`
	PrometheusPort  int    `help:"Listen port" default:"8081" env:"PORT"`
	TracingEnabled  bool   `help:"Enable tracing" default:"false" env:"TRACING_ENABLED"`
	TracingEndpoint string `help:"Tracing endpoint as <host>:<port>"  env:"TRACING_ENDPOINT"`

	ProviderLister        string `help:"Provider lister to use" enum:"static,fc" default:"static" env:"PROVIDER_LISTER"` //nolint:lll
	ProviderCatalogURL    string `help:"Link to the federated catalog" default:"" env:"PROVIDER_CATALOG_URL"`
	ProviderPublicKeyFile string `help:"JSON file with map of provider_url -> base64 JWK public key" default:"" env:"PROVIDER_PUBLIC_KEY_FILE"` //nolint:lll

	StudyManager        string `help:"Study manager to use." enum:"static,dsp" default:"static" env:"STUDY_MANAGER"`
	StudyCatalogBaseUri string `help:"Study catalog base URI." default:"https://study.dev-dataloft-ionos.de/api" env:"STUDY_CATALOG_BASE_URI"` //nolint:lll

	RedisHost                  string `help:"Redis host" default:"localhost" env:"REDIS_HOST"`
	RedisPort                  int    `help:"Redis port" default:"6379" env:"REDIS_PORT"`
	RedisPassword              string `help:"Redis password" default:"" env:"REDIS_PASSWORD"`
	RedisDB                    int    `help:"Redis DB" default:"0" env:"REDIS_DB"`
	RedisTLS                   bool   `help:"Redis enable TLS" default:"false" env:"REDIS_TLS"`
	RedisTLSInsecureSkipVerify bool   `help:"Redis skip TLS verification" default:"false" env:"REDIS_TLS_INSECURE_SKIP_VERIFY"` //nolint:lll
	RedisCacheTimeout          int    `help:"Redis cache timeout in minutes" default:"10" env:"REDIS_CACHE_TIMEOUT"`

	RunDspAddress       string `help:"Address of run-dsp GRPC endpoint" default:"" env:"RUNDSP_URL"`
	RunDspInsecure      bool   `help:"RunDsp connection does not use TLS" default:"false" env:"RUNDSP_INSECURE"`
	RunDspCACert        string `help:"Custom CA certificate for rundsp's TLS certificate" env:"RUNDSP_CA"`
	RunDspClientCert    string `help:"Client certificate to use to authenticate with rundsp" env:"RUNDSP_CLIENT_CERT"`
	RunDspClientCertKey string `help:"Key to the client certificate" env:"RUNDSP_CLIENT_CERT_KEY"`

	static bool `kong:"-"`
}

// Run runs the server.
func (c *Command) Run(p cli.Params) error {
	ctx := p.Context()
	logger := logging.Extract(ctx)
	logger.Info("Starting server", "listenAddr", c.ListenAddr, "port", c.Port)
	logger.Info("Starting prometheus metrics", "listenAddr", c.ListenAddr, "port", c.PrometheusPort)

	var wg sync.WaitGroup
	ctx = waitgroup.Inject(ctx, &wg)
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		logger.Info("Cleaning up")
		cancel()
		wg.Wait()
	}()

	if err := initTracer(ctx, c.TracingEnabled, c.TracingEndpoint, appName); err != nil {
		logger.Error("Fatal error, exiting", "error", err)
		return err
	}

	if p.Debug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if c.ProviderLister == "static" && c.StudyManager == "static" {
		c.static = true
	}

	var err error
	redisClient, err := c.getRedisClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	r := getRouter(logger)

	apiRoutes, err := c.getApiRoutes(ctx, redisClient)
	if err != nil {
		return err
	}

	apiRoutes.AddRoutes(r.Group("/api"))

	promSrv := runPrometheus(ctx, r, c.ListenAddr, c.PrometheusPort)
	appSrv := runBackend(ctx, r, c.ListenAddr, c.Port)

	// Quit and cancel the context, else the webservers will take longer to exit, despite the defer also
	// calling cancel.
	<-quit
	cancel()
	if err := appSrv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown app server", "error", err)
		return err
	}
	if err := promSrv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown prometheus server", "error", err)
		return err
	}
	return nil
}

func (c *Command) getApiRoutes(ctx context.Context, redisClient *redis.Client) (*api.Routes, error) {
	pl, err := c.selectProviderLister(ctx, redisClient)
	if err != nil {
		return nil, err
	}

	client, err := c.getDspClient(ctx)
	if err != nil {
		return nil, err
	}
	dc := dspconnector.New(client, pl)

	sl, err := c.selectStudyManager(ctx, redisClient)
	if err != nil {
		return nil, err
	}
	apiRoutes := api.New(pl, dc, sl)
	return apiRoutes, nil
}

func runBackend(ctx context.Context, handler http.Handler, addr string, port int) *http.Server {
	logger := logging.Extract(ctx).With("service", "app", "listen_addr", addr, "port", port)
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", addr, port),
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start app server", "error", err)
		}
	}()
	return srv
}

func getRouter(logger *slog.Logger) *gin.Engine {
	m := ginmetrics.GetMonitor()
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("cma-backend-main"))
	m.Use(r)
	r.Use(sloggin.New(logger))
	r.Use(middleware.LogContext(logger))
	r.Use(authforwarder.HTTPMiddleware())
	return r
}

func (c *Command) selectProviderLister(ctx context.Context, redisClient *redis.Client) (types.ProviderLister, error) {
	logger := logging.Extract(ctx)
	providerKeys, err := loadProviderPublicKeys(ctx, c.ProviderPublicKeyFile)
	if err != nil {
		return nil, err
	}

	switch c.ProviderLister {
	case "static":
		logger.Info("Using static provider lister")
		return plstatic.New(), nil
	case "fc":
		logger.Info("Using federated catalog provider lister")
		return fc.New(
			ctx,
			redisClient,
			c.ProviderCatalogURL,
			providerKeys)
	default:
		return nil, fmt.Errorf("unknown provider lister %s", c.ProviderLister)
	}
}

func (c *Command) getRedisClient(ctx context.Context) (*redis.Client, error) {
	logger := logging.Extract(ctx)
	ops := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort),
		Password: c.RedisPassword,
		DB:       c.RedisDB,
	}

	if c.RedisTLS {
		ops.TLSConfig = &tls.Config{
			InsecureSkipVerify: c.RedisTLSInsecureSkipVerify, //nolint:gosec
		}
	}

	client := redis.NewClient(ops)
	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, err
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		return nil, err
	}

	var err error
	if !c.static {
		err = client.Ping(ctx).Err()
	} else {
		logger.Info("Skipping redis cache")
	}

	return client, err
}

func (c *Command) selectStudyManager(
	ctx context.Context,
	rc *redis.Client,
) (types.StudyLister, error) {
	logger := logging.Extract(ctx)
	switch c.StudyManager {
	case "static":
		logger.Info("Using static study manager")
		return slstatic.New(), nil
	case "dsp":
		client, err := c.getDspClient(ctx)
		if err != nil {
			return nil, err
		}
		return sldsp.New(ctx, client,
			c.StudyCatalogBaseUri, rc), nil
	default:
		return nil, fmt.Errorf("unknown study manager %s", c.StudyManager)
	}
}

func (c *Command) getDspClient(ctx context.Context) (dspclient.ClientServiceClient, error) {
	logger := logging.Extract(ctx)
	tlsCredentials, err := c.loadTLSCredentials()
	if err != nil {
		return nil, err
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.StartCall, grpclog.FinishCall),
	}
	conn, err := grpc.NewClient(
		c.RunDspAddress,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(interceptorLogger(logger), logOpts...),
			authforwarder.UnaryClientInterceptor,
		),
		grpc.WithChainStreamInterceptor(
			grpclog.StreamClientInterceptor(interceptorLogger(logger), logOpts...),
			authforwarder.StreamClientInterceptor,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not connect to run-dsp: %w", err)
	}

	client := dspclient.NewClientServiceClient(conn)
	_, err = client.Ping(ctx, &dspclient.ClientServicePingRequest{})
	if err != nil {
		return nil, fmt.Errorf("could not ping run-dsp: %w", err)
	}
	return client, nil
}

func (c *Command) loadTLSCredentials() (credentials.TransportCredentials, error) {
	if c.RunDspInsecure {
		return insecure.NewCredentials(), nil
	}

	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if c.RunDspCACert != "" {
		pemServerCA, err := os.ReadFile(c.RunDspCACert)
		if err != nil {
			return nil, fmt.Errorf("couldn't read CA file: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pemServerCA) {
			return nil, fmt.Errorf("failed to add server CA certificate")
		}
		config.RootCAs = certPool
	}

	if c.RunDspClientCert != "" {
		clientCert, err := tls.LoadX509KeyPair(c.RunDspClientCert, c.RunDspClientCertKey)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{clientCert}
	}

	return credentials.NewTLS(config), nil
}

func loadProviderPublicKeys(ctx context.Context, publicKeyFile string) (map[string]string, error) {
	logger := logging.Extract(ctx)
	providerKeys := map[string]string{}

	if publicKeyFile != "" {
		logger.Info(fmt.Sprintf("Trying to load provider public keys from file %s", publicKeyFile))
		jsonFile, err := os.Open(publicKeyFile)
		if err != nil {
			return nil, err
		}
		defer jsonFile.Close()
		data, err := io.ReadAll(jsonFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &providerKeys)
		if err != nil {
			return nil, err
		}
	}
	return providerKeys, nil
}

func interceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
