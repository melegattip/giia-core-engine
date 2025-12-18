# Task 11: Auth Service Registration Flows - Implementation Plan

**Task ID**: task-11-auth-service-registration
**Phase**: 2A - Complete to 100%
**Estimated Effort**: 3-5 days
**Dependencies**: Task 5 (95% complete)

---

## Technical Context

### Technology Stack
- **Go Version**: 1.23.4
- **Framework**: Chi router for HTTP
- **Database**: PostgreSQL 16 with GORM
- **Email**: Go standard library `net/smtp` + external providers (SendGrid SDK optional)
- **Token Generation**: `github.com/google/uuid`
- **Password Hashing**: `golang.org/x/crypto/bcrypt`
- **Template Engine**: Go `html/template`
- **Testing**: Go testing, testify/assert, testify/mock

### Existing Architecture
- **Service**: `services/auth-service/`
- **Clean Architecture**: domain, providers, usecases, infrastructure
- **Already Implemented**: Login, JWT tokens, RBAC, gRPC, multi-tenancy
- **Database**: User, Organization, Role, Permission entities exist
- **Shared Packages**: config, logger, database, errors, events

---

## Project Structure

```
services/auth-service/
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── user.go                        # [UPDATE] Add Status, VerifiedAt fields
│   │   │   ├── verification_token.go          # [NEW] Token entity
│   │   │   └── email.go                       # [NEW] Email entity
│   │   │
│   │   ├── providers/
│   │   │   ├── verification_token_repository.go  # [NEW] Token repository interface
│   │   │   └── email_service.go                  # [NEW] Email service interface
│   │   │
│   │   └── usecases/
│   │       ├── auth/
│   │       │   ├── register_user.go           # [NEW] User registration
│   │       │   ├── verify_email.go            # [NEW] Email verification
│   │       │   ├── request_password_reset.go  # [NEW] Request reset
│   │       │   └── confirm_password_reset.go  # [NEW] Confirm reset
│   │       │
│   │       └── user/
│   │           ├── activate_user.go           # [NEW] Admin activation
│   │           └── deactivate_user.go         # [NEW] Admin deactivation
│   │
│   └── infrastructure/
│       ├── repositories/
│       │   └── verification_token_repository.go  # [NEW] GORM implementation
│       │
│       ├── adapters/
│       │   └── email/
│       │       ├── smtp_service.go            # [NEW] SMTP implementation
│       │       ├── sendgrid_service.go        # [NEW] SendGrid implementation (optional)
│       │       └── templates/                 # [NEW] Email templates
│       │           ├── verification.html
│       │           └── password_reset.html
│       │
│       └── entrypoints/
│           └── http/
│               ├── auth_handlers.go           # [NEW] REST endpoints
│               ├── user_handlers.go           # [NEW] User management
│               └── routes.go                  # [UPDATE] Add new routes
│
├── migrations/
│   └── 006_verification_tokens.sql            # [NEW] Token table migration
│
├── .env.example                               # [UPDATE] Add email config
└── README.md                                  # [UPDATE] Document new endpoints
```

---

## Implementation Phases

### Phase 1: Setup (Foundational)

#### T001: Database Migration - Verification Tokens Table [P]
**File**: `services/auth-service/migrations/006_verification_tokens.sql`

```sql
CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('email_verification', 'password_reset')),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_verification_tokens_user_id (user_id),
    INDEX idx_verification_tokens_token_hash (token_hash),
    INDEX idx_verification_tokens_expires_at (expires_at)
);
```

#### T002: Update User Entity with Status Fields
**File**: `services/auth-service/internal/core/domain/user.go`

