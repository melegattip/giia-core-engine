package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/giia/giia-core-engine/services/auth-service/internal/domain"
	"github.com/giia/giia-core-engine/services/auth-service/pkg/database"

	"github.com/lib/pq"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error

	GetPreferences(ctx context.Context, userID uint) (*domain.Preferences, error)
	UpdatePreferences(ctx context.Context, prefs *domain.Preferences) error

	GetNotifications(ctx context.Context, userID uint) (*domain.NotificationSettings, error)
	UpdateNotifications(ctx context.Context, notif *domain.NotificationSettings) error

	Get2FA(ctx context.Context, userID uint) (*domain.TwoFA, error)
	Update2FA(ctx context.Context, twofa *domain.TwoFA) error

	// Additional methods for user management
	UpdatePassword(ctx context.Context, userID uint, passwordHash string) error
	SetEmailVerified(ctx context.Context, userID uint, verified bool) error
	SetEmailVerificationToken(ctx context.Context, userID uint, token string, expires time.Time) error
	SetPasswordResetToken(ctx context.Context, userID uint, token string, expires time.Time) error
	GetByEmailVerificationToken(ctx context.Context, token string) (*domain.User, error)
	GetByPasswordResetToken(ctx context.Context, token string) (*domain.User, error)
	IncrementFailedLoginAttempts(ctx context.Context, userID uint) error
	ResetFailedLoginAttempts(ctx context.Context, userID uint) error
	SetAccountLocked(ctx context.Context, userID uint, lockedUntil time.Time) error
	UpdateLastLogin(ctx context.Context, userID uint) error
}

type userRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, avatar, is_active, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		user.Email, user.Password, user.FirstName, user.LastName,
		user.Phone, user.Avatar, user.IsActive, user.IsVerified,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Create default preferences
	if err := r.createDefaultPreferences(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to create default preferences: %w", err)
	}

	// Create default notification settings
	if err := r.createDefaultNotificationSettings(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to create default notification settings: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, avatar, is_active, is_verified,
			   email_verification_token, email_verification_expires, password_reset_token,
			   password_reset_expires, last_login, failed_login_attempts, locked_until,
			   created_at, updated_at
		FROM users WHERE id = $1`

	user := &domain.User{}
	var emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil sql.NullTime
	var emailVerificationToken, passwordResetToken sql.NullString
	var avatar sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.Phone, &avatar, &user.IsActive, &user.IsVerified,
		&emailVerificationToken, &emailVerificationExpires,
		&passwordResetToken, &passwordResetExpires,
		&lastLogin, &user.FailedLoginAttempts, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Set avatar if not null
	if avatar.Valid {
		user.Avatar = avatar.String
	}

	r.setNullTimes(user, emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil)

	if emailVerificationToken.Valid {
		user.EmailVerificationToken = emailVerificationToken.String
	}
	if emailVerificationExpires.Valid {
		user.EmailVerificationExpires = &emailVerificationExpires.Time
	}
	if passwordResetToken.Valid {
		user.PasswordResetToken = passwordResetToken.String
	}
	if passwordResetExpires.Valid {
		user.PasswordResetExpires = &passwordResetExpires.Time
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, avatar, is_active, is_verified,
			   email_verification_token, email_verification_expires, password_reset_token,
			   password_reset_expires, last_login, failed_login_attempts, locked_until,
			   created_at, updated_at
		FROM users WHERE email = $1`

	user := &domain.User{}
	var emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil sql.NullTime
	var emailVerificationToken, passwordResetToken sql.NullString
	var avatar sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.Phone, &avatar, &user.IsActive, &user.IsVerified,
		&emailVerificationToken, &emailVerificationExpires,
		&passwordResetToken, &passwordResetExpires,
		&lastLogin, &user.FailedLoginAttempts, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Set avatar if not null
	if avatar.Valid {
		user.Avatar = avatar.String
	}

	// Assign token values
	if emailVerificationToken.Valid {
		user.EmailVerificationToken = emailVerificationToken.String
	}
	if passwordResetToken.Valid {
		user.PasswordResetToken = passwordResetToken.String
	}

	r.setNullTimes(user, emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil)
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, phone = $3, avatar = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6`

	result, err := r.db.ExecContext(ctx, query,
		user.FirstName, user.LastName, user.Phone, user.Avatar, user.IsActive, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", user.ID)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}

func (r *userRepository) GetPreferences(ctx context.Context, userID uint) (*domain.Preferences, error) {
	query := `
		SELECT user_id, currency, language, theme, date_format, timezone
		FROM user_preferences WHERE user_id = $1`

	prefs := &domain.Preferences{}
	var timezone string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.Currency, &prefs.Language,
		&prefs.Theme, &prefs.DateFormat, &timezone,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("preferences for user %d not found", userID)
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return prefs, nil
}

func (r *userRepository) UpdatePreferences(ctx context.Context, prefs *domain.Preferences) error {
	query := `
		UPDATE user_preferences 
		SET currency = $1, language = $2, theme = $3, date_format = $4, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $5`

	result, err := r.db.ExecContext(ctx, query,
		prefs.Currency, prefs.Language, prefs.Theme, prefs.DateFormat, prefs.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("preferences for user %d not found", prefs.UserID)
	}

	return nil
}

func (r *userRepository) GetNotifications(ctx context.Context, userID uint) (*domain.NotificationSettings, error) {
	query := `
		SELECT user_id, email_notifications, push_notifications, weekly_reports, 
			   expense_alerts, budget_alerts, achievement_notifications
		FROM user_notification_settings WHERE user_id = $1`

	notif := &domain.NotificationSettings{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&notif.UserID, &notif.EmailNotifications, &notif.PushNotifications,
		&notif.WeeklyReports, &notif.ExpenseAlerts, &notif.BudgetAlerts,
		&notif.AchievementNotifications,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification settings for user %d not found", userID)
		}
		return nil, fmt.Errorf("failed to get notification settings: %w", err)
	}

	return notif, nil
}

func (r *userRepository) UpdateNotifications(ctx context.Context, notif *domain.NotificationSettings) error {
	query := `
		UPDATE user_notification_settings 
		SET email_notifications = $1, push_notifications = $2, weekly_reports = $3,
			expense_alerts = $4, budget_alerts = $5, achievement_notifications = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $7`

	result, err := r.db.ExecContext(ctx, query,
		notif.EmailNotifications, notif.PushNotifications, notif.WeeklyReports,
		notif.ExpenseAlerts, notif.BudgetAlerts, notif.AchievementNotifications,
		notif.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update notification settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("notification settings for user %d not found", notif.UserID)
	}

	return nil
}

func (r *userRepository) Get2FA(ctx context.Context, userID uint) (*domain.TwoFA, error) {
	query := `
		SELECT user_id, secret, enabled, backup_codes, last_used_code
		FROM user_two_fa WHERE user_id = $1`

	twofa := &domain.TwoFA{}
	var backupCodes pq.StringArray
	var lastUsedCode sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&twofa.UserID, &twofa.Secret, &twofa.Enabled, &backupCodes, &lastUsedCode,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("2FA settings for user %d not found", userID)
		}
		return nil, fmt.Errorf("failed to get 2FA settings: %w", err)
	}

	twofa.BackupCodes = []string(backupCodes)
	if lastUsedCode.Valid {
		twofa.LastUsedCode = lastUsedCode.String
	}

	return twofa, nil
}

func (r *userRepository) Update2FA(ctx context.Context, twofa *domain.TwoFA) error {
	query := `
		INSERT INTO user_two_fa (user_id, secret, enabled, backup_codes, last_used_code)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			secret = EXCLUDED.secret,
			enabled = EXCLUDED.enabled,
			backup_codes = EXCLUDED.backup_codes,
			last_used_code = EXCLUDED.last_used_code,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.ExecContext(ctx, query,
		twofa.UserID, twofa.Secret, twofa.Enabled,
		pq.Array(twofa.BackupCodes), nullString(twofa.LastUsedCode),
	)
	if err != nil {
		return fmt.Errorf("failed to update 2FA settings: %w", err)
	}

	return nil
}

// Additional helper methods implementations continue below...

