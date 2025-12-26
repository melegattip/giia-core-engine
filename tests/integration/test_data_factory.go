// Package integration provides integration testing utilities for the GIIA platform.
package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TestDataFactory creates test data for integration tests.
type TestDataFactory struct {
	env             *TestEnvironment
	authClient      *AuthClient
	catalogClient   *CatalogClient
	executionClient *ExecutionClient
	ddmrpClient     *DDMRPClient
	analyticsClient *AnalyticsClient
	aiHubClient     *AIHubClient
}

// NewTestDataFactory creates a new test data factory.
func NewTestDataFactory(env *TestEnvironment) *TestDataFactory {
	return &TestDataFactory{
		env:             env,
		authClient:      NewAuthClient(env.AuthService.HTTPURL),
		catalogClient:   NewCatalogClient(env.CatalogService.HTTPURL),
		executionClient: NewExecutionClient(env.ExecutionService.HTTPURL),
		ddmrpClient:     NewDDMRPClient(env.DDMRPService.HTTPURL, env.DDMRPService.GRPCURL),
		analyticsClient: NewAnalyticsClient(env.AnalyticsService.HTTPURL),
		aiHubClient:     NewAIHubClient(env.AIHubService.HTTPURL),
	}
}

// TestSetup contains all created test entities.
type TestSetup struct {
	OrganizationID string
	Users          []TestUser
	Products       []TestProduct
	BufferProfiles []TestBufferProfile
	Buffers        []TestBuffer
	PurchaseOrders []TestPurchaseOrder
	SalesOrders    []TestSalesOrder
	AccessToken    string
	RefreshToken   string
}

// TestUser represents a test user.
type TestUser struct {
	ID             string
	Email          string
	Password       string
	FirstName      string
	LastName       string
	OrganizationID string
	Roles          []string
}

// TestProduct represents a test product.
type TestProduct struct {
	ID              string
	SKU             string
	Name            string
	Description     string
	Category        string
	UnitOfMeasure   string
	Status          string
	OrganizationID  string
	BufferProfileID string
}

// TestBufferProfile represents a test buffer profile.
type TestBufferProfile struct {
	ID                string
	Name              string
	LeadTimeFactor    float64
	VariabilityFactor float64
	OrganizationID    string
}

// TestBuffer represents a test buffer.
type TestBuffer struct {
	ID              string
	ProductID       string
	OrganizationID  string
	BufferProfileID string
	TopOfGreen      float64
	TopOfYellow     float64
	TopOfRed        float64
	OnHand          float64
	OnOrder         float64
	NetFlowPosition float64
	Zone            string
}

// TestPurchaseOrder represents a test purchase order.
type TestPurchaseOrder struct {
	ID             string
	OrderNumber    string
	SupplierID     string
	OrganizationID string
	Status         string
	Items          []TestOrderItem
	TotalAmount    float64
}

// TestSalesOrder represents a test sales order.
type TestSalesOrder struct {
	ID             string
	OrderNumber    string
	CustomerID     string
	OrganizationID string
	Status         string
	Items          []TestOrderItem
	TotalAmount    float64
}

// TestOrderItem represents a test order item.
type TestOrderItem struct {
	ID        string
	ProductID string
	SKU       string
	Quantity  float64
	UnitPrice float64
}