```go
type UserStatus string

const (
    UserStatusPending     UserStatus = "pending"
    UserStatusActive      UserStatus = "active"
    UserStatusDeactivated UserStatus = "deactivated"
)

type User struct {
    ID             uuid.UUID
    OrganizationID uuid.UUID
    Email          string
    PasswordHash   string
    Status         UserStatus  // [NEW]
    VerifiedAt     *time.Time  // [NEW]
    FirstName      string
    LastName       string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

#### T003: Create Domain Entities
**Files**:
- `services/auth-service/internal/core/domain/verification_token.go`
- `services/auth-service/internal/core/domain/email.go`

```go
// verification_token.go
type VerificationToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Token     string
    TokenHash string
    Type      TokenType
    ExpiresAt time.Time
    UsedAt    *time.Time
    CreatedAt time.Time
}

type TokenType string

const (
    TokenTypeEmailVerification TokenType = "email_verification"
    TokenTypePasswordReset     TokenType = "password_reset"
)

func (t *VerificationToken) IsExpired() bool {
    return time.Now().After(t.ExpiresAt)
}

func (t *VerificationToken) IsUsed() bool {
    return t.UsedAt != nil
}

// email.go
type EmailMessage struct {
    To      string
    Subject string
    Body    string
    HTML    string
    From    string
}

type EmailTemplate struct {
    Subject string
    Body    string
    HTML    string
}
```

#### T004: Create Provider Interfaces
**Files**:
- `services/auth-service/internal/core/providers/verification_token_repository.go`
- `services/auth-service/internal/core/providers/email_service.go`

```go
// verification_token_repository.go
type VerificationTokenRepository interface {
    Create(ctx context.Context, token *domain.VerificationToken) error
    GetByTokenHash(ctx context.Context, tokenHash string) (*domain.VerificationToken, error)
    MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error
    DeleteExpired(ctx context.Context) error
    InvalidateUserTokens(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error
}

// email_service.go
type EmailService interface {
    SendEmail(ctx context.Context, message *domain.EmailMessage) error
    SendVerificationEmail(ctx context.Context, to, token string) error
    SendPasswordResetEmail(ctx context.Context, to, token string) error
}
```

---

### Phase 2: Infrastructure Implementation

#### T005: Implement Verification Token Repository [US1, US2]
**File**: `services/auth-service/internal/infrastructure/repositories/verification_token_repository.go`

```go
type GORMVerificationTokenRepository struct {
    db     *gorm.DB
    logger pkgLogger.Logger
}

func NewGORMVerificationTokenRepository(db *gorm.DB, logger pkgLogger.Logger) *GORMVerificationTokenRepository {
    return &GORMVerificationTokenRepository{db: db, logger: logger}
}

func (r *GORMVerificationTokenRepository) Create(ctx context.Context, token *domain.VerificationToken) error {
    // Hash token before storing
    token.TokenHash = hashToken(token.Token)
    return r.db.WithContext(ctx).Create(token).Error
}

func (r *GORMVerificationTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.VerificationToken, error) {
    var token domain.VerificationToken
    err := r.db.WithContext(ctx).
        Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
        First(&token).Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, pkgErrors.NewResourceNotFound("token not found or expired")
    }
    return &token, err
}

func (r *GORMVerificationTokenRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&domain.VerificationToken{}).
        Where("id = ?", tokenID).
        Update("used_at", now).Error
}

func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

**Tests**: `verification_token_repository_test.go` - Test all CRUD operations

#### T006: Implement SMTP Email Service [US1, US2]
**File**: `services/auth-service/internal/infrastructure/adapters/email/smtp_service.go`

