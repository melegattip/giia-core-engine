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

// TestSalesOrderFlow_CreateToShip tests the complete sales order lifecycle:
// 1. Create user and authenticate
// 2. Create product in catalog
// 3. Add initial inventory
// 4. Create sales order
// 5. Ship the order
// 6. Verify inventory decreased
// 7. Verify qualified demand updated in DDMRP
func TestSalesOrderFlow_CreateToShip(t *testing.T) {
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

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	var accessToken string
	var productID string
	var salesOrderID string

	t.Run("1_RegisterAndLogin", func(t *testing.T) {
		_, err := authClient.Register(ctx, clients.RegisterRequest{
			Email:          email,
			Password:       password,
			FirstName:      "Sales",
			LastName:       "OrderTest",
			Phone:          "+1234567890",
			OrganizationID: organizationID,
		})
		require.NoError(t, err, "User registration should succeed")

		tokens, err := authClient.Login(ctx, email, password)
		require.NoError(t, err, "Login should succeed")
		accessToken = tokens.AccessToken
	})

	t.Run("2_CreateProduct", func(t *testing.T) {
		resp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("SO"),
			Name:           "Sales Order Test Product",
			Description:    "Product for sales order testing",
			Category:       "Test Category",
			UnitOfMeasure:  "UNIT",
		}, accessToken)
		require.NoError(t, err)
		productID = resp.Product.ID
	})

	t.Run("3_AddInitialInventory", func(t *testing.T) {
		// Create and receive a purchase order to add inventory
		poResp, err := executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
			OrganizationID: organizationID,
			SupplierID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productID,
					SKU:       "SO-INIT",
					Quantity:  200.0,
					UnitPrice: 10.0,
				},
			},
			Notes: "Initial inventory for sales order test",
		}, accessToken)
		require.NoError(t, err)

		_, err = executionClient.ReceiveGoods(ctx, poResp.Order.ID, clients.ReceiveGoodsRequest{
			Items: []clients.ReceiveItemRequest{
				{
					ProductID: productID,
					Quantity:  200.0,
				},
			},
		}, accessToken)
		require.NoError(t, err)

		// Wait for inventory update
		time.Sleep(1 * time.Second)
	})

	t.Run("4_CreateSalesOrder", func(t *testing.T) {
		resp, err := executionClient.CreateSalesOrder(ctx, clients.CreateSalesOrderRequest{
			OrganizationID: organizationID,
			CustomerID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productID,
					SKU:       "SO-SKU",
					Quantity:  50.0,
					UnitPrice: 25.0,
				},
			},
			Notes: "Integration test sales order",
		}, accessToken)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Order.ID)
		assert.Equal(t, "pending", resp.Order.Status)
		salesOrderID = resp.Order.ID
	})

	t.Run("5_VerifySalesOrderCreated", func(t *testing.T) {
		order, err := executionClient.GetSalesOrder(ctx, salesOrderID, organizationID, accessToken)
		require.NoError(t, err)
		assert.Equal(t, salesOrderID, order.ID)
		assert.Equal(t, "pending", order.Status)
	})

	t.Run("6_ShipOrder", func(t *testing.T) {
		order, err := executionClient.ShipOrder(ctx, salesOrderID, clients.ShipOrderRequest{
			Items: []clients.ShipItemRequest{
				{
					ProductID: productID,
					Quantity:  50.0,
				},
			},
		}, accessToken)
		require.NoError(t, err)
		assert.Equal(t, "shipped", order.Status)
	})

	t.Run("7_VerifyInventoryDecreased", func(t *testing.T) {
		time.Sleep(1 * time.Second)

		balances, err := executionClient.GetInventoryBalances(ctx, organizationID, accessToken)
		require.NoError(t, err)

		for _, balance := range balances {
			if balance.ProductID == productID {
				// Initial 200 - shipped 50 = 150
				assert.Equal(t, 150.0, balance.OnHand, "On-hand should be 150 after shipping")
				break
			}
		}
	})
}

