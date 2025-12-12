# Feature Specification: gRPC Server for Auth Service

**Created**: 2025-12-09

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Token Validation for Service-to-Service Communication (Priority: P1)

As a backend microservice (catalog, ddmrp, execution), I need to validate JWT tokens via gRPC so that I can verify user identity and permissions before processing requests.

**Why this priority**: Critical for microservice security. Without token validation, services cannot verify who is making requests. Blocks all inter-service communication.

**Independent Test**: Can be fully tested by calling ValidateToken gRPC method with valid JWT, and verifying response includes user_id, organization_id, and roles. Invalid tokens should return appropriate errors. Delivers standalone value: services can authenticate requests.

**Acceptance Scenarios**:

1. **Scenario**: Valid token validation
   - **Given** a service receives request with valid JWT token in metadata
   - **When** service calls ValidateToken gRPC method
   - **Then** Auth service returns user_id, organization_id, email, and roles

2. **Scenario**: Invalid token rejection
   - **Given** a service receives request with expired or malformed token
   - **When** service calls ValidateToken gRPC method
   - **Then** Auth service returns UNAUTHENTICATED error

3. **Scenario**: Token validation performance
   - **Given** 100 concurrent token validation requests
   - **When** all requests are processed
   - **Then** each validation completes in under 10ms (p95)

---

### User Story 2 - Permission Checking for Authorization (Priority: P1)

As a backend microservice, I need to check if a user has specific permissions via gRPC so that I can enforce access control for operations.

**Why this priority**: Critical for RBAC enforcement. Services must be able to check permissions before allowing operations. Required for security.

**Independent Test**: Can be tested by calling CheckPermission gRPC method with user_id and permission string, and verifying correct boolean response based on user's assigned roles and permissions.

**Acceptance Scenarios**:

1. **Scenario**: Permission granted
   - **Given** user has role with permission "catalog:products:write"
   - **When** service calls CheckPermission(user_id, "catalog:products:write")
   - **Then** Auth service returns allowed=true

2. **Scenario**: Permission denied
   - **Given** user has role without permission "ddmrp:buffers:delete"
   - **When** service calls CheckPermission(user_id, "ddmrp:buffers:delete")
   - **Then** Auth service returns allowed=false

3. **Scenario**: Batch permission check
   - **Given** service needs to check multiple permissions at once
   - **When** service calls BatchCheckPermissions with array of permissions
   - **Then** Auth service returns array of boolean results in single round-trip

---

### User Story 3 - User Info Retrieval (Priority: P2)

As a backend microservice, I need to retrieve user details (name, email, organization) via gRPC so that I can display user information and implement user-specific logic.

**Why this priority**: Important for features that need user details. Can work without this initially by including more claims in JWT, but gRPC call provides fresh data.

**Independent Test**: Can be tested by calling GetUser gRPC method with user_id and receiving full user profile including name, email, organization details, and roles.

**Acceptance Scenarios**:

1. **Scenario**: Get user by ID
   - **Given** a user exists with ID "12345"
   - **When** service calls GetUser(user_id="12345")
   - **Then** Auth service returns user profile with all fields

2. **Scenario**: Get non-existent user
   - **Given** user ID "99999" does not exist
   - **When** service calls GetUser(user_id="99999")
   - **Then** Auth service returns NOT_FOUND error

3. **Scenario**: Tenant isolation in user lookup
   - **Given** user belongs to organization_id "org-1"
   - **When** service calls GetUser from context of "org-2"
   - **Then** Auth service returns PERMISSION_DENIED error

---

### User Story 4 - gRPC Health Checks (Priority: P3)

As a DevOps engineer, I need gRPC health check endpoint so that Kubernetes can monitor service health and restart unhealthy pods automatically.

**Why this priority**: Important for production reliability but not blocking for development. Can use HTTP health endpoint initially.

**Independent Test**: Can be tested by calling gRPC Health Check service and verifying SERVING status is returned.