```go
type SMTPEmailService struct {
    host       string
    port       int
    username   string
    password   string
    from       string
    baseURL    string
    logger     pkgLogger.Logger
    templates  map[string]*template.Template
}

func NewSMTPEmailService(cfg *config.Config, logger pkgLogger.Logger) (*SMTPEmailService, error) {
    service := &SMTPEmailService{
        host:     cfg.GetString("email.smtp.host"),
        port:     cfg.GetInt("email.smtp.port"),
        username: cfg.GetString("email.smtp.username"),
        password: cfg.GetString("email.smtp.password"),
        from:     cfg.GetString("email.from"),
        baseURL:  cfg.GetString("app.base_url"),
        logger:   logger,
        templates: make(map[string]*template.Template),
    }

    // Load templates
    if err := service.loadTemplates(); err != nil {
        return nil, err
    }

    return service, nil
}

func (s *SMTPEmailService) SendEmail(ctx context.Context, message *domain.EmailMessage) error {
    auth := smtp.PlainAuth("", s.username, s.password, s.host)

    msg := []byte("To: " + message.To + "\r\n" +
        "From: " + message.From + "\r\n" +
        "Subject: " + message.Subject + "\r\n" +
        "MIME-version: 1.0;\r\n" +
        "Content-Type: text/html; charset=\"UTF-8\";\r\n" +
        "\r\n" +
        message.HTML + "\r\n")

    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    err := smtp.SendMail(addr, auth, s.from, []string{message.To}, msg)

    if err != nil {
        s.logger.Error(ctx, err, "Failed to send email", pkgLogger.Tags{"to": message.To})
        return pkgErrors.NewInternalServerError("failed to send email")
    }

    s.logger.Info(ctx, "Email sent successfully", pkgLogger.Tags{"to": message.To})
    return nil
}

func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, to, token string) error {
    verifyURL := fmt.Sprintf("%s/verify?token=%s", s.baseURL, token)

    var body bytes.Buffer
    if err := s.templates["verification"].Execute(&body, map[string]string{
        "VerifyURL": verifyURL,
    }); err != nil {
        return err
    }

    return s.SendEmail(ctx, &domain.EmailMessage{
        To:      to,
        From:    s.from,
        Subject: "Verify Your Email Address",
        HTML:    body.String(),
    })
}

func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
    resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

    var body bytes.Buffer
    if err := s.templates["password_reset"].Execute(&body, map[string]string{
        "ResetURL": resetURL,
    }); err != nil {
        return err
    }

    return s.SendEmail(ctx, &domain.EmailMessage{
        To:      to,
        From:    s.from,
        Subject: "Reset Your Password",
        HTML:    body.String(),
    })
}

func (s *SMTPEmailService) loadTemplates() error {
    // Load verification template
    verifyTmpl, err := template.ParseFiles("internal/infrastructure/adapters/email/templates/verification.html")
    if err != nil {
        return err
    }
    s.templates["verification"] = verifyTmpl

    // Load password reset template
    resetTmpl, err := template.ParseFiles("internal/infrastructure/adapters/email/templates/password_reset.html")
    if err != nil {
        return err
    }
    s.templates["password_reset"] = resetTmpl

    return nil
}
```

**Tests**: `smtp_service_test.go` - Test email sending with mock SMTP server

#### T007: Create Email Templates
**Files**:
- `services/auth-service/internal/infrastructure/adapters/email/templates/verification.html`
- `services/auth-service/internal/infrastructure/adapters/email/templates/password_reset.html`

```html
<!-- verification.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <h2>Welcome to GIIA Platform!</h2>
    <p>Please verify your email address by clicking the button below:</p>
    <a href="{{.VerifyURL}}" style="display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 4px;">
        Verify Email
    </a>
    <p>Or copy this link: {{.VerifyURL}}</p>
    <p>This link will expire in 24 hours.</p>
    <p>If you didn't create an account, please ignore this email.</p>
</body>
</html>

<!-- password_reset.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <h2>Password Reset Request</h2>
    <p>Click the button below to reset your password:</p>
    <a href="{{.ResetURL}}" style="display: inline-block; padding: 12px 24px; background-color: #2196F3; color: white; text-decoration: none; border-radius: 4px;">
        Reset Password
    </a>
    <p>Or copy this link: {{.ResetURL}}</p>
    <p>This link will expire in 1 hour.</p>
    <p>If you didn't request a password reset, please ignore this email.</p>
</body>
</html>
```

---

### Phase 3: Use Case Implementation