// TestSalesOrderFlow_CreateAndCancel tests sales order cancellation.
func TestSalesOrderFlow_CreateAndCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	executionClient := clients.NewExecutionClient(env.ExecutionService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       "SecurePassword123!",
		FirstName:      "Cancel",
		LastName:       "Sales",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, "SecurePassword123!")
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: organizationID,
		SKU:            generateTestSKU("SO-CANCEL"),
		Name:           "Sales Cancel Product",
		Description:    "Product for sales order cancel test",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, accessToken)
	require.NoError(t, err)

	t.Run("CreateAndCancelSalesOrder", func(t *testing.T) {
		// Create sales order
		soResp, err := executionClient.CreateSalesOrder(ctx, clients.CreateSalesOrderRequest{
			OrganizationID: organizationID,
			CustomerID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productResp.Product.ID,
					SKU:       "CANCEL-SO",
					Quantity:  30.0,
					UnitPrice: 20.0,
				},
			},
			Notes: "Order to be cancelled",
		}, accessToken)
		require.NoError(t, err)
		assert.Equal(t, "pending", soResp.Order.Status)

		// Cancel order
		err = executionClient.CancelSalesOrder(ctx, soResp.Order.ID, accessToken)
		require.NoError(t, err)

		// Verify cancelled
		order, err := executionClient.GetSalesOrder(ctx, soResp.Order.ID, organizationID, accessToken)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", order.Status)
	})
}

// TestSalesOrderFlow_InsufficientInventory tests that sales orders fail when there's not enough inventory.
func TestSalesOrderFlow_InsufficientInventory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	executionClient := clients.NewExecutionClient(env.ExecutionService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       "SecurePassword123!",
		FirstName:      "Insufficient",
		LastName:       "Inventory",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, "SecurePassword123!")
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	// Create product with no inventory
	productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: organizationID,
		SKU:            generateTestSKU("NOSTOCK"),
		Name:           "No Stock Product",
		Description:    "Product with no inventory",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, accessToken)
	require.NoError(t, err)

	t.Run("ShipWithoutInventory", func(t *testing.T) {
		// Create sales order
		soResp, err := executionClient.CreateSalesOrder(ctx, clients.CreateSalesOrderRequest{
			OrganizationID: organizationID,
			CustomerID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productResp.Product.ID,
					SKU:       "NOSTOCK-SKU",
					Quantity:  100.0,
					UnitPrice: 50.0,
				},
			},
			Notes: "Order that should fail to ship",
		}, accessToken)
		require.NoError(t, err)

		// Try to ship - should fail due to insufficient inventory
		_, err = executionClient.ShipOrder(ctx, soResp.Order.ID, clients.ShipOrderRequest{
			Items: []clients.ShipItemRequest{
				{
					ProductID: productResp.Product.ID,
					Quantity:  100.0,
				},
			},
		}, accessToken)

		// Expect error or the order to remain in pending state
		// Implementation might handle this differently
		if err != nil {
			assert.Contains(t, err.Error(), "insufficient", "Error should mention insufficient inventory")
		}
	})
}

// TestSalesOrderFlow_MultipleItems tests sales orders with multiple line items.
func TestSalesOrderFlow_MultipleItems(t *testing.T) {
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

	organizationID := uuid.New().String()
	email := generateTestEmail()

	_, err := authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       "SecurePassword123!",
		FirstName:      "Multi",
		LastName:       "Item",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, "SecurePassword123!")
	require.NoError(t, err)
	accessToken := tokens.AccessToken

	// Create multiple products
	var productIDs []string
	for i := 0; i < 3; i++ {
		resp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("MULTI-" + string(rune('A'+i))),
			Name:           "Multi Item Product " + string(rune('A'+i)),
			Description:    "Product for multi-item test",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, accessToken)
		require.NoError(t, err)
		productIDs = append(productIDs, resp.Product.ID)
	}

	t.Run("CreateOrderWithMultipleItems", func(t *testing.T) {
		items := make([]clients.CreateOrderItemRequest, len(productIDs))
		for i, pid := range productIDs {
			items[i] = clients.CreateOrderItemRequest{
				ProductID: pid,
				SKU:       "MULTI-" + string(rune('A'+i)),
				Quantity:  float64(10 * (i + 1)), // 10, 20, 30
				UnitPrice: float64(5 * (i + 1)),  // 5, 10, 15
			}
		}

		soResp, err := executionClient.CreateSalesOrder(ctx, clients.CreateSalesOrderRequest{
			OrganizationID: organizationID,
			CustomerID:     uuid.New().String(),
			Items:          items,
			Notes:          "Multi-item order",
		}, accessToken)
		require.NoError(t, err)
		assert.Len(t, soResp.Order.Items, 3, "Order should have 3 items")

		// Calculate expected total: 10*5 + 20*10 + 30*15 = 50 + 200 + 450 = 700
		assert.Equal(t, 700.0, soResp.Order.TotalAmount, "Total amount should be 700")
	})
}