**Acceptance Scenarios**:

1. **Scenario**: Health check when service is healthy
   - **Given** Auth service is running and database is connected
   - **When** Kubernetes calls Health Check
   - **Then** service returns SERVING status

2. **Scenario**: Health check when service is unhealthy
   - **Given** Auth service cannot connect to database
   - **When** Kubernetes calls Health Check
   - **Then** service returns NOT_SERVING status

---

### Edge Cases

- What happens when gRPC server is under high load (10,000+ req/s)?
- How to handle partial database outages during gRPC calls?
- What happens when token validation cache is invalidated?
- How to version gRPC API for backward compatibility?
- How to handle gRPC deadlines and timeouts?
- What happens when gRPC client uses wrong service name or method?
- How to secure gRPC communication between services (TLS)?
- How to handle gRPC metadata for tracing (request_id, span_id)?

## Requirements *(mandatory)*

### Functional Requirements

#### gRPC Server Setup
- **FR-001**: System MUST expose gRPC server on port 9091 (separate from HTTP port 8081)
- **FR-002**: System MUST use Protocol Buffers (proto3) for service definitions
- **FR-003**: System MUST implement gRPC server with graceful shutdown
- **FR-004**: System MUST support gRPC reflection for debugging with grpcurl
- **FR-005**: System MUST implement standard gRPC Health Check service

#### Token Validation Service
- **FR-006**: System MUST provide ValidateToken(token: string) -> UserInfo RPC method
- **FR-007**: ValidateToken MUST verify JWT signature, expiration, and claims
- **FR-008**: ValidateToken MUST return user_id, organization_id, email, roles in response
- **FR-009**: ValidateToken MUST cache valid tokens in Redis (5-minute TTL)
- **FR-010**: ValidateToken MUST complete in under 10ms (p95) with caching

#### Permission Service
- **FR-011**: System MUST provide CheckPermission(user_id, permission) -> bool RPC method
- **FR-012**: System MUST provide BatchCheckPermissions(user_id, permissions[]) -> bool[] RPC method
- **FR-013**: CheckPermission MUST evaluate role-based permissions with inheritance
- **FR-014**: CheckPermission MUST cache permission results per user (5-minute TTL)
- **FR-015**: CheckPermission MUST complete in under 10ms (p95) with caching

#### User Info Service
- **FR-016**: System MUST provide GetUser(user_id) -> User RPC method
- **FR-017**: GetUser MUST return complete user profile (name, email, organization, roles)
- **FR-018**: GetUser MUST enforce tenant isolation based on organization_id

#### gRPC Middleware
- **FR-019**: System MUST implement gRPC interceptor for request logging
- **FR-020**: System MUST implement gRPC interceptor for error handling
- **FR-021**: System MUST implement gRPC interceptor for metrics (Prometheus)
- **FR-022**: System MUST implement gRPC interceptor for distributed tracing

### Key Entities

- **AuthServiceGRPC**: gRPC server exposing ValidateToken, CheckPermission, GetUser
- **ValidateTokenRequest**: token (string)
- **ValidateTokenResponse**: user_id, organization_id, email, roles[], permissions[]
- **CheckPermissionRequest**: user_id, organization_id, permission (string)
- **CheckPermissionResponse**: allowed (bool), reason (string)
- **GetUserRequest**: user_id
- **GetUserResponse**: User object with full profile

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: gRPC server handles 10,000 requests per second without errors
- **SC-002**: ValidateToken completes in under 10ms (p95) with caching enabled
- **SC-003**: CheckPermission completes in under 10ms (p95) with caching enabled
- **SC-004**: 100% of other microservices can successfully call Auth gRPC endpoints
- **SC-005**: gRPC server stays operational for 24+ hours without restart
- **SC-006**: gRPC error rate is below 0.1% under normal load
- **SC-007**: All gRPC methods have comprehensive unit and integration tests
- **SC-008**: gRPC API documentation is auto-generated from proto files
