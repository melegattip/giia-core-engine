/**
 * GIIA Platform JavaScript/TypeScript Client Example
 * 
 * This module provides a complete JavaScript client for GIIA APIs.
 * Works in both Node.js and browser environments.
 * 
 * Usage:
 *   const client = new GIIAClient('http://localhost');
 *   await client.login('user@example.com', 'password');
 *   const products = await client.catalog.listProducts();
 */

// ========== Types ==========

/**
 * @typedef {Object} User
 * @property {string} id
 * @property {string} email
 * @property {string} firstName
 * @property {string} lastName
 * @property {string} organizationId
 * @property {string[]} roles
 */

/**
 * @typedef {Object} Product
 * @property {string} id
 * @property {string} sku
 * @property {string} name
 * @property {string} category
 * @property {string} unitOfMeasure
 * @property {string} status
 */

/**
 * @typedef {Object} Buffer
 * @property {string} productId
 * @property {string} zone
 * @property {number} netFlowPosition
 * @property {number} bufferPenetration
 * @property {number} redZone
 * @property {number} yellowZone
 * @property {number} greenZone
 */

// ========== Configuration ==========

const PORTS = {
    AUTH: 8081,
    CATALOG: 8082,
    DDMRP: 8083,
    EXECUTION: 8084,
    ANALYTICS: 8085,
    AI_HUB: 8086,
};

// ========== Main Client ==========

class GIIAClient {
    /**
     * @param {string} baseUrl - Base URL of the API server
     */
    constructor(baseUrl = 'http://localhost') {
        this.baseUrl = baseUrl;
        this.accessToken = null;
        this.orgId = null;
        this.user = null;

        // Service-specific clients
        this.auth = new AuthClient(this);
        this.catalog = new CatalogClient(this);
        this.ddmrp = new DDMRPClient(this);
        this.execution = new ExecutionClient(this);
        this.analytics = new AnalyticsClient(this);
        this.notifications = new NotificationsClient(this);
    }

    /**
     * Get headers for API requests
     * @returns {Object}
     */
    getHeaders() {
        const headers = {
            'Content-Type': 'application/json',
        };
        if (this.accessToken) {
            headers['Authorization'] = `Bearer ${this.accessToken}`;
        }
        if (this.orgId) {
            headers['X-Organization-ID'] = this.orgId;
        }
        return headers;
    }

    /**
     * Make an API request
     * @param {string} port - Service port
     * @param {string} path - API path
     * @param {Object} options - Fetch options
     */
    async request(port, path, options = {}) {
        const url = `${this.baseUrl}:${port}${path}`;
        const response = await fetch(url, {
            ...options,
            headers: {
                ...this.getHeaders(),
                ...options.headers,
            },
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Unknown error' }));
            throw new Error(`API Error ${response.status}: ${error.message || error.error_code}`);
        }

        return response.json();
    }

    /**
     * Login and get access token
     * @param {string} email
     * @param {string} password
     * @returns {Promise<User>}
     */
    async login(email, password) {
        const data = await this.request(PORTS.AUTH, '/api/v1/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
        });

        this.accessToken = data.access_token;
        this.orgId = data.user.organization_id;
        this.user = {
            id: data.user.id,
            email: data.user.email,
            firstName: data.user.first_name,
            lastName: data.user.last_name,
            organizationId: data.user.organization_id,
            roles: data.user.roles || [],
        };

        return this.user;
    }

    /**
     * Logout
     */
    async logout() {
        await this.request(PORTS.AUTH, '/api/v1/auth/logout', { method: 'POST' });
        this.accessToken = null;
        this.orgId = null;
        this.user = null;
    }
}

// ========== Service Clients ==========

class AuthClient {
    constructor(client) {
        this.client = client;
    }

    async getCurrentUser() {
        return this.client.request(PORTS.AUTH, '/api/v1/users/me');
    }

