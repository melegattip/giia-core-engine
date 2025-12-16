# Task 11: Auth Service Registration Flows - Specification

**Task ID**: task-11-auth-service-registration
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Completion**: 5% remaining work on auth-service
**Dependencies**: Task 5 (95% complete)

---

## Overview

Complete the auth-service by implementing user registration flows including email verification, password reset, and account activation. This task brings auth-service from 95% to 100% completion by adding the remaining authentication use cases.

---

## User Scenarios

### US1: User Registration with Email Verification (P1)

**As a** new user
**I want to** register for an account with email verification
**So that** I can securely access the GIIA platform with a verified email address

**Acceptance Criteria**:
- User submits registration form with email, password, organization details
- System validates email format and password strength
- System creates user account in "pending" status
- System generates verification token and sends email
- User clicks verification link in email
- System activates account and user can log in
- Verification tokens expire after 24 hours
- Multi-tenancy: User is associated with organization

**Success Metrics**:
- 95%+ successful registrations
- <5% bounce rate on verification emails
- <10s p95 for registration endpoint

---

### US2: Password Reset Flow (P1)

**As a** registered user
**I want to** reset my forgotten password
**So that** I can regain access to my account

**Acceptance Criteria**:
- User requests password reset via email
- System validates email exists in database
- System generates password reset token
- System sends password reset email with link
- User clicks reset link and enters new password
- System validates password strength
- System updates password hash
- Reset tokens expire after 1 hour
- Old tokens are invalidated after password change

**Success Metrics**:
- <2s p95 for reset request
- 90%+ successful resets
- Zero security vulnerabilities (token leaks, timing attacks)

---

### US3: Account Activation (P2)

**As an** administrator
**I want to** manually activate/deactivate user accounts
**So that** I can control access to the platform

**Acceptance Criteria**:
- Admin can activate pending user accounts
- Admin can deactivate active user accounts
- Deactivated users cannot log in
- System logs all activation/deactivation events
- Multi-tenancy: Admin can only manage users in their organization

**Success Metrics**:
- <1s p95 for activation operations
- 100% audit trail coverage

---

### US4: REST API Endpoints (P2)

**As a** frontend developer
**I want to** access auth service via REST API
**So that** I can build web and mobile applications

**Acceptance Criteria**:
- POST /api/v1/auth/register - User registration
- POST /api/v1/auth/verify - Email verification
- POST /api/v1/auth/reset-password - Request password reset
- POST /api/v1/auth/confirm-reset - Confirm password reset
- PUT /api/v1/users/:id/activate - Activate user (admin only)
- PUT /api/v1/users/:id/deactivate - Deactivate user (admin only)
- All endpoints return consistent error responses
- OpenAPI/Swagger documentation generated

**Success Metrics**:
- 100% API endpoint coverage
- <100ms p50 response time
- Complete API documentation

---

## Functional Requirements

### FR1: Email Verification System
- Generate cryptographically secure verification tokens (UUID v4)
- Store tokens in database with expiration timestamp
- Send verification emails via SMTP or email service (SendGrid, AWS SES)
- Verify tokens on callback and activate user account
- Handle token expiration gracefully

### FR2: Password Reset System
- Generate secure password reset tokens (UUID v4)
- Store tokens in database with 1-hour expiration
- Send password reset emails with secure links
- Validate new password strength (min 8 chars, uppercase, lowercase, number, special char)
- Hash new password with bcrypt (cost factor 10)
- Invalidate all previous tokens on successful reset

### FR3: Account Activation Management
- Admin-only endpoints protected by RBAC (permission: `users:activate`)
- Activation state: `pending`, `active`, `deactivated`
- Event publishing: `user.activated`, `user.deactivated`
- Audit logging for all state changes

### FR4: Email Service Integration
- Abstract email provider interface for flexibility
- Support SMTP and external providers (SendGrid, AWS SES, Mailgun)
- Email templates for verification and password reset
- Configurable email subjects and sender addresses
- Retry logic for failed email sends