#### T008: Implement Register User Use Case [US1]
**File**: `services/auth-service/internal/core/usecases/auth/register_user.go`

```go
type RegisterUserUseCase struct {
    userRepo      providers.UserRepository
    tokenRepo     providers.VerificationTokenRepository
    emailService  providers.EmailService
    eventPublisher providers.EventPublisher
    logger        pkgLogger.Logger
}

type RegisterUserInput struct {
    Email          string
    Password       string
    FirstName      string
    LastName       string
    OrganizationID uuid.UUID
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input *RegisterUserInput) (*domain.User, error) {
    // Validate input
    if err := validateEmail(input.Email); err != nil {
        return nil, pkgErrors.NewBadRequest("invalid email format")
    }
    if err := validatePassword(input.Password); err != nil {
        return nil, pkgErrors.NewBadRequest("password too weak")
    }

    // Check if user exists
    existing, _ := uc.userRepo.GetByEmail(ctx, input.Email, input.OrganizationID)
    if existing != nil {
        return nil, pkgErrors.NewBadRequest("email already registered")
    }

    // Hash password
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, pkgErrors.NewInternalServerError("failed to hash password")
    }

    // Create user in "pending" status
    user := &domain.User{
        ID:             uuid.New(),
        OrganizationID: input.OrganizationID,
        Email:          input.Email,
        PasswordHash:   string(passwordHash),
        Status:         domain.UserStatusPending,
        FirstName:      input.FirstName,
        LastName:       input.LastName,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    // Generate verification token
    token := &domain.VerificationToken{
        ID:        uuid.New(),
        UserID:    user.ID,
        Token:     uuid.New().String(),
        Type:      domain.TokenTypeEmailVerification,
        ExpiresAt: time.Now().Add(24 * time.Hour),
        CreatedAt: time.Now(),
    }

    if err := uc.tokenRepo.Create(ctx, token); err != nil {
        // Rollback user creation (or mark for cleanup)
        uc.logger.Error(ctx, err, "Failed to create verification token", pkgLogger.Tags{"user_id": user.ID.String()})
        return nil, pkgErrors.NewInternalServerError("failed to create verification token")
    }

    // Send verification email (async, don't block on failure)
    go func() {
        if err := uc.emailService.SendVerificationEmail(context.Background(), user.Email, token.Token); err != nil {
            uc.logger.Error(context.Background(), err, "Failed to send verification email", pkgLogger.Tags{
                "user_id": user.ID.String(),
                "email":   user.Email,
            })
        }
    }()

    // Publish event
    uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
        Type:    "user.registered",
        Subject: fmt.Sprintf("users/%s", user.ID.String()),
        Data: map[string]interface{}{
            "user_id": user.ID.String(),
            "email":   user.Email,
        },
    })

    uc.logger.Info(ctx, "User registered", pkgLogger.Tags{"user_id": user.ID.String()})
    return user, nil
}

func validatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }
    // Add more validation (uppercase, lowercase, number, special char)
    return nil
}
```

**Tests**: `register_user_test.go` - Test validation, user creation, token generation, email sending

#### T009: Implement Verify Email Use Case [US1]
**File**: `services/auth-service/internal/core/usecases/auth/verify_email.go`

