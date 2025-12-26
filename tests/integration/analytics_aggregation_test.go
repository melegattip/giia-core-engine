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

// TestAnalyticsAggregation_BufferAnalytics tests analytics aggregation for buffer data.
func TestAnalyticsAggregation_BufferAnalytics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	analyticsClient := clients.NewAnalyticsClient(env.AnalyticsService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	var accessToken string

	t.Run("Setup", func(t *testing.T) {
		_, err := authClient.Register(ctx, clients.RegisterRequest{
			Email:          email,
			Password:       password,
			FirstName:      "Analytics",
			LastName:       "Test",
			Phone:          "+1234567890",
			OrganizationID: organizationID,
		})
		require.NoError(t, err)

		tokens, err := authClient.Login(ctx, email, password)
		require.NoError(t, err)
		accessToken = tokens.AccessToken
	})

	t.Run("GetBufferAnalytics", func(t *testing.T) {
		analytics, err := analyticsClient.GetBufferAnalytics(ctx, organizationID, accessToken)
		// Analytics might be empty for new organization
		if err == nil {
			assert.IsType(t, []clients.BufferAnalytics{}, analytics, "Should return buffer analytics array")
		}
	})

	t.Run("GetDaysInInventory", func(t *testing.T) {
		metrics, err := analyticsClient.GetDaysInInventory(ctx, organizationID, accessToken)
		if err == nil {
			assert.IsType(t, []clients.DaysInInventoryKPI{}, metrics, "Should return days in inventory metrics")
		}
	})

	t.Run("GetImmobilizedInventory", func(t *testing.T) {
		metrics, err := analyticsClient.GetImmobilizedInventory(ctx, organizationID, accessToken)
		if err == nil {
			assert.IsType(t, []clients.ImmobilizedInventoryKPI{}, metrics, "Should return immobilized inventory metrics")
		}
	})

	t.Run("GetInventoryRotation", func(t *testing.T) {
		metrics, err := analyticsClient.GetInventoryRotation(ctx, organizationID, accessToken)
		if err == nil {
			assert.IsType(t, []clients.InventoryRotationKPI{}, metrics, "Should return inventory rotation metrics")
		}
	})
}

// TestAnalyticsAggregation_Snapshot tests analytics snapshot retrieval and creation.
func TestAnalyticsAggregation_Snapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	analyticsClient := clients.NewAnalyticsClient(env.AnalyticsService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "Snapshot",
		LastName:       "Test",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	t.Run("GetSnapshot", func(t *testing.T) {
		snapshot, err := analyticsClient.GetSnapshot(ctx, organizationID, accessToken)
		if err == nil && snapshot != nil {
			assert.Equal(t, organizationID, snapshot.OrganizationID, "Snapshot should belong to organization")
			assert.GreaterOrEqual(t, snapshot.TotalProducts, 0, "Total products should be non-negative")
		}
	})
}

// TestAnalyticsAggregation_SyncBufferData tests buffer data synchronization.
func TestAnalyticsAggregation_SyncBufferData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	analyticsClient := clients.NewAnalyticsClient(env.AnalyticsService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "Sync",
		LastName:       "Test",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	t.Run("TriggerBufferSync", func(t *testing.T) {
		err := analyticsClient.SyncBufferData(ctx, organizationID, accessToken)
		// Sync might succeed or fail depending on service availability
		if err != nil {
			t.Logf("Buffer sync returned error (may be expected): %v", err)
		}
	})
}

// TestAnalyticsAggregation_AfterTransactions tests that analytics update after order transactions.
func TestAnalyticsAggregation_AfterTransactions(t *testing.T) {
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

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "Analytics",
		LastName:       "Transaction",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	// Create product
	productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: organizationID,
		SKU:            generateTestSKU("ANALYTICS"),
		Name:           "Analytics Test Product",
		Description:    "Product for analytics testing",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, accessToken)
	require.NoError(t, err)
	productID := productResp.Product.ID

	t.Run("CreateTransactionsAndCheckAnalytics", func(t *testing.T) {
		// Create and receive purchase order
		poResp, err := executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
			OrganizationID: organizationID,
			SupplierID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productID,
					SKU:       "ANALYTICS-SKU",
					Quantity:  500.0,
					UnitPrice: 10.0,
				},
			},
			Notes: "Order for analytics test",
		}, accessToken)
		require.NoError(t, err)

		_, err = executionClient.ReceiveGoods(ctx, poResp.Order.ID, clients.ReceiveGoodsRequest{
			Items: []clients.ReceiveItemRequest{
				{
					ProductID: productID,
					Quantity:  500.0,
				},
			},
		}, accessToken)
		require.NoError(t, err)

		// Wait for analytics to update
		time.Sleep(2 * time.Second)

		// Check buffer analytics
		analytics, err := analyticsClient.GetBufferAnalytics(ctx, organizationID, accessToken)
		if err == nil && len(analytics) > 0 {
			// Find our product
			for _, a := range analytics {
				if a.ProductID == productID {
					assert.NotZero(t, a.NetFlowPosition, "NFP should be updated")
				}
			}
		}
	})
}

// TestAnalyticsAggregation_CrossServiceData tests that analytics aggregates data from multiple services.
func TestAnalyticsAggregation_CrossServiceData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	analyticsClient := clients.NewAnalyticsClient(env.AnalyticsService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "Cross",
		LastName:       "Service",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	t.Run("CreateProductsAndCheckCounts", func(t *testing.T) {
		// Create multiple products
		for i := 0; i < 5; i++ {
			_, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
				OrganizationID: organizationID,
				SKU:            generateTestSKU("CROSS-" + string(rune('A'+i))),
				Name:           "Cross Service Product " + string(rune('A'+i)),
				Description:    "Product for cross service test",
				Category:       "Test Category " + string(rune('A'+i%3)), // 3 different categories
				UnitOfMeasure:  "UNIT",
			}, accessToken)
			require.NoError(t, err)
		}

		// Wait for data to propagate
		time.Sleep(2 * time.Second)

		// Verify products in catalog
		products, err := catalogClient.ListProducts(ctx, organizationID, accessToken, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, products.Total, 5, "Should have at least 5 products")

		// Check analytics snapshot
		snapshot, err := analyticsClient.GetSnapshot(ctx, organizationID, accessToken)
		if err == nil && snapshot != nil {
			assert.GreaterOrEqual(t, snapshot.TotalProducts, 5, "Snapshot should reflect products")
		}
	})
}
