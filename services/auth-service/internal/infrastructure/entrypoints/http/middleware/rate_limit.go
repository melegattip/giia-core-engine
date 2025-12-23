package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type RateLimitMiddleware struct {
	rateLimiter providers.RateLimiter
}

func NewRateLimitMiddleware(rateLimiter providers.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: rateLimiter,
	}
}

func (m *RateLimitMiddleware) LimitLogin() gin.HandlerFunc {
	return m.limit(5, 15*time.Minute, "login")
}

func (m *RateLimitMiddleware) LimitRegister() gin.HandlerFunc {
	return m.limit(3, 60*time.Minute, "register")
}

func (m *RateLimitMiddleware) limit(maxAttempts int, window time.Duration, operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("%s:%s", operation, ip)

		allowed, retryAfter, err := m.rateLimiter.CheckRateLimit(c.Request.Context(), key, maxAttempts, window)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, pkgErrors.ToHTTPResponse(
				pkgErrors.NewBadRequest(fmt.Sprintf("too many %s attempts, please try again later", operation)),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
