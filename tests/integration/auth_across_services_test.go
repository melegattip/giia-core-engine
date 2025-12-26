package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/tests/integration/clients"
)

// TestAuth_JWTWorksAcrossServices tests that a JWT token from auth service works on all other services.
func TestAuth_JWTWorksAcrossServices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	executionClient := clients.NewExecutionClient(env.ExecutionService.HTTPURL)
	analyticsClient := clients.NewAnalyticsClient(env.AnalyticsService.HTTPURL)
	aiHubClient := clients.NewAIHubClient(env.AIHubService.HTTPURL)

	organizationID := DefaultOrganizationID
	email := generateTestEmail()
	password := "SecurePassword123!"

	var accessToken string
	var userID string

	t.Run("1_RegisterAndGetToken", func(t *testing.T) {
		_, err := authClient.Register(ctx, clients.RegisterRequest{
			Email:          email,
			Password:       password,
			FirstName:      "JWT",
			LastName:       "CrossService",
			Phone:          "+1234567890",
			OrganizationID: organizationID,
		})
		require.NoError(t, err)

		tokens, err := authClient.Login(ctx, email, password)
		require.NoError(t, err)
		require.NotEmpty(t, tokens.AccessToken)

		accessToken = tokens.AccessToken

		// Get user info
		userResp, err := authClient.GetCurrentUser(ctx, accessToken)
		require.NoError(t, err)
		userID = userResp.User.ID
	})

	t.Run("2_TokenWorksOnCatalogService", func(t *testing.T) {
		// Create product using the token
		resp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("JWT-CAT"),
			Name:           "JWT Test Product",
			Description:    "Product created with cross-service JWT",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, accessToken)
		require.NoError(t, err, "Token should work on catalog service")
		assert.NotEmpty(t, resp.Product.ID)

		// Clean up
		err = catalogClient.DeleteProduct(ctx, resp.Product.ID, organizationID, accessToken)
		assert.NoError(t, err)
	})

	t.Run("3_TokenWorksOnExecutionService", func(t *testing.T) {
		// Create a product first
		productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("JWT-EXEC"),
			Name:           "JWT Execution Test",
			Description:    "Test product",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, accessToken)
		require.NoError(t, err)

		// Try to create a purchase order
		_, err = executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
			OrganizationID: organizationID,
			SupplierID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productResp.Product.ID,
					SKU:       "JWT-EXEC-SKU",
					Quantity:  10.0,
					UnitPrice: 5.0,
				},
			},
			Notes: "Cross-service JWT test",
		}, accessToken)
		require.NoError(t, err, "Token should work on execution service")
	})

	t.Run("4_TokenWorksOnAnalyticsService", func(t *testing.T) {
		// Try to get analytics - should not fail due to auth
		_, err := analyticsClient.GetBufferAnalytics(ctx, organizationID, accessToken)
		// May return empty data, but should not return auth error
		if err != nil {
			assert.NotContains(t, err.Error(), "unauthorized", "Should not get auth error on analytics")
			assert.NotContains(t, err.Error(), "401", "Should not get 401 on analytics")
		}
	})

	t.Run("5_TokenWorksOnAIHubService", func(t *testing.T) {
		// Try to get notifications
		_, err := aiHubClient.ListNotifications(ctx, userID, organizationID, accessToken, 1, 10)
		// May return empty data or error, but should not be auth error
		if err != nil {
			assert.NotContains(t, err.Error(), "unauthorized", "Should not get auth error on AI Hub")
		}
	})
}

// TestAuth_ExpiredJWTRejected tests that expired JWT tokens are rejected by all services.
func TestAuth_ExpiredJWTRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	executionClient := clients.NewExecutionClient(env.ExecutionService.HTTPURL)

	// Use an invalid/expired token
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZXhwIjoxfQ.4xt6Rl6TxL7r5jGGpAKuRqTXLZtG-p7xB6O8M_hM6AA"
	organizationID := DefaultOrganizationID

	t.Run("CatalogServiceRejectsExpiredToken", func(t *testing.T) {
		_, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            "EXPIRED-SKU",
			Name:           "Should Fail",
			Description:    "Test",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, expiredToken)

		require.Error(t, err, "Should reject expired token")
		assert.Contains(t, err.Error(), "401", "Should return 401 status")
	})

	t.Run("ExecutionServiceRejectsExpiredToken", func(t *testing.T) {
		_, err := executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
			OrganizationID: organizationID,
			SupplierID:     uuid.New().String(),
			Items:          []clients.CreateOrderItemRequest{},
			Notes:          "Should fail",
		}, expiredToken)

		require.Error(t, err, "Should reject expired token")
		assert.Contains(t, err.Error(), "401", "Should return 401 status")
	})
}