```go
type VerifyEmailUseCase struct {
    userRepo      providers.UserRepository
    tokenRepo     providers.VerificationTokenRepository
    eventPublisher providers.EventPublisher
    logger        pkgLogger.Logger
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, token string) error {
    // Hash token for lookup
    tokenHash := hashToken(token)

    // Get token from database
    verificationToken, err := uc.tokenRepo.GetByTokenHash(ctx, tokenHash)
    if err != nil {
        return pkgErrors.NewBadRequest("invalid or expired token")
    }

    if verificationToken.IsExpired() {
        return pkgErrors.NewBadRequest("token has expired")
    }

    if verificationToken.IsUsed() {
        return pkgErrors.NewBadRequest("token already used")
    }

    // Get user
    user, err := uc.userRepo.GetByID(ctx, verificationToken.UserID)
    if err != nil {
        return err
    }

    // Update user status and verified_at
    user.Status = domain.UserStatusActive
    now := time.Now()
    user.VerifiedAt = &now
    user.UpdatedAt = now

    if err := uc.userRepo.Update(ctx, user); err != nil {
        return err
    }

    // Mark token as used
    if err := uc.tokenRepo.MarkAsUsed(ctx, verificationToken.ID); err != nil {
        uc.logger.Error(ctx, err, "Failed to mark token as used", pkgLogger.Tags{"token_id": verificationToken.ID.String()})
    }

    // Publish event
    uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
        Type:    "user.verified",
        Subject: fmt.Sprintf("users/%s", user.ID.String()),
        Data: map[string]interface{}{
            "user_id": user.ID.String(),
            "email":   user.Email,
        },
    })

    uc.logger.Info(ctx, "Email verified", pkgLogger.Tags{"user_id": user.ID.String()})
    return nil
}
```

**Tests**: `verify_email_test.go` - Test happy path, expired token, used token, invalid token

#### T010: Implement Request Password Reset Use Case [US2]
**File**: `services/auth-service/internal/core/usecases/auth/request_password_reset.go`

```go
type RequestPasswordResetUseCase struct {
    userRepo     providers.UserRepository
    tokenRepo    providers.VerificationTokenRepository
    emailService providers.EmailService
    logger       pkgLogger.Logger
}

func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, email string, organizationID uuid.UUID) error {
    // Get user by email
    user, err := uc.userRepo.GetByEmail(ctx, email, organizationID)
    if err != nil {
        // Don't reveal if user exists for security
        uc.logger.Info(ctx, "Password reset requested for non-existent email", pkgLogger.Tags{"email": email})
        return nil
    }

    // Invalidate existing password reset tokens
    if err := uc.tokenRepo.InvalidateUserTokens(ctx, user.ID, domain.TokenTypePasswordReset); err != nil {
        uc.logger.Error(ctx, err, "Failed to invalidate tokens", pkgLogger.Tags{"user_id": user.ID.String()})
    }

    // Generate password reset token
    token := &domain.VerificationToken{
        ID:        uuid.New(),
        UserID:    user.ID,
        Token:     uuid.New().String(),
        Type:      domain.TokenTypePasswordReset,
        ExpiresAt: time.Now().Add(1 * time.Hour),
        CreatedAt: time.Now(),
    }

    if err := uc.tokenRepo.Create(ctx, token); err != nil {
        return pkgErrors.NewInternalServerError("failed to create reset token")
    }

    // Send password reset email (async)
    go func() {
        if err := uc.emailService.SendPasswordResetEmail(context.Background(), user.Email, token.Token); err != nil {
            uc.logger.Error(context.Background(), err, "Failed to send reset email", pkgLogger.Tags{
                "user_id": user.ID.String(),
            })
        }
    }()

    uc.logger.Info(ctx, "Password reset requested", pkgLogger.Tags{"user_id": user.ID.String()})
    return nil
}
```

**Tests**: `request_password_reset_test.go`

#### T011: Implement Confirm Password Reset Use Case [US2]
**File**: `services/auth-service/internal/core/usecases/auth/confirm_password_reset.go`