func (r *userRepository) UpdatePassword(ctx context.Context, userID uint, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) SetEmailVerified(ctx context.Context, userID uint, verified bool) error {
	query := `UPDATE users SET is_verified = $1, email_verification_token = NULL, email_verification_expires = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, verified, userID)
	if err != nil {
		return fmt.Errorf("failed to set email verified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) SetEmailVerificationToken(ctx context.Context, userID uint, token string, expires time.Time) error {
	query := `UPDATE users SET email_verification_token = $1, email_verification_expires = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, token, expires, userID)
	if err != nil {
		return fmt.Errorf("failed to set email verification token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) SetPasswordResetToken(ctx context.Context, userID uint, token string, expires time.Time) error {
	query := `UPDATE users SET password_reset_token = $1, password_reset_expires = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, token, expires, userID)
	if err != nil {
		return fmt.Errorf("failed to set password reset token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) GetByEmailVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, is_active, is_verified,
			   email_verification_token, email_verification_expires, password_reset_token,
			   password_reset_expires, last_login, failed_login_attempts, locked_until,
			   created_at, updated_at
		FROM users 
		WHERE email_verification_token = $1 AND email_verification_expires > CURRENT_TIMESTAMP`

	user := &domain.User{}
	var emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil sql.NullTime

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.Phone, &user.IsActive, &user.IsVerified,
		&user.EmailVerificationToken, &emailVerificationExpires,
		&user.PasswordResetToken, &passwordResetExpires,
		&lastLogin, &user.FailedLoginAttempts, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired email verification token")
		}
		return nil, fmt.Errorf("failed to get user by email verification token: %w", err)
	}

	r.setNullTimes(user, emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil)
	return user, nil
}

func (r *userRepository) GetByPasswordResetToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, is_active, is_verified,
			   email_verification_token, email_verification_expires, password_reset_token,
			   password_reset_expires, last_login, failed_login_attempts, locked_until,
			   created_at, updated_at
		FROM users 
		WHERE password_reset_token = $1 AND password_reset_expires > CURRENT_TIMESTAMP`

	user := &domain.User{}
	var emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil sql.NullTime

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.Phone, &user.IsActive, &user.IsVerified,
		&user.EmailVerificationToken, &emailVerificationExpires,
		&user.PasswordResetToken, &passwordResetExpires,
		&lastLogin, &user.FailedLoginAttempts, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired password reset token")
		}
		return nil, fmt.Errorf("failed to get user by password reset token: %w", err)
	}

	r.setNullTimes(user, emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil)
	return user, nil
}

func (r *userRepository) IncrementFailedLoginAttempts(ctx context.Context, userID uint) error {
	query := `UPDATE users SET failed_login_attempts = failed_login_attempts + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) ResetFailedLoginAttempts(ctx context.Context, userID uint) error {
	query := `UPDATE users SET failed_login_attempts = 0, locked_until = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) SetAccountLocked(ctx context.Context, userID uint, lockedUntil time.Time) error {
	query := `UPDATE users SET locked_until = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, lockedUntil, userID)
	if err != nil {
		return fmt.Errorf("failed to set account locked: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", userID)
	}

	return nil
}

// Helper methods
func (r *userRepository) createDefaultPreferences(ctx context.Context, userID uint) error {
	query := `
		INSERT INTO user_preferences (user_id, currency, language, theme, date_format, timezone)
		VALUES ($1, 'USD', 'en', 'light', 'YYYY-MM-DD', 'UTC')`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *userRepository) createDefaultNotificationSettings(ctx context.Context, userID uint) error {
	query := `
		INSERT INTO user_notification_settings (user_id, email_notifications, push_notifications, 
			weekly_reports, expense_alerts, budget_alerts, achievement_notifications)
		VALUES ($1, true, true, true, true, true, true)`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *userRepository) setNullTimes(user *domain.User, emailVerificationExpires, passwordResetExpires, lastLogin, lockedUntil sql.NullTime) {
	if emailVerificationExpires.Valid {
		user.EmailVerificationExpires = &emailVerificationExpires.Time
	}
	if passwordResetExpires.Valid {
		user.PasswordResetExpires = &passwordResetExpires.Time
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
