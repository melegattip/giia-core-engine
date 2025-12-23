package auth

import (
	"crypto/rand"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
)

type TwoFAService interface {
	GenerateSecret(userEmail string) (*TwoFASetup, error)
	GenerateQRCode(secret, userEmail string) ([]byte, error)
	ValidateCode(secret, code string) bool
	GenerateBackupCodes(count int) ([]string, error)
	ValidateBackupCode(codes []string, providedCode string) ([]string, bool)
}

type twoFAService struct {
	issuer string
}

type TwoFASetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

func NewTwoFAService(issuer string) TwoFAService {
	return &twoFAService{
		issuer: issuer,
	}
}

func (t *twoFAService) GenerateSecret(userEmail string) (*TwoFASetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: userEmail,
		SecretSize:  32,
	})
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to generate 2FA secret")
	}

	// Generate backup codes
	backupCodes, err := t.GenerateBackupCodes(8)
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to generate backup codes")
	}

	return &TwoFASetup{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

func (t *twoFAService) GenerateQRCode(secret, userEmail string) ([]byte, error) {
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		t.issuer, userEmail, secret, t.issuer))
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to create 2FA key")
	}

	// Generate QR code image
	img, err := key.Image(256, 256)
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to generate QR code image")
	}

	// Convert image to PNG bytes
	var buf strings.Builder
	if err := png.Encode(&buf, img); err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to encode QR code")
	}

	return []byte(buf.String()), nil
}

func (t *twoFAService) ValidateCode(secret, code string) bool {
	// Remove any spaces or special characters from the code
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")

	return totp.Validate(code, secret)
}

func (t *twoFAService) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		code, err := t.generateRandomCode(8)
		if err != nil {
			return nil, pkgErrors.NewInternalServerError("failed to generate backup code")
		}
		codes[i] = code
	}

	return codes, nil
}

func (t *twoFAService) ValidateBackupCode(codes []string, providedCode string) ([]string, bool) {
	// Clean the provided code
	providedCode = strings.ToUpper(strings.TrimSpace(providedCode))

	for i, code := range codes {
		if strings.ToUpper(code) == providedCode {
			// Remove the used code from the list
			newCodes := make([]string, 0, len(codes)-1)
			newCodes = append(newCodes, codes[:i]...)
			newCodes = append(newCodes, codes[i+1:]...)
			return newCodes, true
		}
	}

	return codes, false
}

func (t *twoFAService) generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", pkgErrors.NewInternalServerError("failed to generate random code")
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	// Format as XXXX-XXXX for better readability
	if length == 8 {
		return fmt.Sprintf("%s-%s", string(bytes[:4]), string(bytes[4:])), nil
	}

	return string(bytes), nil
}

// Additional helper methods
func (t *twoFAService) GetTimeBasedCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func (t *twoFAService) ValidateCodeWithWindow(secret, code string, window int) bool {
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      uint(window),
		Digits:    6,
		Algorithm: otp.AlgorithmSHA1,
	}

	valid, _ := totp.ValidateCustom(code, secret, time.Now(), opts)
	return valid
}