```go
type ConfirmPasswordResetUseCase struct {
    userRepo  providers.UserRepository
    tokenRepo providers.VerificationTokenRepository
    logger    pkgLogger.Logger
}

func (uc *ConfirmPasswordResetUseCase) Execute(ctx context.Context, token, newPassword string) error {
    // Validate password
    if err := validatePassword(newPassword); err != nil {
        return pkgErrors.NewBadRequest("password too weak")
    }

    // Hash token for lookup
    tokenHash := hashToken(token)

    // Get token
    resetToken, err := uc.tokenRepo.GetByTokenHash(ctx, tokenHash)
    if err != nil {
        return pkgErrors.NewBadRequest("invalid or expired token")
    }

    if resetToken.IsExpired() {
        return pkgErrors.NewBadRequest("token has expired")
    }

    if resetToken.IsUsed() {
        return pkgErrors.NewBadRequest("token already used")
    }

    // Get user
    user, err := uc.userRepo.GetByID(ctx, resetToken.UserID)
    if err != nil {
        return err
    }

    // Hash new password
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return pkgErrors.NewInternalServerError("failed to hash password")
    }

    // Update password
    user.PasswordHash = string(passwordHash)
    user.UpdatedAt = time.Now()

    if err := uc.userRepo.Update(ctx, user); err != nil {
        return err
    }

    // Mark token as used
    if err := uc.tokenRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
        uc.logger.Error(ctx, err, "Failed to mark token as used", pkgLogger.Tags{"token_id": resetToken.ID.String()})
    }

    // Invalidate all other password reset tokens for this user
    if err := uc.tokenRepo.InvalidateUserTokens(ctx, user.ID, domain.TokenTypePasswordReset); err != nil {
        uc.logger.Error(ctx, err, "Failed to invalidate tokens", pkgLogger.Tags{"user_id": user.ID.String()})
    }

    uc.logger.Info(ctx, "Password reset completed", pkgLogger.Tags{"user_id": user.ID.String()})
    return nil
}
```

**Tests**: `confirm_password_reset_test.go`

#### T012: Implement Activate/Deactivate User Use Cases [US3]
**Files**:
- `services/auth-service/internal/core/usecases/user/activate_user.go`
- `services/auth-service/internal/core/usecases/user/deactivate_user.go`

```go
// activate_user.go
type ActivateUserUseCase struct {
    userRepo       providers.UserRepository
    permissionCheck providers.PermissionChecker // RBAC check
    eventPublisher providers.EventPublisher
    logger         pkgLogger.Logger
}

func (uc *ActivateUserUseCase) Execute(ctx context.Context, adminUserID, targetUserID uuid.UUID) error {
    // Check admin permission
    hasPermission, err := uc.permissionCheck.CheckPermission(ctx, adminUserID, "users:activate")
    if err != nil || !hasPermission {
        return pkgErrors.NewUnauthorizedRequest("insufficient permissions")
    }

    // Get target user
    user, err := uc.userRepo.GetByID(ctx, targetUserID)
    if err != nil {
        return err
    }

    if user.Status == domain.UserStatusActive {
        return pkgErrors.NewBadRequest("user is already active")
    }

    // Activate user
    user.Status = domain.UserStatusActive
    user.UpdatedAt = time.Now()

    if err := uc.userRepo.Update(ctx, user); err != nil {
        return err
    }

    // Publish event
    uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
        Type:    "user.activated",
        Subject: fmt.Sprintf("users/%s", user.ID.String()),
        Data: map[string]interface{}{
            "user_id":     user.ID.String(),
            "activated_by": adminUserID.String(),
        },
    })

    uc.logger.Info(ctx, "User activated", pkgLogger.Tags{
        "user_id":     user.ID.String(),
        "activated_by": adminUserID.String(),
    })
    return nil
}

// deactivate_user.go - Similar structure
```

**Tests**: `activate_user_test.go`, `deactivate_user_test.go`

---

### Phase 4: REST API Endpoints [US4]

#### T013: Create HTTP Handlers
**File**: `services/auth-service/internal/infrastructure/entrypoints/http/auth_handlers.go`