// CreateCompleteSetup creates a complete test setup with all entities.
func (f *TestDataFactory) CreateCompleteSetup(ctx context.Context) (*TestSetup, error) {
	setup := &TestSetup{
		OrganizationID: uuid.New().String(),
	}

	// Create user and authenticate
	user, err := f.CreateUser(ctx, setup.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	setup.Users = append(setup.Users, *user)

	// Login to get token
	tokens, err := f.authClient.Login(ctx, user.Email, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	setup.AccessToken = tokens.AccessToken
	setup.RefreshToken = tokens.RefreshToken

	// Create products
	for i := 0; i < 3; i++ {
		product, err := f.CreateProduct(ctx, setup.OrganizationID, setup.AccessToken, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create product %d: %w", i, err)
		}
		setup.Products = append(setup.Products, *product)
	}

	return setup, nil
}

// CreateUser creates a test user.
func (f *TestDataFactory) CreateUser(ctx context.Context, organizationID string) (*TestUser, error) {
	user := &TestUser{
		Email:          fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Password:       "SecurePassword123!",
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: organizationID,
		Roles:          []string{"admin"},
	}

	resp, err := f.authClient.Register(ctx, RegisterRequest{
		Email:          user.Email,
		Password:       user.Password,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Phone:          "+1234567890",
		OrganizationID: user.OrganizationID,
	})
	if err != nil {
		return nil, err
	}

	user.ID = resp.UserID
	return user, nil
}

// CreateProduct creates a test product.
func (f *TestDataFactory) CreateProduct(ctx context.Context, organizationID, accessToken string, index int) (*TestProduct, error) {
	product := &TestProduct{
		SKU:            fmt.Sprintf("SKU-%d-%d", time.Now().Unix(), index),
		Name:           fmt.Sprintf("Test Product %d", index),
		Description:    fmt.Sprintf("Test product description %d", index),
		Category:       "Test Category",
		UnitOfMeasure:  "UNIT",
		Status:         "active",
		OrganizationID: organizationID,
	}

	resp, err := f.catalogClient.CreateProduct(ctx, CreateProductRequest{
		OrganizationID: product.OrganizationID,
		SKU:            product.SKU,
		Name:           product.Name,
		Description:    product.Description,
		Category:       product.Category,
		UnitOfMeasure:  product.UnitOfMeasure,
	}, accessToken)
	if err != nil {
		return nil, err
	}

	product.ID = resp.Product.ID
	return product, nil
}

// CreatePurchaseOrder creates a test purchase order.
func (f *TestDataFactory) CreatePurchaseOrder(ctx context.Context, organizationID, accessToken string, products []TestProduct) (*TestPurchaseOrder, error) {
	items := make([]CreateOrderItemRequest, len(products))
	for i, p := range products {
		items[i] = CreateOrderItemRequest{
			ProductID: p.ID,
			SKU:       p.SKU,
			Quantity:  100.0,
			UnitPrice: 10.0,
		}
	}

	resp, err := f.executionClient.CreatePurchaseOrder(ctx, CreatePurchaseOrderRequest{
		OrganizationID: organizationID,
		SupplierID:     uuid.New().String(),
		Items:          items,
		Notes:          "Test purchase order",
	}, accessToken)
	if err != nil {
		return nil, err
	}

	po := &TestPurchaseOrder{
		ID:             resp.Order.ID,
		OrderNumber:    resp.Order.OrderNumber,
		SupplierID:     resp.Order.SupplierID,
		OrganizationID: organizationID,
		Status:         resp.Order.Status,
		TotalAmount:    resp.Order.TotalAmount,
	}

	for _, item := range resp.Order.Items {
		po.Items = append(po.Items, TestOrderItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return po, nil
}

// CreateSalesOrder creates a test sales order.
func (f *TestDataFactory) CreateSalesOrder(ctx context.Context, organizationID, accessToken string, products []TestProduct) (*TestSalesOrder, error) {
	items := make([]CreateOrderItemRequest, len(products))
	for i, p := range products {
		items[i] = CreateOrderItemRequest{
			ProductID: p.ID,
			SKU:       p.SKU,
			Quantity:  10.0,
			UnitPrice: 15.0,
		}
	}

	resp, err := f.executionClient.CreateSalesOrder(ctx, CreateSalesOrderRequest{
		OrganizationID: organizationID,
		CustomerID:     uuid.New().String(),
		Items:          items,
		Notes:          "Test sales order",
	}, accessToken)
	if err != nil {
		return nil, err
	}

	so := &TestSalesOrder{
		ID:             resp.Order.ID,
		OrderNumber:    resp.Order.OrderNumber,
		CustomerID:     resp.Order.CustomerID,
		OrganizationID: organizationID,
		Status:         resp.Order.Status,
		TotalAmount:    resp.Order.TotalAmount,
	}

	for _, item := range resp.Order.Items {
		so.Items = append(so.Items, TestOrderItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return so, nil
}

// CreateMultipleTenants creates test data for multiple organizations.
func (f *TestDataFactory) CreateMultipleTenants(ctx context.Context, count int) ([]*TestSetup, error) {
	setups := make([]*TestSetup, count)

	for i := 0; i < count; i++ {
		setup, err := f.CreateCompleteSetup(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create tenant %d: %w", i, err)
		}
		setups[i] = setup
	}

	return setups, nil
}

// Cleanup removes all test data created by this factory.
func (f *TestDataFactory) Cleanup(ctx context.Context, setup *TestSetup) error {
	if setup == nil {
		return nil
	}
	return f.env.CleanupTestData(setup.OrganizationID)
}