### FR5: REST API Layer
- Chi router for HTTP endpoints
- Middleware: authentication, authorization, logging, error handling
- Request validation with struct tags
- JSON request/response format
- HTTP status codes: 200 (OK), 201 (Created), 400 (Bad Request), 401 (Unauthorized), 404 (Not Found), 500 (Internal Server Error)

---

## Key Entities

### VerificationToken
```go
type VerificationToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Token     string    // UUID v4
    Type      TokenType // "email_verification", "password_reset"
    ExpiresAt time.Time
    UsedAt    *time.Time
    CreatedAt time.Time
}

type TokenType string

const (
    TokenTypeEmailVerification TokenType = "email_verification"
    TokenTypePasswordReset     TokenType = "password_reset"
)
```

### User (Updated)
```go
type User struct {
    // Existing fields...
    Status    UserStatus // "pending", "active", "deactivated"
    VerifiedAt *time.Time
    // ...
}

type UserStatus string

const (
    UserStatusPending     UserStatus = "pending"
    UserStatusActive      UserStatus = "active"
    UserStatusDeactivated UserStatus = "deactivated"
)
```

### EmailMessage
```go
type EmailMessage struct {
    To      string
    Subject string
    Body    string
    HTML    string
    From    string
}
```

---

## Non-Functional Requirements

### Security
- Verification tokens must be cryptographically secure (UUID v4)
- Tokens stored as hashed values in database (SHA-256)
- Password reset links must be single-use
- Rate limiting on registration and reset endpoints (5 requests/minute per IP)
- No sensitive data in email bodies (only tokens)

### Performance
- Registration endpoint: <5s p95 (including email send)
- Email verification: <2s p95
- Password reset request: <2s p95
- Password reset confirmation: <3s p95

### Reliability
- Email service failures should not block user creation
- Failed emails should be retried (3 attempts with exponential backoff)
- Database transactions for user creation and token generation

### Observability
- Log all registration attempts (success and failure)
- Log all email sends (success and failure)
- Log all token verifications (success and failure)
- Metrics: registration_count, verification_count, reset_count, email_failures

---

## Success Criteria

### Mandatory (Must Have)
- ✅ User registration with email verification working end-to-end
- ✅ Password reset flow working end-to-end
- ✅ Account activation/deactivation by admin working
- ✅ All REST API endpoints implemented and documented
- ✅ Email service integration (at least SMTP)
- ✅ Unit tests for all new use cases (80%+ coverage)
- ✅ Integration tests with real database and email service (mock)

### Optional (Nice to Have)
- ⚪ Support for multiple email providers (SendGrid, AWS SES)
- ⚪ Email template system with variables
- ⚪ Resend verification email endpoint
- ⚪ Admin dashboard for user management
- ⚪ Email delivery tracking and analytics

---

## Out of Scope

- ❌ Social login (Google, Facebook, GitHub) - Future task
- ❌ Two-factor authentication (2FA) - Future task
- ❌ Magic link authentication - Future task
- ❌ SMS verification - Future task
- ❌ Password complexity customization per organization - Future task

---

## Dependencies

- **Task 5**: Auth service at 95% (Clean Architecture, RBAC, gRPC, multi-tenancy)
- **External**: Email service provider (SMTP server or SendGrid/AWS SES account)
- **Infrastructure**: Redis for rate limiting (optional, can use in-memory for MVP)

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Email deliverability issues | High | Medium | Use reputable email provider, implement retry logic |
| Token security vulnerabilities | Critical | Low | Use UUID v4, hash tokens in DB, implement expiration |
| Rate limiting bypass | Medium | Medium | Implement IP-based and email-based rate limiting |
| Email service downtime | Medium | Medium | Queue emails for retry, don't block user creation |
| Spam registrations | Medium | High | Implement CAPTCHA or honeypot fields (future) |

---

## References

- [Task 5 Spec](../task-05-auth-service-migration/spec.md) - Auth service foundation
- [Task 6 Spec](../task-06-rbac-implementation/spec.md) - RBAC for admin permissions
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Go Email Libraries](https://github.com/jordan-wright/email) - Email sending in Go

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning
**Next Step**: Create implementation plan (plan.md)