```go
type AuthHandlers struct {
    registerUseCase            *usecases.RegisterUserUseCase
    verifyEmailUseCase         *usecases.VerifyEmailUseCase
    requestPasswordResetUseCase *usecases.RequestPasswordResetUseCase
    confirmPasswordResetUseCase *usecases.ConfirmPasswordResetUseCase
    logger                     pkgLogger.Logger
}

// POST /api/v1/auth/register
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email          string `json:"email" validate:"required,email"`
        Password       string `json:"password" validate:"required,min=8"`
        FirstName      string `json:"first_name" validate:"required"`
        LastName       string `json:"last_name" validate:"required"`
        OrganizationID string `json:"organization_id" validate:"required,uuid"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, pkgErrors.NewBadRequest("invalid request body"))
        return
    }

    // Validate
    if err := validate.Struct(req); err != nil {
        respondError(w, pkgErrors.NewBadRequest(err.Error()))
        return
    }

    orgID, _ := uuid.Parse(req.OrganizationID)

    user, err := h.registerUseCase.Execute(r.Context(), &usecases.RegisterUserInput{
        Email:          req.Email,
        Password:       req.Password,
        FirstName:      req.FirstName,
        LastName:       req.LastName,
        OrganizationID: orgID,
    })

    if err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusCreated, map[string]interface{}{
        "id":     user.ID.String(),
        "email":  user.Email,
        "status": user.Status,
        "message": "Registration successful. Please check your email to verify your account.",
    })
}

// POST /api/v1/auth/verify
func (h *AuthHandlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Token string `json:"token" validate:"required"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, pkgErrors.NewBadRequest("invalid request body"))
        return
    }

    if err := h.verifyEmailUseCase.Execute(r.Context(), req.Token); err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Email verified successfully. You can now log in.",
    })
}

// POST /api/v1/auth/reset-password
func (h *AuthHandlers) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email          string `json:"email" validate:"required,email"`
        OrganizationID string `json:"organization_id" validate:"required,uuid"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, pkgErrors.NewBadRequest("invalid request body"))
        return
    }

    orgID, _ := uuid.Parse(req.OrganizationID)

    if err := h.requestPasswordResetUseCase.Execute(r.Context(), req.Email, orgID); err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "If the email exists, a password reset link has been sent.",
    })
}

// POST /api/v1/auth/confirm-reset
func (h *AuthHandlers) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Token       string `json:"token" validate:"required"`
        NewPassword string `json:"new_password" validate:"required,min=8"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, pkgErrors.NewBadRequest("invalid request body"))
        return
    }

    if err := h.confirmPasswordResetUseCase.Execute(r.Context(), req.Token, req.NewPassword); err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Password reset successfully. You can now log in with your new password.",
    })
}
```

#### T014: Create User Management Handlers [US3]
**File**: `services/auth-service/internal/infrastructure/entrypoints/http/user_handlers.go`

```go
type UserHandlers struct {
    activateUseCase   *usecases.ActivateUserUseCase
    deactivateUseCase *usecases.DeactivateUserUseCase
    authMiddleware    *AuthMiddleware
    logger            pkgLogger.Logger
}

// PUT /api/v1/users/:id/activate
func (h *UserHandlers) ActivateUser(w http.ResponseWriter, r *http.Request) {
    // Get admin user ID from JWT (set by auth middleware)
    adminUserID := r.Context().Value("user_id").(uuid.UUID)

    // Get target user ID from URL
    targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        respondError(w, pkgErrors.NewBadRequest("invalid user ID"))
        return
    }

    if err := h.activateUseCase.Execute(r.Context(), adminUserID, targetUserID); err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "User activated successfully",
    })
}

// PUT /api/v1/users/:id/deactivate
func (h *UserHandlers) DeactivateUser(w http.ResponseWriter, r *http.Request) {
    // Similar to ActivateUser
}
```

#### T015: Update Routes [US4]
**File**: `services/auth-service/internal/infrastructure/entrypoints/http/routes.go`

