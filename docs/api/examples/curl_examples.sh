#!/bin/bash
# GIIA Platform cURL Examples
# Source this file: source ./curl_examples.sh

# Configuration
export GIIA_BASE_URL="${GIIA_BASE_URL:-http://localhost}"
export GIIA_AUTH_PORT=":8081"
export GIIA_CATALOG_PORT=":8082"
export GIIA_DDMRP_PORT=":8083"
export GIIA_EXEC_PORT=":8084"
export GIIA_ANALYTICS_PORT=":8085"
export GIIA_AI_PORT=":8086"

# Token storage (set after login)
export GIIA_ACCESS_TOKEN=""
export GIIA_ORG_ID=""

# ========== Helper Functions ==========

giia_headers() {
    echo -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
         -H "X-Organization-ID: $GIIA_ORG_ID" \
         -H "Content-Type: application/json"
}

# ========== Authentication ==========

giia_login() {
    local email="${1:-$GIIA_EMAIL}"
    local password="${2:-$GIIA_PASSWORD}"
    
    echo "Logging in as $email..."
    
    response=$(curl -s -X POST "${GIIA_BASE_URL}${GIIA_AUTH_PORT}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$email\", \"password\": \"$password\"}")
    
    GIIA_ACCESS_TOKEN=$(echo "$response" | jq -r '.access_token')
    GIIA_ORG_ID=$(echo "$response" | jq -r '.user.organization_id')
    
    if [ "$GIIA_ACCESS_TOKEN" != "null" ]; then
        echo "✓ Login successful"
        echo "  Token: ${GIIA_ACCESS_TOKEN:0:20}..."
        echo "  Org ID: $GIIA_ORG_ID"
        export GIIA_ACCESS_TOKEN
        export GIIA_ORG_ID
    else
        echo "✗ Login failed:"
        echo "$response" | jq
    fi
}

giia_logout() {
    curl -s -X POST "${GIIA_BASE_URL}${GIIA_AUTH_PORT}/api/v1/auth/logout" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" | jq
    
    GIIA_ACCESS_TOKEN=""
    GIIA_ORG_ID=""
    echo "✓ Logged out"
}

giia_me() {
    curl -s "${GIIA_BASE_URL}${GIIA_AUTH_PORT}/api/v1/users/me" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" | jq
}

# ========== Health Checks ==========

giia_health() {
    echo "=== Service Health ==="
    echo -n "Auth:      "; curl -s "${GIIA_BASE_URL}${GIIA_AUTH_PORT}/health" | jq -r '.status'
    echo -n "Catalog:   "; curl -s "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/health" | jq -r '.status'
    echo -n "DDMRP:     "; curl -s "${GIIA_BASE_URL}${GIIA_DDMRP_PORT}/health" | jq -r '.status'
    echo -n "Execution: "; curl -s "${GIIA_BASE_URL}${GIIA_EXEC_PORT}/health" | jq -r '.status'
    echo -n "Analytics: "; curl -s "${GIIA_BASE_URL}${GIIA_ANALYTICS_PORT}/health" | jq -r '.status'
    echo -n "AI Hub:    "; curl -s "${GIIA_BASE_URL}${GIIA_AI_PORT}/health" | jq -r '.status'
}

# ========== Products ==========