    async refreshToken() {
        const data = await this.client.request(PORTS.AUTH, '/api/v1/auth/refresh', {
            method: 'POST',
        });
        this.client.accessToken = data.access_token;
        return data;
    }
}

class CatalogClient {
    constructor(client) {
        this.client = client;
    }

    /**
     * List products
     * @param {Object} params
     * @returns {Promise<{products: Product[], total: number}>}
     */
    async listProducts({ page = 1, pageSize = 20, status, category } = {}) {
        const params = new URLSearchParams({ page, page_size: pageSize });
        if (status) params.append('status', status);
        if (category) params.append('category', category);

        return this.client.request(PORTS.CATALOG, `/api/v1/products?${params}`);
    }

    /**
     * Get a product
     * @param {string} id
     * @returns {Promise<Product>}
     */
    async getProduct(id) {
        return this.client.request(PORTS.CATALOG, `/api/v1/products/${id}`);
    }

    /**
     * Create a product
     * @param {Object} product
     * @returns {Promise<Product>}
     */
    async createProduct({ sku, name, category, unitOfMeasure = 'units' }) {
        return this.client.request(PORTS.CATALOG, '/api/v1/products', {
            method: 'POST',
            body: JSON.stringify({
                sku,
                name,
                category,
                unit_of_measure: unitOfMeasure,
            }),
        });
    }

    async listSuppliers() {
        return this.client.request(PORTS.CATALOG, '/api/v1/suppliers');
    }

    async listBufferProfiles() {
        return this.client.request(PORTS.CATALOG, '/api/v1/buffer-profiles');
    }
}

class DDMRPClient {
    constructor(client) {
        this.client = client;
    }

    /**
     * Get buffer status
     * @param {string} productId
     * @returns {Promise<{buffer: Buffer}>}
     */
    async getBuffer(productId) {
        return this.client.request(PORTS.DDMRP, `/api/v1/buffers/${productId}`);
    }

    async listBuffers({ zone, alertLevel } = {}) {
        const params = new URLSearchParams();
        if (zone) params.append('zone', zone);
        if (alertLevel) params.append('alert_level', alertLevel);
        return this.client.request(PORTS.DDMRP, `/api/v1/buffers?${params}`);
    }

    async calculateBuffer(productId) {
        return this.client.request(PORTS.DDMRP, `/api/v1/buffers/${productId}/calculate`, {
            method: 'POST',
        });
    }
}

class ExecutionClient {
    constructor(client) {
        this.client = client;
    }

    async listPurchaseOrders({ status } = {}) {
        const params = new URLSearchParams();
        if (status) params.append('status', status);
        return this.client.request(PORTS.EXECUTION, `/api/v1/purchase-orders?${params}`);
    }

    async createPurchaseOrder({ poNumber, supplierId, lineItems, expectedDays = 14 }) {
        const orderDate = new Date().toISOString().split('T')[0];
        const arrivalDate = new Date(Date.now() + expectedDays * 24 * 60 * 60 * 1000)
            .toISOString().split('T')[0];

        return this.client.request(PORTS.EXECUTION, '/api/v1/purchase-orders', {
            method: 'POST',
            body: JSON.stringify({
                po_number: poNumber,
                supplier_id: supplierId,
                order_date: orderDate,
                expected_arrival_date: arrivalDate,
                line_items: lineItems,
            }),
        });
    }

    async getInventoryBalances(productId) {
        const params = productId ? `?product_id=${productId}` : '';
        return this.client.request(PORTS.EXECUTION, `/api/v1/inventory/balances${params}`);
    }
}

class AnalyticsClient {
    constructor(client) {
        this.client = client;
    }

    async getKPISnapshot() {
        return this.client.request(PORTS.ANALYTICS, '/api/v1/analytics/snapshot');
    }

    async getDaysInInventory() {
        return this.client.request(PORTS.ANALYTICS, '/api/v1/analytics/days-in-inventory');
    }