```go
func SetupRoutes(r chi.Router, handlers *Handlers, authMiddleware *AuthMiddleware) {
    // Existing routes...

    // [NEW] Public auth endpoints
    r.Post("/api/v1/auth/register", handlers.Auth.Register)
    r.Post("/api/v1/auth/verify", handlers.Auth.VerifyEmail)
    r.Post("/api/v1/auth/reset-password", handlers.Auth.RequestPasswordReset)
    r.Post("/api/v1/auth/confirm-reset", handlers.Auth.ConfirmPasswordReset)

    // [NEW] Protected user management endpoints
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        r.Use(authMiddleware.RequirePermission("users:activate"))

        r.Put("/api/v1/users/{id}/activate", handlers.User.ActivateUser)
        r.Put("/api/v1/users/{id}/deactivate", handlers.User.DeactivateUser)
    })
}
```

---

### Phase 5: Configuration and Documentation

#### T016: Update Configuration
**File**: `services/auth-service/.env.example`

Add:
```env
# Email Configuration
EMAIL_PROVIDER=smtp
EMAIL_SMTP_HOST=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USERNAME=your-email@gmail.com
EMAIL_SMTP_PASSWORD=your-app-password
EMAIL_FROM=noreply@giia.com

# Application
APP_BASE_URL=http://localhost:8083
```

#### T017: Update README with API Documentation
**File**: `services/auth-service/README.md`

Add section:
```markdown
## Registration and Password Reset API

### Register New User
```bash
curl -X POST http://localhost:8083/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### Verify Email
```bash
curl -X POST http://localhost:8083/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "token": "verification-token-from-email"
  }'
```

### Request Password Reset
```bash
curl -X POST http://localhost:8083/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### Confirm Password Reset
```bash
curl -X POST http://localhost:8083/api/v1/auth/confirm-reset \
  -H "Content-Type: application/json" \
  -d '{
    "token": "reset-token-from-email",
    "new_password": "NewSecurePass123!"
  }'
```
```

---

## Testing Strategy

### Unit Tests (80%+ Coverage Target)

1. **Use Cases** (`*_test.go` for each use case):
   - Test happy paths
   - Test validation errors
   - Test edge cases (expired tokens, used tokens, etc.)
   - Mock repositories and email service

2. **Repositories** (`*_repository_test.go`):
   - Test CRUD operations
   - Test queries with filters
   - Use in-memory SQLite for tests

3. **Email Service** (`*_service_test.go`):
   - Test email sending with mock SMTP
   - Test template rendering

### Integration Tests

1. **Database Integration**:
   - Test repositories with real PostgreSQL (Docker container)
   - Test transactions and rollbacks

2. **Email Integration**:
   - Test with Mailhog or similar SMTP test server
   - Verify email content and headers

3. **End-to-End Flow**:
   - Register user → verify email → login
   - Request reset → confirm reset → login with new password

---

## Dependencies and Execution Order

### Critical Path
1. **Phase 1** (Setup) must complete first - database migration, domain entities
2. **Phase 2** (Infrastructure) can start after Phase 1
3. **Phase 3** (Use Cases) depends on Phase 2
4. **Phase 4** (REST API) depends on Phase 3
5. **Phase 5** (Docs) can run in parallel with Phase 4

### External Dependencies
- SMTP server or email service account (SendGrid, AWS SES)
- Email templates must be designed before implementation

---

## Acceptance Checklist

Before marking task as complete:
- [ ] All database migrations applied successfully
- [ ] All unit tests pass (80%+ coverage)
- [ ] Integration tests pass
- [ ] Email sending works (at least with SMTP)
- [ ] All REST endpoints return correct status codes
- [ ] API documentation updated in README
- [ ] .env.example updated with email config
- [ ] Registration flow works end-to-end
- [ ] Password reset flow works end-to-end
- [ ] Account activation works for admins
- [ ] All NATS events published correctly
- [ ] No security vulnerabilities (token leaks, timing attacks)
- [ ] Performance metrics meet targets (<5s registration p95)

---

## Rollback Plan

If issues arise:
1. Revert database migration (`006_verification_tokens.sql`)
2. Remove new HTTP routes
3. Revert User entity changes
4. Remove email service configuration

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
**Estimated Duration**: 3-5 days