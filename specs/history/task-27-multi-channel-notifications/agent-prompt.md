# Agent Prompt: Task 27 - Multi-Channel Notification Delivery

## 🤖 Agent Identity
Expert Go Engineer for multi-channel notification systems with email, Slack, SMS, and scheduled digests.

---

## 📋 Mission
Build multi-channel notification delivery: SendGrid email, Slack webhooks, Twilio SMS, and daily digest generation.

---

## 📂 Files to Create

### Delivery Adapters (internal/adapters/delivery/)
- `email/sendgrid_client.go` + `_test.go`
- `slack/client.go` + `_test.go`
- `sms/twilio_client.go` + `_test.go`

### Core Components
- `internal/usecases/deliver_notification.go` + `_test.go`
- `internal/usecases/generate_digest.go` + `_test.go`
- `internal/domain/entities/delivery_queue.go`
- `internal/domain/entities/channel_config.go`

---

## 🔧 Email Delivery (SendGrid)

```go
type SendGridClient struct {
    apiKey string
    client *sendgrid.Client
}

func (s *SendGridClient) SendNotification(ctx context.Context, notif Notification, user User) error
func (s *SendGridClient) SendDigest(ctx context.Context, digest Digest, user User) error
```

---

## 🔧 Slack Integration

```go
type SlackClient struct {
    webhookURL string
}

func (s *SlackClient) PostNotification(ctx context.Context, notif Notification, channel string) error {
    // Format as Slack Block Kit message
    // Include: product, zone status, top recommendation
}
```

---

## 🔧 SMS Delivery (Twilio)

```go
type TwilioClient struct {
    accountSID string
    authToken  string
    fromNumber string
}

func (t *TwilioClient) SendSMS(ctx context.Context, notif Notification, phoneNumber string) error {
    // Only for critical alerts, max 160 chars
}
```

---

## 🔧 Delivery Logic

- Check quiet hours before sending
- Deliver by channel based on priority
- Respect user preferences
- Use retry queue for failed deliveries

---

## 🔧 Daily Digest

Generate and send at configured time with:
- Total count, counts by priority and type
- Unacted count, top items by priority

---

## ✅ Success Criteria
- [ ] Email delivery >99% success rate
- [ ] Slack delivery <2s from notification
- [ ] SMS delivery <30s for critical
- [ ] Daily digest within 5 min of configured time
- [ ] Zero notification loss (retry queue)
- [ ] Quiet hours respected
- [ ] 85%+ test coverage

---

## 🚀 Commands
```bash
cd services/ai-intelligence-hub
export SENDGRID_API_KEY=your_key
export SLACK_WEBHOOK_URL=https://hooks.slack.com/...
export TWILIO_ACCOUNT_SID=your_sid
export TWILIO_AUTH_TOKEN=your_token
go test ./internal/adapters/delivery/... -cover
```