giia_list_products() {
    local page="${1:-1}"
    local page_size="${2:-10}"
    
    curl -s "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/api/v1/products?page=$page&page_size=$page_size" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_get_product() {
    local product_id="$1"
    
    curl -s "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/api/v1/products/$product_id" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_create_product() {
    local sku="$1"
    local name="$2"
    local category="${3:-General}"
    
    curl -s -X POST "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/api/v1/products" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" \
        -H "Content-Type: application/json" \
        -d "{
            \"sku\": \"$sku\",
            \"name\": \"$name\",
            \"category\": \"$category\",
            \"unit_of_measure\": \"units\"
        }" | jq
}

giia_search_products() {
    local query="$1"
    
    curl -s "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/api/v1/products?search=$query" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq '.products[] | {sku, name}'
}

# ========== Suppliers ==========

giia_list_suppliers() {
    curl -s "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/api/v1/suppliers" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_create_supplier() {
    local code="$1"
    local name="$2"
    
    curl -s -X POST "${GIIA_BASE_URL}${GIIA_CATALOG_PORT}/api/v1/suppliers" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" \
        -H "Content-Type: application/json" \
        -d "{
            \"code\": \"$code\",
            \"name\": \"$name\"
        }" | jq
}

# ========== Buffers ==========

giia_get_buffer() {
    local product_id="$1"
    
    curl -s "${GIIA_BASE_URL}${GIIA_DDMRP_PORT}/api/v1/buffers/$product_id" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_list_buffers() {
    local zone="${1:-}"
    local params=""
    [ -n "$zone" ] && params="?zone=$zone"
    
    curl -s "${GIIA_BASE_URL}${GIIA_DDMRP_PORT}/api/v1/buffers$params" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_calculate_buffer() {
    local product_id="$1"
    
    curl -s -X POST "${GIIA_BASE_URL}${GIIA_DDMRP_PORT}/api/v1/buffers/$product_id/calculate" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

# ========== Purchase Orders ==========

giia_list_pos() {
    local status="${1:-}"
    local params=""
    [ -n "$status" ] && params="?status=$status"
    
    curl -s "${GIIA_BASE_URL}${GIIA_EXEC_PORT}/api/v1/purchase-orders$params" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_get_po() {
    local po_id="$1"
    
    curl -s "${GIIA_BASE_URL}${GIIA_EXEC_PORT}/api/v1/purchase-orders/$po_id" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_create_po() {
    local supplier_id="$1"
    local product_id="$2"
    local quantity="${3:-100}"
    local unit_cost="${4:-10.00}"
    
    local po_number="PO-$(date +%Y%m%d%H%M%S)"
    local order_date=$(date +%Y-%m-%d)
    local arrival_date=$(date -d "+14 days" +%Y-%m-%d 2>/dev/null || date -v+14d +%Y-%m-%d)
    
    curl -s -X POST "${GIIA_BASE_URL}${GIIA_EXEC_PORT}/api/v1/purchase-orders" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" \
        -H "Content-Type: application/json" \
        -d "{
            \"po_number\": \"$po_number\",
            \"supplier_id\": \"$supplier_id\",
            \"order_date\": \"$order_date\",
            \"expected_arrival_date\": \"$arrival_date\",
            \"line_items\": [{
                \"product_id\": \"$product_id\",
                \"quantity\": $quantity,
                \"unit_cost\": $unit_cost
            }]
        }" | jq
}

# ========== Inventory ==========

giia_get_inventory() {
    local product_id="${1:-}"
    local params=""
    [ -n "$product_id" ] && params="?product_id=$product_id"
    
    curl -s "${GIIA_BASE_URL}${GIIA_EXEC_PORT}/api/v1/inventory/balances$params" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

# ========== Analytics ==========

giia_kpi_snapshot() {
    curl -s "${GIIA_BASE_URL}${GIIA_ANALYTICS_PORT}/api/v1/analytics/snapshot" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_inventory_rotation() {
    curl -s "${GIIA_BASE_URL}${GIIA_ANALYTICS_PORT}/api/v1/analytics/inventory-rotation" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

# ========== Notifications ==========

giia_notifications() {
    local unread="${1:-false}"
    
    curl -s "${GIIA_BASE_URL}${GIIA_AI_PORT}/api/v1/notifications?unread_only=$unread" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

giia_unread_count() {
    curl -s "${GIIA_BASE_URL}${GIIA_AI_PORT}/api/v1/notifications/unread-count" \
        -H "Authorization: Bearer $GIIA_ACCESS_TOKEN" \
        -H "X-Organization-ID: $GIIA_ORG_ID" | jq
}

# ========== Quick Demo ==========

giia_demo() {
    echo "=== GIIA API Demo ==="
    echo ""
    
    # Health check
    echo "1. Checking service health..."
    giia_health
    echo ""
    
    # Login
    echo "2. Logging in..."
    giia_login
    echo ""
    
    # List products
    echo "3. Listing products..."
    giia_list_products 1 3
    echo ""
    
    # Get KPIs
    echo "4. Getting KPI snapshot..."
    giia_kpi_snapshot
    echo ""
    
    # Check notifications
    echo "5. Checking notifications..."
    giia_unread_count
    echo ""
    
    echo "=== Demo Complete ==="
}

# Print available functions
giia_help() {
    echo "GIIA cURL Helper Functions"
    echo "=========================="
    echo ""
    echo "Authentication:"
    echo "  giia_login [email] [password]  - Login and get token"
    echo "  giia_logout                    - Logout"
    echo "  giia_me                        - Get current user"
    echo ""
    echo "Health:"
    echo "  giia_health                    - Check all services"
    echo ""
    echo "Products:"
    echo "  giia_list_products [page] [size] - List products"
    echo "  giia_get_product <id>            - Get product by ID"
    echo "  giia_create_product <sku> <name> - Create product"
    echo "  giia_search_products <query>     - Search products"
    echo ""
    echo "Suppliers:"
    echo "  giia_list_suppliers            - List suppliers"
    echo "  giia_create_supplier <code> <name> - Create supplier"
    echo ""
    echo "Buffers:"
    echo "  giia_list_buffers [zone]       - List buffers"
    echo "  giia_get_buffer <product_id>   - Get buffer status"
    echo "  giia_calculate_buffer <id>     - Recalculate buffer"
    echo ""
    echo "Purchase Orders:"
    echo "  giia_list_pos [status]         - List purchase orders"
    echo "  giia_get_po <id>               - Get PO details"
    echo "  giia_create_po <supplier> <product> [qty] [cost]"
    echo ""
    echo "Analytics:"
    echo "  giia_kpi_snapshot              - Get KPI snapshot"
    echo "  giia_inventory_rotation        - Get rotation metrics"
    echo ""
    echo "Notifications:"
    echo "  giia_notifications [unread]    - List notifications"
    echo "  giia_unread_count              - Get unread count"
    echo ""
    echo "Demo:"
    echo "  giia_demo                      - Run quick demo"
    echo ""
}

echo "GIIA cURL helpers loaded. Run 'giia_help' for available commands."
