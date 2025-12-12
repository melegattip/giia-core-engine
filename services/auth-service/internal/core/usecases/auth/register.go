package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type RegisterUseCase struct {
	userRepo  providers.UserRepository
	orgRepo   providers.OrganizationRepository
	tokenRepo providers.TokenRepository
	logger    pkgLogger.Logger
}

func NewRegisterUseCase(
	userRepo providers.UserRepository,
	orgRepo providers.OrganizationRepository,
	tokenRepo providers.TokenRepository,
	logger pkgLogger.Logger,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:  userRepo,
		orgRepo:   orgRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req *domain.RegisterRequest) error {
	if req.Email == "" {
		return pkgErrors.NewBadRequest("email is required")
	}

	if req.Password == "" {
		return pkgErrors.NewBadRequest("password is required")
	}

	if req.FirstName == "" {
		return pkgErrors.NewBadRequest("first name is required")
	}

	if req.LastName == "" {
		return pkgErrors.NewBadRequest("last name is required")
	}

	if req.OrganizationID == "" {
		return pkgErrors.NewBadRequest("organization ID is required")
	}

	if err := validateEmail(req.Email); err != nil {
		return err
	}

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return pkgErrors.NewBadRequest("invalid organization ID format")
	}

	_, err = uc.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgErrors.NewBadRequest("organization not found")
		}
		uc.logger.Error(ctx, err, "Failed to get organization", pkgLogger.Tags{
			"organization_id": orgID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to verify organization")
	}

	existingUser, err := uc.userRepo.GetByEmailAndOrg(ctx, req.Email, orgID)
	if err == nil && existingUser != nil {
		return pkgErrors.NewBadRequest("email already registered in this organization")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to hash password", nil)
		return pkgErrors.NewInternalServerError("failed to hash password")
	}

	user := &domain.User{
		Email:          req.Email,
		Password:       string(hashedPassword),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Phone:          req.Phone,
		Status:         domain.UserStatusInactive,
		OrganizationID: orgID,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.Error(ctx, err, "Failed to create user", pkgLogger.Tags{
			"email":           req.Email,
			"organization_id": orgID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to create user")
	}

	activationToken := uuid.New().String()
	tokenHash := hashActivationToken(activationToken)

	activation := &domain.ActivationToken{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
	}

	if err := uc.tokenRepo.StoreActivationToken(ctx, activation); err != nil {
		uc.logger.Error(ctx, err, "Failed to store activation token", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}

	uc.logger.Info(ctx, "User registered successfully", pkgLogger.Tags{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"organization_id": user.OrganizationID.String(),
	})

	return nil
}

func validateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return pkgErrors.NewBadRequest("invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return pkgErrors.NewBadRequest("password must be at least 8 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	if !hasUpper {
		return pkgErrors.NewBadRequest("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return pkgErrors.NewBadRequest("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return pkgErrors.NewBadRequest("password must contain at least one number")
	}

	if !hasSpecial {
		return pkgErrors.NewBadRequest("password must contain at least one special character")
	}

	return nil
}

func hashActivationToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
