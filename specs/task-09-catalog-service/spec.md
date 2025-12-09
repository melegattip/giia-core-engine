# Feature Specification: Catalog Service with Clean Architecture

**Created**: 2025-12-09

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Product Master Data Management (Priority: P1)

As a supply chain manager, I need to create and manage product catalog (SKUs, descriptions, units) so that I can track inventory items in the DDMRP system.

**Why this priority**: Absolutely critical. Products are the foundation of inventory management. Without product master data, no other DDMRP features can work. Blocks everything.

**Independent Test**: Can be fully tested by creating products via REST API, retrieving them, updating properties, and verifying changes persist. Delivers standalone value: users can manage product catalog.

**Acceptance Scenarios**:

1. **Scenario**: Create product
   - **Given** a supply chain manager is authenticated
   - **When** manager submits POST /api/v1/products with SKU, name, description, unit
   - **Then** system creates product and returns 201 Created with product ID

2. **Scenario**: List products with pagination
   - **Given** organization has 100 products
   - **When** manager requests GET /api/v1/products?page=1&size=20
   - **Then** system returns first 20 products with pagination metadata

3. **Scenario**: Update product details
   - **Given** a product exists with ID "prod-123"
   - **When** manager submits PUT /api/v1/products/prod-123 with updated fields
   - **Then** system updates product and returns 200 OK with updated data

4. **Scenario**: Multi-tenant product isolation
   - **Given** user belongs to Organization A
   - **When** user queries products
   - **Then** system returns only products belonging to Organization A

---

### User Story 2 - Supplier Management (Priority: P1)

As a supply chain manager, I need to manage supplier information (name, lead time, reliability) so that I can configure replenishment parameters for DDMRP buffers.

**Why this priority**: Critical for DDMRP calculations. Lead time from suppliers directly affects buffer sizing. Required before implementing buffer management.

**Independent Test**: Can be tested by creating suppliers via API, associating them with products, and verifying supplier data is returned with product queries.

**Acceptance Scenarios**:

1. **Scenario**: Create supplier
   - **Given** a manager is authenticated
   - **When** manager submits POST /api/v1/suppliers with name, lead_time_days, reliability_rating
   - **Then** system creates supplier and returns 201 Created

2. **Scenario**: Associate supplier with product
   - **Given** product "prod-123" and supplier "sup-456" exist
   - **When** manager creates product-supplier relationship
   - **Then** system links supplier to product with lead time and pricing

3. **Scenario**: Query products with supplier info
   - **Given** products have associated suppliers
   - **When** manager requests GET /api/v1/products?include=suppliers
   - **Then** system returns products with embedded supplier details

---

### User Story 3 - Buffer Profile Templates (Priority: P2)

As a DDMRP consultant, I need to create buffer profile templates (e.g., "Fast Movers", "Slow Movers") so that I can quickly assign buffer policies to multiple products.

**Why this priority**: Important for usability and implementation speed. Makes DDMRP configuration faster. Can work without templates initially by configuring each product individually.

**Independent Test**: Can be tested by creating buffer profile template, assigning it to multiple products, and verifying buffer calculations use template parameters.

**Acceptance Scenarios**:

1. **Scenario**: Create buffer profile template
   - **Given** a DDMRP consultant is authenticated
   - **When** consultant creates template with lead_time_factor, variability_factor
   - **Then** system saves template for organization

2. **Scenario**: Assign template to products
   - **Given** template "Fast Movers" exists
   - **When** consultant assigns template to 50 products in bulk
   - **Then** system applies template to all products

3. **Scenario**: Override template for specific product
   - **Given** product uses template but needs custom parameters
   - **When** consultant sets product-specific buffer parameters
   - **Then** product parameters override template defaults

---

### User Story 4 - Product Search and Filtering (Priority: P3)

As a supply chain manager, I need to search products by SKU, name, or category and filter by attributes so that I can quickly find items in large catalogs.

**Why this priority**: Nice-to-have for user experience. Critical for organizations with 1000+ products. Can launch MVP with simple list and manual scrolling.

