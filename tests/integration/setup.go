// Package integration provides integration testing utilities for the GIIA platform.
package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ServiceConfig contains configuration for a microservice.
type ServiceConfig struct {
	Name    string
	HTTPURL string
	GRPCURL string
	Healthy bool
}

// TestEnvironment holds all configuration for the integration test environment.
type TestEnvironment struct {
	AuthService      ServiceConfig
	CatalogService   ServiceConfig
	ExecutionService ServiceConfig
	DDMRPService     ServiceConfig
	AnalyticsService ServiceConfig
	AIHubService     ServiceConfig

	// Infrastructure
	PostgresURL string
	RedisURL    string
	NATSUrl     string

	// Shared secrets
	JWTSecret string
	JWTIssuer string

	// Test context
	Ctx    context.Context
	Cancel context.CancelFunc
}

// DefaultTestEnvironment returns the default test environment configuration.
func DefaultTestEnvironment() *TestEnvironment {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	return &TestEnvironment{
		AuthService: ServiceConfig{
			Name:    "auth-service",
			HTTPURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8183"),
			GRPCURL: getEnv("AUTH_SERVICE_GRPC_URL", "localhost:9191"),
		},
		CatalogService: ServiceConfig{
			Name:    "catalog-service",
			HTTPURL: getEnv("CATALOG_SERVICE_URL", "http://localhost:8182"),
			GRPCURL: getEnv("CATALOG_SERVICE_GRPC_URL", "localhost:9182"),
		},
		ExecutionService: ServiceConfig{
			Name:    "execution-service",
			HTTPURL: getEnv("EXECUTION_SERVICE_URL", "http://localhost:8184"),
			GRPCURL: getEnv("EXECUTION_SERVICE_GRPC_URL", "localhost:9192"),
		},
		DDMRPService: ServiceConfig{
			Name:    "ddmrp-engine-service",
			HTTPURL: getEnv("DDMRP_SERVICE_URL", "http://localhost:8185"),
			GRPCURL: getEnv("DDMRP_SERVICE_GRPC_URL", "localhost:9193"),
		},
		AnalyticsService: ServiceConfig{
			Name:    "analytics-service",
			HTTPURL: getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8186"),
			GRPCURL: getEnv("ANALYTICS_SERVICE_GRPC_URL", "localhost:9194"),
		},
		AIHubService: ServiceConfig{
			Name:    "ai-intelligence-hub",
			HTTPURL: getEnv("AI_HUB_SERVICE_URL", "http://localhost:8187"),
			GRPCURL: getEnv("AI_HUB_SERVICE_GRPC_URL", "localhost:9195"),
		},
		PostgresURL: getEnv("POSTGRES_URL", "postgres://giia_test:giia_test_password@localhost:5433/giia_test?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://:giia_test_redis_password@localhost:6380/0"),
		NATSUrl:     getEnv("NATS_URL", "nats://localhost:4223"),
		JWTSecret:   getEnv("JWT_SECRET", "test_jwt_secret_for_integration_testing_min_32_chars"),
		JWTIssuer:   getEnv("JWT_ISSUER", "giia-auth-service"),
		Ctx:         ctx,
		Cancel:      cancel,
	}
}

// Setup initializes the test environment and waits for all services to be healthy.
func (env *TestEnvironment) Setup() error {
	fmt.Println("🚀 Setting up integration test environment...")

	// Wait for all services to be healthy
	services := []ServiceConfig{
		env.AuthService,
		env.CatalogService,
		env.ExecutionService,
		env.DDMRPService,
		env.AnalyticsService,
		env.AIHubService,
	}

	for i := range services {
		if err := env.waitForService(&services[i]); err != nil {
			return fmt.Errorf("service %s failed to become healthy: %w", services[i].Name, err)
		}
	}

	// Update environment with healthy service status
	env.AuthService = services[0]
	env.CatalogService = services[1]
	env.ExecutionService = services[2]
	env.DDMRPService = services[3]
	env.AnalyticsService = services[4]
	env.AIHubService = services[5]

	fmt.Println("✅ All services are healthy!")
	return nil
}

// waitForService waits for a service to become healthy.
func (env *TestEnvironment) waitForService(svc *ServiceConfig) error {
	fmt.Printf("  ⏳ Waiting for %s...\n", svc.Name)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	healthURL := fmt.Sprintf("%s/health", svc.HTTPURL)

	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		select {
		case <-env.Ctx.Done():
			return env.Ctx.Err()
		default:
		}

		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			svc.Healthy = true
			fmt.Printf("  ✅ %s is healthy\n", svc.Name)
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("service %s did not become healthy within timeout", svc.Name)
}

// WaitForServices waits for specific services to be healthy.
func (env *TestEnvironment) WaitForServices(serviceNames ...string) error {
	services := map[string]*ServiceConfig{
		"auth":      &env.AuthService,
		"catalog":   &env.CatalogService,
		"execution": &env.ExecutionService,
		"ddmrp":     &env.DDMRPService,
		"analytics": &env.AnalyticsService,
		"ai-hub":    &env.AIHubService,
	}

	for _, name := range serviceNames {
		if svc, ok := services[name]; ok {
			if err := env.waitForService(svc); err != nil {
				return err
			}
		}
	}

	return nil
}

// IsServiceHealthy checks if a specific service is healthy.
func (env *TestEnvironment) IsServiceHealthy(serviceName string) bool {
	services := map[string]ServiceConfig{
		"auth":      env.AuthService,
		"catalog":   env.CatalogService,
		"execution": env.ExecutionService,
		"ddmrp":     env.DDMRPService,
		"analytics": env.AnalyticsService,
		"ai-hub":    env.AIHubService,
	}

	if svc, ok := services[serviceName]; ok {
		return svc.Healthy
	}
	return false
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