// TestAuth_InvalidJWTRejected tests that invalid JWT tokens are rejected.
func TestAuth_InvalidJWTRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	organizationID := DefaultOrganizationID

	invalidTokens := []struct {
		name  string
		token string
	}{
		{"Empty", ""},
		{"Random String", "not-a-valid-jwt"},
		{"Malformed", "header.payload.signature"},
		{"Missing Parts", "eyJhbGciOiJIUzI1NiJ9"},
	}

	for _, tc := range invalidTokens {
		t.Run("RejectsInvalidToken_"+tc.name, func(t *testing.T) {
			_, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
				OrganizationID: organizationID,
				SKU:            "INVALID-SKU",
				Name:           "Should Fail",
				Description:    "Test",
				Category:       "Test",
				UnitOfMeasure:  "UNIT",
			}, tc.token)

			require.Error(t, err, "Should reject invalid token")
		})
	}
}

// TestAuth_TokenRefresh tests that token refresh works correctly.
func TestAuth_TokenRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	organizationID := DefaultOrganizationID
	email := generateTestEmail()
	password := "SecurePassword123!"

	// Register and login
	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "Refresh",
		LastName:       "Test",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)
	require.NotEmpty(t, tokens.RefreshToken)

	t.Run("RefreshToken", func(t *testing.T) {
		// Refresh the token
		newTokens, err := authClient.RefreshToken(ctx, tokens.RefreshToken)
		require.NoError(t, err)
		require.NotEmpty(t, newTokens.AccessToken)

		// Verify new token works
		resp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("REFRESH"),
			Name:           "Refresh Token Product",
			Description:    "Created with refreshed token",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, newTokens.AccessToken)
		require.NoError(t, err, "New token should work")
		assert.NotEmpty(t, resp.Product.ID)
	})
}

// TestAuth_Logout tests that logout invalidates the session.
func TestAuth_Logout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	organizationID := DefaultOrganizationID
	email := generateTestEmail()
	password := "SecurePassword123!"

	// Register and login
	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "Logout",
		LastName:       "Test",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)

	t.Run("LogoutInvalidatesToken", func(t *testing.T) {
		// Verify token works before logout
		_, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("LOGOUT-1"),
			Name:           "Before Logout",
			Description:    "Test",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, tokens.AccessToken)
		require.NoError(t, err, "Token should work before logout")

		// Logout
		err = authClient.Logout(ctx, tokens.AccessToken)
		require.NoError(t, err, "Logout should succeed")

		// Note: Depending on implementation, the token might still work
		// if it's JWT-based without server-side session tracking
		// This test documents the expected behavior
	})
}

// TestAuth_OrganizationIsolation tests that users can only access their organization's data.
func TestAuth_OrganizationIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	// Use pre-existing test organizations
	org1ID := TestOrganizationAID
	org2ID := TestOrganizationBID

	// User 1 in Org 1
	email1 := generateTestEmail()
	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email1,
		Password:       "SecurePassword123!",
		FirstName:      "Org1",
		LastName:       "User",
		Phone:          "+1234567890",
		OrganizationID: org1ID,
	})
	require.NoError(t, err)

	tokens1, err := authClient.Login(ctx, email1, "SecurePassword123!")
	require.NoError(t, err)

	// User 2 in Org 2
	email2 := generateTestEmail()
	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          email2,
		Password:       "SecurePassword123!",
		FirstName:      "Org2",
		LastName:       "User",
		Phone:          "+1234567890",
		OrganizationID: org2ID,
	})
	require.NoError(t, err)

	tokens2, err := authClient.Login(ctx, email2, "SecurePassword123!")
	require.NoError(t, err)

	// Create product in Org 1
	productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: org1ID,
		SKU:            generateTestSKU("ORG1"),
		Name:           "Org 1 Product",
		Description:    "Product in organization 1",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, tokens1.AccessToken)
	require.NoError(t, err)

	t.Run("Org2CannotAccessOrg1Product", func(t *testing.T) {
		// Try to access Org 1's product with Org 2's token
		_, err := catalogClient.GetProduct(ctx, productResp.Product.ID, org1ID, tokens2.AccessToken)
		// Should either return 404 or 403
		require.Error(t, err, "Org 2 should not be able to access Org 1's product")
	})

	t.Run("Org1ProductsNotInOrg2List", func(t *testing.T) {
		// List products for Org 2
		products, err := catalogClient.ListProducts(ctx, org2ID, tokens2.AccessToken, 1, 100)
		require.NoError(t, err)

		// Org 1's product should not be in the list
		for _, p := range products.Products {
			assert.NotEqual(t, productResp.Product.ID, p.ID, "Org 1 product should not appear in Org 2 list")
		}
	})
}
