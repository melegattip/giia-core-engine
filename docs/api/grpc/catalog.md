# Catalog Service gRPC API

**Package:** `catalog.v1`  
**Port:** 9082  
**Proto File:** `services/catalog-service/api/proto/catalog/v1/catalog.proto`

---

## Service Definition

```protobuf
service CatalogService {
  // Product operations
  rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
  rpc UpdateProduct(UpdateProductRequest) returns (UpdateProductResponse);
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse);
  rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse);

  // Supplier operations
  rpc CreateSupplier(CreateSupplierRequest) returns (CreateSupplierResponse);
  rpc UpdateSupplier(UpdateSupplierRequest) returns (UpdateSupplierResponse);
  rpc GetSupplier(GetSupplierRequest) returns (GetSupplierResponse);
  rpc ListSuppliers(ListSuppliersRequest) returns (ListSuppliersResponse);
  rpc DeleteSupplier(DeleteSupplierRequest) returns (DeleteSupplierResponse);

  // Buffer Profile operations
  rpc CreateBufferProfile(CreateBufferProfileRequest) returns (CreateBufferProfileResponse);
  rpc UpdateBufferProfile(UpdateBufferProfileRequest) returns (UpdateBufferProfileResponse);
  rpc GetBufferProfile(GetBufferProfileRequest) returns (GetBufferProfileResponse);
  rpc ListBufferProfiles(ListBufferProfilesRequest) returns (ListBufferProfilesResponse);
  rpc DeleteBufferProfile(DeleteBufferProfileRequest) returns (DeleteBufferProfileResponse);

  // Product-Supplier associations
  rpc AssociateSupplier(AssociateSupplierRequest) returns (AssociateSupplierResponse);
  rpc GetProductSuppliers(GetProductSuppliersRequest) returns (GetProductSuppliersResponse);
  rpc RemoveSupplierAssociation(RemoveSupplierAssociationRequest) returns (RemoveSupplierAssociationResponse);
}
```

---

## Message Types

### Product

```protobuf
message Product {
  string id = 1;
  string organization_id = 2;
  string sku = 3;
  string name = 4;
  string description = 5;
  string category = 6;
  string unit_of_measure = 7;
  string buffer_profile_id = 8;
  string status = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}
```

### Supplier

```protobuf
message Supplier {
  string id = 1;
  string organization_id = 2;
  string code = 3;
  string name = 4;
  string contact_name = 5;
  string contact_email = 6;
  string contact_phone = 7;
  string status = 8;
  string address = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}
```

### BufferProfile

```protobuf
message BufferProfile {
  string id = 1;
  string organization_id = 2;
  string name = 3;
  string description = 4;
  string adu_method = 5;        // average, exponential, weighted
  double lead_time_factor = 6;
  double variability_factor = 7;
  string status = 8;
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp updated_at = 10;
}
```

### ProductSupplier

```protobuf
message ProductSupplier {
  string product_id = 1;
  string supplier_id = 2;
  int32 lead_time_days = 3;
  bool is_primary = 4;
}
```

---

## Key Methods

### GetProduct

Retrieves a product by ID for inter-service communication.

**Request:**
```protobuf
message GetProductRequest {
  string id = 1;
  string organization_id = 2;
}
```

**Response:**
```protobuf
message GetProductResponse {
  Product product = 1;
}
```

**Example (Go):**
```go
resp, err := catalogClient.GetProduct(ctx, &catalogv1.GetProductRequest{
    Id:             productID,
    OrganizationId: orgID,
})
if err != nil {
    return err
}
product := resp.Product
```

---

### ListProducts

Lists products with pagination and filtering.

**Request:**
```protobuf
message ListProductsRequest {
  string organization_id = 1;
  int32 page = 2;
  int32 page_size = 3;
  string status = 4;
  string category = 5;
}
```

**Response:**
```protobuf
message ListProductsResponse {
  repeated Product products = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}
```

---

### GetProductSuppliers

Retrieves all suppliers associated with a product.

**Request:**
```protobuf
message GetProductSuppliersRequest {
  string product_id = 1;
  string organization_id = 2;
}
```

**Response:**
```protobuf
message GetProductSuppliersResponse {
  repeated ProductSupplier product_suppliers = 1;
}
```

**Example (Go):**
```go
resp, err := catalogClient.GetProductSuppliers(ctx, &catalogv1.GetProductSuppliersRequest{
    ProductId:      productID,
    OrganizationId: orgID,
})
// Use resp.ProductSuppliers to get lead times
for _, ps := range resp.ProductSuppliers {
    if ps.IsPrimary {
        leadTimeDays := ps.LeadTimeDays
        // Use for buffer calculations
    }
}
```

---

### GetBufferProfile

Retrieves buffer profile for DDMRP calculations.

**Request:**
```protobuf
message GetBufferProfileRequest {
  string id = 1;
  string organization_id = 2;
}
```

**Response:**
```protobuf
message GetBufferProfileResponse {
  BufferProfile buffer_profile = 1;
}
```

---

## Usage from DDMRP Engine

The Catalog Service is called by DDMRP Engine to get product and profile data:

```go
func (s *DDMRPService) CalculateBuffer(ctx context.Context, productID string) (*Buffer, error) {
    // Get product details
    prodResp, err := s.catalogClient.GetProduct(ctx, &catalogv1.GetProductRequest{
        Id:             productID,
        OrganizationId: s.orgID,
    })
    if err != nil {
        return nil, fmt.Errorf("get product: %w", err)
    }
    
    // Get buffer profile
    profileResp, err := s.catalogClient.GetBufferProfile(ctx, &catalogv1.GetBufferProfileRequest{
        Id:             prodResp.Product.BufferProfileId,
        OrganizationId: s.orgID,
    })
    if err != nil {
        return nil, fmt.Errorf("get buffer profile: %w", err)
    }
    
    // Get supplier lead time
    suppResp, err := s.catalogClient.GetProductSuppliers(ctx, &catalogv1.GetProductSuppliersRequest{
        ProductId:      productID,
        OrganizationId: s.orgID,
    })
    
    // Calculate buffer using profile factors and lead time
    return s.calculateBuffer(
        prodResp.Product,
        profileResp.BufferProfile,
        getPrimaryLeadTime(suppResp.ProductSuppliers),
    )
}
```

---

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| `NOT_FOUND` (5) | Product/Supplier/Profile not found |
| `ALREADY_EXISTS` (6) | SKU/Code already exists |
| `INVALID_ARGUMENT` (3) | Invalid request parameters |
| `FAILED_PRECONDITION` (9) | Cannot delete - has dependencies |
| `INTERNAL` (13) | Server error |

---

## Connection Example

```go
import (
    catalogv1 "github.com/giia/giia-core-engine/services/catalog-service/api/proto/gen/go/catalog/v1"
    "google.golang.org/grpc"
)

func NewCatalogClient() (catalogv1.CatalogServiceClient, error) {
    conn, err := grpc.Dial("catalog-service:9082", grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return catalogv1.NewCatalogServiceClient(conn), nil
}
```

---

**Related Documentation:**
- [Catalog Service OpenAPI](/services/catalog-service/docs/openapi.yaml)
- [gRPC Contracts Overview](/docs/api/GRPC_CONTRACTS.md)
