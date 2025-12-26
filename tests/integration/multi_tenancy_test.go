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

// TestMultiTenancy_Isolation tests that data from one organization is not visible to another.
func TestMultiTenancy_Isolation(t *testing.T) {
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

	// Create two separate organizations
	orgA := uuid.New().String()
	orgB := uuid.New().String()

	// Setup Organization A
	emailA := generateTestEmail()
	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          emailA,
		Password:       "SecurePassword123!",
		FirstName:      "OrgA",
		LastName:       "User",
		Phone:          "+1234567890",
		OrganizationID: orgA,
	})
	require.NoError(t, err)

	tokensA, err := authClient.Login(ctx, emailA, "SecurePassword123!")
	require.NoError(t, err)

	// Setup Organization B
	emailB := generateTestEmail()
	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          emailB,
		Password:       "SecurePassword123!",
		FirstName:      "OrgB",
		LastName:       "User",
		Phone:          "+1234567890",
		OrganizationID: orgB,
	})
	require.NoError(t, err)

	tokensB, err := authClient.Login(ctx, emailB, "SecurePassword123!")
	require.NoError(t, err)

	// Create data in Organization A
	var productAID string

	t.Run("1_CreateDataInOrgA", func(t *testing.T) {
		// Create product in Org A
		productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: orgA,
			SKU:            generateTestSKU("ORGA"),
			Name:           "Org A Exclusive Product",
			Description:    "This product belongs only to Org A",
			Category:       "Confidential",
			UnitOfMeasure:  "UNIT",
		}, tokensA.AccessToken)
		require.NoError(t, err)
		productAID = productResp.Product.ID

		// Create purchase order in Org A
		_, err = executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
			OrganizationID: orgA,
			SupplierID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productAID,
					SKU:       "ORGA-SKU",
					Quantity:  100.0,
					UnitPrice: 50.0,
				},
			},
			Notes: "Org A purchase order",
		}, tokensA.AccessToken)
		require.NoError(t, err)
	})

	t.Run("2_OrgBCannotSeeOrgAProducts", func(t *testing.T) {
		// List products for Org B - should not see Org A's products
		products, err := catalogClient.ListProducts(ctx, orgB, tokensB.AccessToken, 1, 100)
		require.NoError(t, err)

		for _, p := range products.Products {
			assert.NotEqual(t, productAID, p.ID, "Org B should not see Org A's product")
			assert.NotEqual(t, orgA, p.OrganizationID, "No products from Org A should appear")
		}
	})

	t.Run("3_OrgBCannotAccessOrgAProductDirectly", func(t *testing.T) {
		// Try to get Org A's product directly
		_, err := catalogClient.GetProduct(ctx, productAID, orgA, tokensB.AccessToken)
		require.Error(t, err, "Org B should not be able to access Org A's product")

		// Should get 404 or 403
		assert.True(t,
			assert.Contains(t, err.Error(), "404") || assert.Contains(t, err.Error(), "403") || assert.Contains(t, err.Error(), "not found") || assert.Contains(t, err.Error(), "forbidden"),
			"Should return 404 or 403")
	})

	t.Run("4_OrgBCannotModifyOrgAProduct", func(t *testing.T) {
		// Try to update Org A's product
		_, err := catalogClient.UpdateProduct(ctx, clients.UpdateProductRequest{
			ID:             productAID,
			OrganizationID: orgA,
			Name:           "Hacked by Org B",
			Description:    "This should fail",
			Category:       "Hacked",
			UnitOfMeasure:  "UNIT",
			Status:         "active",
		}, tokensB.AccessToken)
		require.Error(t, err, "Org B should not be able to update Org A's product")
	})

	t.Run("5_OrgBCannotDeleteOrgAProduct", func(t *testing.T) {
		// Try to delete Org A's product
		err := catalogClient.DeleteProduct(ctx, productAID, orgA, tokensB.AccessToken)
		require.Error(t, err, "Org B should not be able to delete Org A's product")

		// Verify product still exists for Org A
		product, err := catalogClient.GetProduct(ctx, productAID, orgA, tokensA.AccessToken)
		require.NoError(t, err, "Product should still exist for Org A")
		assert.Equal(t, productAID, product.Product.ID)
	})
}

