# Feature Specification: Auth/IAM Service Migration with Multi-Tenancy

**Created**: 2025-12-09

## User Scenarios & Testing *(mandatory)*

### User Story 1 - User Authentication (Priority: P1)

As a user, I need to authenticate with email/password so that I can access the GIIA platform securely.

**Why this priority**: Absolutely critical. No user can access the system without authentication. Blocks all other features.

**Independent Test**: Can be fully tested by calling POST /api/v1/auth/login with valid credentials and receiving a JWT token. Can then use token to access protected endpoints. Delivers standalone value: users can log in and access the system.

**Acceptance Scenarios**:

1. **Scenario**: Successful login
   - **Given** a user exists with email "user@example.com" and correct password
   - **When** user submits login request
   - **Then** system returns JWT access token and refresh token

2. **Scenario**: Failed login with invalid credentials
   - **Given** a user submits incorrect password
   - **When** login request is processed
   - **Then** system returns 401 Unauthorized with error message

3. **Scenario**: Token validation
   - **Given** a user has valid JWT access token
   - **When** user makes request to protected endpoint
   - **Then** system validates token and allows access

---

### User Story 2 - Multi-Tenant Organization Isolation (Priority: P1)

As a system administrator, I need data isolation between tenant organizations so that Company A cannot access Company B's data.

**Why this priority**: Critical security requirement for SaaS platform. Without tenant isolation, platform cannot be used by multiple customers. Legal and compliance requirement.

**Independent Test**: Can be tested by creating two organizations, creating users in each, and verifying that User A from Org 1 cannot access any data from Org 2 via API calls. Delivers standalone value: secure multi-tenancy.

**Acceptance Scenarios**:

1. **Scenario**: Tenant data isolation
   - **Given** User A belongs to Organization 1 and User B belongs to Organization 2
   - **When** User A attempts to access Organization 2's resources
   - **Then** system returns 403 Forbidden

2. **Scenario**: Tenant context in token
   - **Given** a user logs in
   - **When** JWT token is generated
   - **Then** token includes organization_id (tenant_id) in claims

3. **Scenario**: Automatic tenant filtering
   - **Given** a user makes request with valid token containing tenant_id
   - **When** service queries database
   - **Then** all queries automatically filter by tenant_id from token

---

### User Story 3 - User Registration and Activation (Priority: P2)

As a new user, I need to register an account and activate it via email so that I can start using the GIIA platform.

**Why this priority**: Important for self-service onboarding but not blocking for initial pilot with manually created users. Can launch MVP with admin-created accounts.

**Independent Test**: Can be tested by calling POST /api/v1/auth/register, receiving confirmation email, and activating account via token link. Then logging in successfully.

**Acceptance Scenarios**:

1. **Scenario**: User registration
   - **Given** a new user provides email, password, and organization details
   - **When** user submits registration request
   - **Then** system creates inactive user account and sends activation email

2. **Scenario**: Email activation
   - **Given** user received activation email with token
   - **When** user clicks activation link
   - **Then** system activates account and allows login

3. **Scenario**: Duplicate email prevention
   - **Given** email "user@example.com" already exists
   - **When** another user tries to register with same email
   - **Then** system returns 409 Conflict error

---

### User Story 4 - Password Reset (Priority: P3)

As a user who forgot my password, I need to reset it via email so that I can regain access to my account.

**Why this priority**: Nice-to-have for production but not blocking for MVP. Admins can manually reset passwords initially.

**Independent Test**: Can be tested by requesting password reset, receiving email with reset token, and successfully setting new password.

**Acceptance Scenarios**:

1. **Scenario**: Password reset request
   - **Given** user exists with email "user@example.com"
   - **When** user requests password reset
   - **Then** system sends reset email with time-limited token

2. **Scenario**: Password reset completion
   - **Given** user has valid reset token
   - **When** user submits new password
   - **Then** system updates password and invalidates reset token

3. **Scenario**: Expired reset token
   - **Given** reset token was created more than 1 hour ago
   - **When** user attempts to use token
   - **Then** system returns error and requires new reset request

---

### Edge Cases

- What happens when JWT token expires during active session?
- How to handle user switching between organizations (multi-tenant users)?
- What happens if email service is unavailable during registration?
- How to migrate existing users from legacy users-service?
- How to handle password complexity requirements?
- What happens when refresh token is used after logout?
- How to handle concurrent login attempts from same user?
- What happens when organization is deactivated but users still have valid tokens?

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication
- **FR-001**: System MUST support email/password authentication
- **FR-002**: System MUST generate JWT access tokens (15-minute expiry) and refresh tokens (7-day expiry)
- **FR-003**: System MUST hash passwords using bcrypt with cost factor 12
- **FR-004**: System MUST validate password complexity (min 8 characters, uppercase, lowercase, number, special char)
- **FR-005**: System MUST implement rate limiting for login attempts (5 attempts per 15 minutes per IP)
- **FR-006**: System MUST support token refresh without requiring re-authentication
- **FR-007**: System MUST implement logout functionality that invalidates tokens

#### Multi-Tenancy
- **FR-008**: System MUST associate every user with exactly one organization (tenant)
- **FR-009**: System MUST include organization_id in JWT token claims
- **FR-010**: System MUST automatically filter all database queries by organization_id from token
- **FR-011**: System MUST prevent users from accessing resources belonging to other organizations
- **FR-012**: System MUST support organization-level configuration (branding, settings)

#### User Management
- **FR-013**: System MUST support user registration with email verification
- **FR-014**: System MUST send activation emails with time-limited tokens (24 hours)
- **FR-015**: System MUST prevent duplicate user emails across all organizations
- **FR-016**: System MUST support user profile management (name, email, password)
- **FR-017**: System MUST support password reset flow with email tokens (1 hour expiry)
- **FR-018**: System MUST track user login history and last login timestamp

#### Migration
- **FR-019**: System MUST migrate existing users from legacy users-service database
- **FR-020**: System MUST update all import paths from "users-service" to "auth-service"
- **FR-021**: System MUST maintain backward compatibility with existing API contracts during migration
- **FR-022**: System MUST migrate existing user passwords securely (no plaintext exposure)

### Key Entities

- **User**: Email, hashed_password, name, status (active/inactive), organization_id, created_at, last_login_at
- **Organization**: Name, slug, status (active/suspended), settings (JSON), created_at
- **RefreshToken**: Token hash, user_id, expires_at, revoked (boolean)
- **PasswordResetToken**: Token hash, user_id, expires_at, used (boolean)
- **ActivationToken**: Token hash, user_id, expires_at, used (boolean)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully log in and receive JWT tokens in under 500ms (p95)
- **SC-002**: Token validation completes in under 50ms (p95)
- **SC-003**: Multi-tenant isolation is verified with zero cross-tenant data leaks in security audit
- **SC-004**: 100% of legacy users are migrated successfully without data loss
- **SC-005**: All API endpoints enforce organization_id filtering automatically
- **SC-006**: Password reset flow completes successfully in under 5 minutes end-to-end
- **SC-007**: Service handles 1000 concurrent authentication requests without degradation
- **SC-008**: All import paths are updated and CI/CD builds pass without errors