    async getInventoryRotation() {
        return this.client.request(PORTS.ANALYTICS, '/api/v1/analytics/inventory-rotation');
    }
}

class NotificationsClient {
    constructor(client) {
        this.client = client;
        this.ws = null;
    }

    async list({ unreadOnly = false, type, priority } = {}) {
        const params = new URLSearchParams({ unread_only: unreadOnly });
        if (type) params.append('type', type);
        if (priority) params.append('priority', priority);
        return this.client.request(PORTS.AI_HUB, `/api/v1/notifications?${params}`);
    }

    async getUnreadCount() {
        return this.client.request(PORTS.AI_HUB, '/api/v1/notifications/unread-count');
    }

    async markAsRead(id) {
        return this.client.request(PORTS.AI_HUB, `/api/v1/notifications/${id}`, {
            method: 'PATCH',
            body: JSON.stringify({ read: true }),
        });
    }

    /**
     * Subscribe to real-time notifications
     * @param {Function} onMessage - Callback for new notifications
     * @param {Function} onError - Error callback
     */
    subscribe(onMessage, onError) {
        const wsUrl = `ws://${this.client.baseUrl.replace('http://', '')}:${PORTS.AI_HUB}/ws/notifications?token=${this.client.accessToken}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            onMessage(data);
        };

        this.ws.onerror = (error) => {
            if (onError) onError(error);
        };

        this.ws.onclose = () => {
            console.log('WebSocket connection closed');
        };
    }

    unsubscribe() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

// ========== Example Usage ==========

async function main() {
    console.log('='.repeat(50));
    console.log('GIIA Platform JavaScript Client Example');
    console.log('='.repeat(50));

    const client = new GIIAClient(process.env.GIIA_API_URL || 'http://localhost');

    try {
        // Step 1: Login
        console.log('\n1. Logging in...');
        const user = await client.login(
            process.env.GIIA_EMAIL || 'demo@example.com',
            process.env.GIIA_PASSWORD || 'password'
        );
        console.log(`   ✓ Logged in as ${user.email}`);

        // Step 2: List products
        console.log('\n2. Listing products...');
        const { products, total } = await client.catalog.listProducts({ pageSize: 5 });
        console.log(`   ✓ Found ${total} products`);
        products.slice(0, 3).forEach(p => {
            console.log(`      - ${p.sku}: ${p.name}`);
        });

        // Step 3: Create a product
        console.log('\n3. Creating product...');
        const newProduct = await client.catalog.createProduct({
            sku: `JS-${Date.now()}`,
            name: 'JavaScript Demo Product',
            category: 'Demo',
        });
        console.log(`   ✓ Created: ${newProduct.sku}`);

        // Step 4: Get buffer
        console.log('\n4. Checking buffer status...');
        if (products.length > 0) {
            const { buffer } = await client.ddmrp.getBuffer(products[0].id);
            console.log(`   ✓ Buffer for ${products[0].sku}:`);
            console.log(`      Zone: ${buffer.zone}`);
            console.log(`      NFP: ${buffer.net_flow_position?.toFixed(2)}`);
        }

        // Step 5: Get KPIs
        console.log('\n5. Getting KPI snapshot...');
        const kpis = await client.analytics.getKPISnapshot();
        console.log(`   ✓ Service Level: ${((kpis.service_level || 0) * 100).toFixed(1)}%`);

        // Step 6: Get notifications
        console.log('\n6. Checking notifications...');
        const { count } = await client.notifications.getUnreadCount();
        console.log(`   ✓ ${count} unread notifications`);

        console.log('\n' + '='.repeat(50));
        console.log('Example complete!');
        console.log('='.repeat(50));

    } catch (error) {
        console.error('Error:', error.message);
        process.exit(1);
    }
}

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { GIIAClient, PORTS };
}

// Run if executed directly
if (typeof require !== 'undefined' && require.main === module) {
    main();
}
