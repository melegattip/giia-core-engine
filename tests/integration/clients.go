// Package integration provides re-exports for client types.
package integration

import (
	"github.com/melegattip/giia-core-engine/tests/integration/clients"
)

// Client type aliases for convenience
type (
	AuthClient      = clients.AuthClient
	CatalogClient   = clients.CatalogClient
	ExecutionClient = clients.ExecutionClient
	DDMRPClient     = clients.DDMRPClient
	AnalyticsClient = clients.AnalyticsClient
	AIHubClient     = clients.AIHubClient
)

// Request/Response type aliases
type (
	RegisterRequest             = clients.RegisterRequest
	RegisterResponse            = clients.RegisterResponse
	CreateProductRequest        = clients.CreateProductRequest
	CreateProductResponse       = clients.CreateProductResponse
	CreatePurchaseOrderRequest  = clients.CreatePurchaseOrderRequest
	CreatePurchaseOrderResponse = clients.CreatePurchaseOrderResponse
	CreateSalesOrderRequest     = clients.CreateSalesOrderRequest
	CreateSalesOrderResponse    = clients.CreateSalesOrderResponse
	CreateOrderItemRequest      = clients.CreateOrderItemRequest
	ReceiveGoodsRequest         = clients.ReceiveGoodsRequest
	ReceiveItemRequest          = clients.ReceiveItemRequest
	ShipOrderRequest            = clients.ShipOrderRequest
	ShipItemRequest             = clients.ShipItemRequest
)

// Client constructor aliases
var (
	NewAuthClient      = clients.NewAuthClient
	NewCatalogClient   = clients.NewCatalogClient
	NewExecutionClient = clients.NewExecutionClient
	NewDDMRPClient     = clients.NewDDMRPClient
	NewAnalyticsClient = clients.NewAnalyticsClient
	NewAIHubClient     = clients.NewAIHubClient
)
