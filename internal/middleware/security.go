package middleware

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

// SecurityHeaders добавляет заголовки безопасности
func SecurityHeaders(isProd bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        if isProd {
            // HSTS - требовать HTTPS
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

            // Защита от XSS
            c.Header("X-XSS-Protection", "1; mode=block")

            // Запрет на встраивание в iframe
            c.Header("X-Frame-Options", "DENY")

            // Защита от MIME-sniffing
            c.Header("X-Content-Type-Options", "nosniff")

            // Политика Referrer
            c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        }
        c.Next()
    }
}

// RateLimiter защита от DDoS (простая)
func RateLimiter(limit rate.Limit, burst int) gin.HandlerFunc {
    limiter := rate.NewLimiter(limit, burst)

    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(429, gin.H{
                "error": "Too many requests. Please slow down.",
            })
            return
        }
        c.Next()
    }
}

// CORSMiddleware с настройками под окружение
func CORSMiddleware(isProd bool, allowedOrigin string) gin.HandlerFunc {
    return func(c *gin.Context) {
        if isProd {
            // В продакшене - только твой домен
            origin := c.GetHeader("Origin")
            if origin == allowedOrigin || origin == "" {
                c.Header("Access-Control-Allow-Origin", origin)
            }
            c.Header("Access-Control-Allow-Credentials", "true")
            c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        } else {
            // В разработке - всё разрешено
            c.Header("Access-Control-Allow-Origin", "*")
            c.Header("Access-Control-Allow-Methods", "*")
            c.Header("Access-Control-Allow-Headers", "*")
        }

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
