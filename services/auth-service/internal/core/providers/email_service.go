package providers

import "context"

type EmailService interface {
	SendActivationEmail(ctx context.Context, to, token, userName string) error
	SendPasswordResetEmail(ctx context.Context, to, token, userName string) error
	SendWelcomeEmail(ctx context.Context, to, userName string) error
}
