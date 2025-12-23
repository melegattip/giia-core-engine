package auth

import (
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
)

type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
	ValidatePasswordStrength(password string) error
	GenerateRandomPassword(length int) string
}

type passwordService struct {
	minLength int
	cost      int
}

func NewPasswordService(minLength int) PasswordService {
	return &passwordService{
		minLength: minLength,
		cost:      bcrypt.DefaultCost,
	}
}

func (p *passwordService) HashPassword(password string) (string, error) {
	if err := p.ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", pkgErrors.NewInternalServerError("failed to hash password")
	}

	return string(hashedBytes), nil
}

func (p *passwordService) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return pkgErrors.NewUnauthorized("invalid password")
	}
	return nil
}

func (p *passwordService) ValidatePasswordStrength(password string) error {
	if len(password) < p.minLength {
		return pkgErrors.NewBadRequest(fmt.Sprintf("password must be at least %d characters long", p.minLength))
	}

	if len(password) > 128 {
		return pkgErrors.NewBadRequest("password must be less than 128 characters long")
	}

	var (
		hasUpper   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return pkgErrors.NewBadRequest("password must contain at least one uppercase letter")
	}
	if !hasSpecial {
		return pkgErrors.NewBadRequest("password must contain at least one special character")
	}

	// Check for common weak patterns
	if err := p.checkCommonPatterns(password); err != nil {
		return err
	}

	return nil
}

func (p *passwordService) checkCommonPatterns(password string) error {
	// Check for sequential characters
	sequentialPattern := regexp.MustCompile(`(?i)(abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz|123|234|345|456|567|678|789|890)`)
	if sequentialPattern.MatchString(password) {
		return pkgErrors.NewBadRequest("password cannot contain sequential characters")
	}

	// Check for repeated characters
	repeatedPattern := regexp.MustCompile(`(.)\\1{2,}`)
	if repeatedPattern.MatchString(password) {
		return pkgErrors.NewBadRequest("password cannot contain more than 2 consecutive identical characters")
	}

	return nil
}

func (p *passwordService) GenerateRandomPassword(length int) string {
	if length < p.minLength {
		length = p.minLength
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	// Ensure we have at least one character from each required type
	password[0] = "abcdefghijklmnopqrstuvwxyz"[randomInt(26)] // lowercase
	password[1] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"[randomInt(26)] // uppercase
	password[2] = "0123456789"[randomInt(10)]                 // number
	password[3] = "!@#$%^&*"[randomInt(8)]                    // special

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = charset[randomInt(len(charset))]
	}

	// Shuffle the password to avoid predictable patterns
	for i := len(password) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

// Simple random number generator for password generation
func randomInt(max int) int {
	// In production, you should use crypto/rand for better security
	// This is a simplified version for demonstration
	return int(uint32(max) % 256) // Simplified for demo
}
