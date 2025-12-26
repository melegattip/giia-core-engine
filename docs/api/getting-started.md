# Getting Started with GIIA APIs

This guide will help you integrate with the GIIA platform APIs in under 2 hours.

---

## Prerequisites

- API credentials (contact your administrator)
- Organization ID for your tenant
- HTTP client (curl, Postman, or your language of choice)

---

## Step 1: Authentication

### 1.1 Login to Get Tokens

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your.email@company.com",
    "password": "your_password"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "your.email@company.com",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### 1.2 Store Your Credentials

```bash
# Save for subsequent requests
export ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIs..."
export ORG_ID="550e8400-e29b-41d4-a716-446655440000"
```

### 1.3 Token Refresh

Access tokens expire after 15 minutes. Refresh before expiry:

```bash
curl -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Cookie: refresh_token=..."
```

---

## Step 2: Making API Requests

### Standard Request Headers

```bash
-H "Authorization: Bearer $ACCESS_TOKEN"
-H "X-Organization-ID: $ORG_ID"
-H "Content-Type: application/json"
```

### Example: List Products

```bash
curl http://localhost:8082/api/v1/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: $ORG_ID"
```

**Response:**
```json
{
  "products": [
    {
      "id": "abc123",
      "sku": "PROD-001",
      "name": "Widget A",
      "status": "active"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

---

## Step 3: Common Operations

### 3.1 Create a Product

```bash
curl -X POST http://localhost:8082/api/v1/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: $ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "PROD-002",
    "name": "Widget B",
    "category": "Electronics",
    "unit_of_measure": "units"
  }'
```

### 3.2 Create a Purchase Order

```bash
curl -X POST http://localhost:8084/api/v1/purchase-orders \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: $ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "po_number": "PO-2024-001",
    "supplier_id": "supplier-uuid",
    "order_date": "2024-01-15",
    "expected_arrival_date": "2024-01-25",
    "line_items": [
      {
        "product_id": "product-uuid",
        "quantity": 100,
        "unit_cost": 25.50
      }
    ]
  }'
```

### 3.3 Check Buffer Status

```bash
curl http://localhost:8083/api/v1/buffers/product-uuid \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: $ORG_ID"
```

**Response:**
```json
{
  "buffer": {
    "product_id": "product-uuid",
    "zone": "yellow",
    "net_flow_position": 150.5,
    "buffer_penetration": 0.45,
    "red_zone": 100,
    "yellow_zone": 150,
    "green_zone": 200,
    "on_hand": 180,
    "on_order": 50,
    "qualified_demand": 79.5
  }
}
```

### 3.4 Get Analytics KPIs

```bash
curl http://localhost:8085/api/v1/analytics/snapshot \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: $ORG_ID"
```

### 3.5 Subscribe to Notifications (WebSocket)

```javascript
const ws = new WebSocket(
  `ws://localhost:8086/ws/notifications?token=${ACCESS_TOKEN}`
);

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('New notification:', notification);
};
```

---

## Step 4: Error Handling

### Error Response Format

```json
{
  "error_code": "NOT_FOUND",
  "message": "Product not found"
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `BAD_REQUEST` | Invalid request parameters |
| 400 | `VALIDATION_ERROR` | Field validation failed |
| 401 | `UNAUTHORIZED` | Invalid or missing token |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |

### Retry Strategy

```python
import time
import requests

def api_call_with_retry(url, headers, max_retries=3):
    for attempt in range(max_retries):
        response = requests.get(url, headers=headers)
        
        if response.status_code == 429:
            wait = 2 ** attempt  # Exponential backoff
            time.sleep(wait)
            continue
            
        if response.status_code == 401:
            # Refresh token and retry
            headers['Authorization'] = f'Bearer {refresh_token()}'
            continue
            
        return response
    
    raise Exception("Max retries exceeded")
```

---

## Step 5: SDK Examples

### Go Client

```go
package main

import (
    "context"
    "fmt"
    "net/http"
)

type GIIAClient struct {
    baseURL     string
    accessToken string
    orgID       string
    httpClient  *http.Client
}

func NewGIIAClient(baseURL, token, orgID string) *GIIAClient {
    return &GIIAClient{
        baseURL:     baseURL,
        accessToken: token,
        orgID:       orgID,
        httpClient:  &http.Client{},
    }
}

func (c *GIIAClient) ListProducts(ctx context.Context) ([]Product, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", 
        c.baseURL+"/api/v1/products", nil)
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    req.Header.Set("X-Organization-ID", c.orgID)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Parse response...
    return products, nil
}
```

### Python Client

```python
import requests

class GIIAClient:
    def __init__(self, base_url, access_token, org_id):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {access_token}',
            'X-Organization-ID': org_id,
            'Content-Type': 'application/json'
        })
    
    def list_products(self, page=1, page_size=20):
        response = self.session.get(
            f'{self.base_url}/api/v1/products',
            params={'page': page, 'page_size': page_size}
        )
        response.raise_for_status()
        return response.json()
    
    def create_purchase_order(self, po_data):
        response = self.session.post(
            f'{self.base_url}/api/v1/purchase-orders',
            json=po_data
        )
        response.raise_for_status()
        return response.json()

# Usage
client = GIIAClient(
    base_url='http://localhost:8084',
    access_token='your_token',
    org_id='your_org_id'
)

products = client.list_products()
```

### JavaScript/TypeScript Client

```typescript
class GIIAClient {
  private baseUrl: string;
  private headers: HeadersInit;

  constructor(baseUrl: string, accessToken: string, orgId: string) {
    this.baseUrl = baseUrl;
    this.headers = {
      'Authorization': `Bearer ${accessToken}`,
      'X-Organization-ID': orgId,
      'Content-Type': 'application/json'
    };
  }

  async listProducts(): Promise<Product[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/products`, {
      headers: this.headers
    });
    
    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }
    
    const data = await response.json();
    return data.products;
  }

  async createPurchaseOrder(order: CreatePORequest): Promise<PurchaseOrder> {
    const response = await fetch(`${this.baseUrl}/api/v1/purchase-orders`, {
      method: 'POST',
      headers: this.headers,
      body: JSON.stringify(order)
    });
    
    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }
    
    return response.json();
  }
}
```

---

## Next Steps

1. **Explore OpenAPI Specs** - Import into Postman for interactive testing
2. **Set Up Webhooks** - Subscribe to events via WebSocket
3. **Review Full Docs** - [API Reference](./index.md)
4. **Authentication Details** - [Authentication Guide](./authentication.md)
5. **Code Examples** - [Examples Directory](./examples/)

---

## Need Help?

- Check the [Authentication Guide](./authentication.md) for token issues
- Review [Error Handling](#step-4-error-handling) for debugging
- Contact support for API access issues