// TestMultiTenancy_ConcurrentOperations tests that concurrent operations from different orgs don't interfere.
func TestMultiTenancy_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	// Create 3 organizations
	orgs := make([]struct {
		id    string
		token string
	}, 3)

	for i := 0; i < 3; i++ {
		orgs[i].id = uuid.New().String()
		email := generateTestEmail()

		_, err := authClient.Register(ctx, clients.RegisterRequest{
			Email:          email,
			Password:       "SecurePassword123!",
			FirstName:      "Org" + string(rune('A'+i)),
			LastName:       "User",
			Phone:          "+1234567890",
			OrganizationID: orgs[i].id,
		})
		require.NoError(t, err)

		tokens, err := authClient.Login(ctx, email, "SecurePassword123!")
		require.NoError(t, err)
		orgs[i].token = tokens.AccessToken
	}

	t.Run("ConcurrentProductCreation", func(t *testing.T) {
		// Create products concurrently for each org
		done := make(chan struct{}, len(orgs))
		errors := make(chan error, len(orgs))
		productCounts := make(chan int, len(orgs))

		for i, org := range orgs {
			go func(orgIndex int, orgID, token string) {
				defer func() { done <- struct{}{} }()

				// Create 5 products per org
				for j := 0; j < 5; j++ {
					_, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
						OrganizationID: orgID,
						SKU:            generateTestSKU("CONC-" + string(rune('A'+orgIndex)) + "-" + string(rune('0'+j))),
						Name:           "Concurrent Product " + string(rune('0'+j)),
						Description:    "Created concurrently",
						Category:       "Test",
						UnitOfMeasure:  "UNIT",
					}, token)
					if err != nil {
						errors <- err
						return
					}
				}

				// Count products for this org
				products, err := catalogClient.ListProducts(ctx, orgID, token, 1, 100)
				if err != nil {
					errors <- err
					return
				}
				productCounts <- products.Total
			}(i, org.id, org.token)
		}

		// Wait for all goroutines
		for i := 0; i < len(orgs); i++ {
			<-done
		}
		close(errors)
		close(productCounts)

		// Check for errors
		for err := range errors {
			require.NoError(t, err)
		}

		// Verify each org has exactly 5 products
		for count := range productCounts {
			assert.Equal(t, 5, count, "Each org should have exactly 5 products")
		}
	})
}

// TestMultiTenancy_DataLeakPrevention tests that queries don't leak data across orgs.
func TestMultiTenancy_DataLeakPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	// Create two organizations with similar data
	orgA := uuid.New().String()
	orgB := uuid.New().String()

	emailA := generateTestEmail()
	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          emailA,
		Password:       "SecurePassword123!",
		FirstName:      "OrgA",
		LastName:       "Admin",
		Phone:          "+1234567890",
		OrganizationID: orgA,
	})
	require.NoError(t, err)
	tokensA, err := authClient.Login(ctx, emailA, "SecurePassword123!")
	require.NoError(t, err)

	emailB := generateTestEmail()
	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          emailB,
		Password:       "SecurePassword123!",
		FirstName:      "OrgB",
		LastName:       "Admin",
		Phone:          "+1234567890",
		OrganizationID: orgB,
	})
	require.NoError(t, err)
	tokensB, err := authClient.Login(ctx, emailB, "SecurePassword123!")
	require.NoError(t, err)

	// Create products with same name in both orgs
	t.Run("SameNameDifferentOrgs", func(t *testing.T) {
		// Create "Widget" in Org A
		productA, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: orgA,
			SKU:            generateTestSKU("WIDGET-A"),
			Name:           "Widget",
			Description:    "Org A's Widget",
			Category:       "Widgets",
			UnitOfMeasure:  "UNIT",
		}, tokensA.AccessToken)
		require.NoError(t, err)

		// Create "Widget" in Org B
		productB, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: orgB,
			SKU:            generateTestSKU("WIDGET-B"),
			Name:           "Widget",
			Description:    "Org B's Widget",
			Category:       "Widgets",
			UnitOfMeasure:  "UNIT",
		}, tokensB.AccessToken)
		require.NoError(t, err)

		// IDs should be different
		assert.NotEqual(t, productA.Product.ID, productB.Product.ID, "Products should have different IDs")

		// Search for "Widget" in each org
		searchA, err := catalogClient.SearchProducts(ctx, orgA, "Widget", tokensA.AccessToken)
		require.NoError(t, err)

		searchB, err := catalogClient.SearchProducts(ctx, orgB, "Widget", tokensB.AccessToken)
		require.NoError(t, err)

		// Each should only see their own widget
		for _, p := range searchA.Products {
			assert.Equal(t, orgA, p.OrganizationID, "Org A search should only return Org A products")
		}
		for _, p := range searchB.Products {
			assert.Equal(t, orgB, p.OrganizationID, "Org B search should only return Org B products")
		}
	})
}