**Independent Test**: Can be tested by creating products with various attributes, then performing search queries and verifying correct results are returned.

**Acceptance Scenarios**:

1. **Scenario**: Search by SKU or name
   - **Given** catalog has 500 products
   - **When** manager searches for "widget"
   - **Then** system returns all products with "widget" in SKU or name

2. **Scenario**: Filter by category
   - **Given** products have category tags
   - **When** manager filters by category="electronics"
   - **Then** system returns only electronics products

---

### Edge Cases

- What happens when duplicate SKU is created in same organization?
- How to handle product deletion when product has active buffers?
- What happens when supplier lead time changes (update existing buffers)?
- How to handle product import from CSV (bulk operations)?
- What happens when product unit of measure changes (impact on inventory)?
- How to handle inactive/discontinued products (soft delete vs hard delete)?
- What happens when organization reaches product limit (if any)?

## Requirements *(mandatory)*

### Functional Requirements

#### Product Management
- **FR-001**: System MUST support CRUD operations for products (create, read, update, delete)
- **FR-002**: System MUST enforce unique SKU per organization (allow same SKU across different orgs)
- **FR-003**: System MUST store product attributes: SKU, name, description, category, unit_of_measure, status (active/inactive)
- **FR-004**: System MUST support product search by SKU, name, category
- **FR-005**: System MUST implement pagination for product lists (default 20 per page, max 100)
- **FR-006**: System MUST enforce multi-tenant isolation (organization_id filter on all queries)
- **FR-007**: System MUST support soft delete for products (mark as inactive)

#### Supplier Management
- **FR-008**: System MUST support CRUD operations for suppliers
- **FR-009**: System MUST store supplier attributes: name, code, lead_time_days, reliability_rating (0-100), contact_info
- **FR-010**: System MUST support product-supplier relationships (many-to-many)
- **FR-011**: System MUST track supplier-specific lead times and pricing per product

#### Buffer Profile Templates
- **FR-012**: System MUST support CRUD operations for buffer profile templates
- **FR-013**: System MUST store template parameters: name, lead_time_factor, variability_factor, target_service_level
- **FR-014**: System MUST support assigning templates to products (one template per product)
- **FR-015**: System MUST allow product-level parameter overrides

#### Clean Architecture Implementation
- **FR-016**: Service MUST follow Clean Architecture with clear separation: domain, use cases, adapters, infrastructure
- **FR-017**: Domain layer MUST NOT depend on external frameworks (GORM, HTTP libraries)
- **FR-018**: Use cases MUST implement business logic with dependency injection via interfaces
- **FR-019**: Infrastructure layer MUST implement repository pattern for data access
- **FR-020**: API layer MUST implement REST endpoints with OpenAPI documentation
- **FR-021**: Service MUST validate all inputs at entry points (HTTP handlers)
- **FR-022**: Service MUST use shared packages (pkg/config, pkg/logger, pkg/database, pkg/errors)

### Key Entities

- **Product**: SKU, name, description, category, unit_of_measure, status, organization_id, created_at, updated_at
- **Supplier**: Code, name, lead_time_days, reliability_rating, contact_info, organization_id
- **ProductSupplier**: product_id, supplier_id, lead_time_days, unit_cost, is_primary_supplier
- **BufferProfile**: name, description, lead_time_factor, variability_factor, target_service_level, organization_id

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: API endpoints respond in under 200ms (p95) for single-entity operations
- **SC-002**: Product list with 1000 items loads in under 500ms with pagination
- **SC-003**: 100% of API endpoints enforce multi-tenant isolation (verified by tests)
- **SC-004**: Service maintains >80% code coverage with unit and integration tests
- **SC-005**: Clean Architecture layers are correctly separated (verified by architecture tests)
- **SC-006**: Service can handle 1,000 concurrent requests without errors
- **SC-007**: OpenAPI documentation is auto-generated and up-to-date
- **SC-008**: Service successfully integrates with Auth service for token validation and permission checks