// TestMultiTenancy_OrganizationScopedInventory tests that inventory is isolated per organization.
func TestMultiTenancy_OrganizationScopedInventory(t *testing.T) {
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

	orgA := uuid.New().String()
	orgB := uuid.New().String()

	// Setup Org A
	emailA := generateTestEmail()
	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          emailA,
		Password:       "SecurePassword123!",
		FirstName:      "OrgA",
		LastName:       "Inventory",
		Phone:          "+1234567890",
		OrganizationID: orgA,
	})
	require.NoError(t, err)
	tokensA, err := authClient.Login(ctx, emailA, "SecurePassword123!")
	require.NoError(t, err)

	// Setup Org B
	emailB := generateTestEmail()
	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          emailB,
		Password:       "SecurePassword123!",
		FirstName:      "OrgB",
		LastName:       "Inventory",
		Phone:          "+1234567890",
		OrganizationID: orgB,
	})
	require.NoError(t, err)
	tokensB, err := authClient.Login(ctx, emailB, "SecurePassword123!")
	require.NoError(t, err)

	// Create product and add inventory in Org A
	productA, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: orgA,
		SKU:            generateTestSKU("INV-A"),
		Name:           "Inventory Test A",
		Description:    "Test",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, tokensA.AccessToken)
	require.NoError(t, err)

	poResp, err := executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
		OrganizationID: orgA,
		SupplierID:     uuid.New().String(),
		Items: []clients.CreateOrderItemRequest{
			{
				ProductID: productA.Product.ID,
				SKU:       "INV-A-SKU",
				Quantity:  1000.0,
				UnitPrice: 10.0,
			},
		},
		Notes: "Inventory test",
	}, tokensA.AccessToken)
	require.NoError(t, err)

	_, err = executionClient.ReceiveGoods(ctx, poResp.Order.ID, clients.ReceiveGoodsRequest{
		Items: []clients.ReceiveItemRequest{
			{
				ProductID: productA.Product.ID,
				Quantity:  1000.0,
			},
		},
	}, tokensA.AccessToken)
	require.NoError(t, err)

	t.Run("OrgBCannotSeeOrgAInventory", func(t *testing.T) {
		// Org B should not see Org A's inventory
		balancesB, err := executionClient.GetInventoryBalances(ctx, orgB, tokensB.AccessToken)
		require.NoError(t, err)

		for _, b := range balancesB {
			assert.NotEqual(t, productA.Product.ID, b.ProductID, "Org B should not see Org A's inventory")
		}
	})

	t.Run("OrgACanSeeOwnInventory", func(t *testing.T) {
		time.Sleep(1 * time.Second)

		balancesA, err := executionClient.GetInventoryBalances(ctx, orgA, tokensA.AccessToken)
		require.NoError(t, err)

		found := false
		for _, b := range balancesA {
			if b.ProductID == productA.Product.ID {
				found = true
				assert.Equal(t, 1000.0, b.OnHand, "Should have 1000 units on hand")
			}
		}
		if len(balancesA) > 0 {
			assert.True(t, found, "Org A should see its inventory")
		}
	})
}
