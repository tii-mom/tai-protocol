package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles token issuance and validation for TAI Protocol API.
type JWTManager struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		expiration: 7 * 24 * time.Hour, // 7 days
	}
}

// Claims is the JWT payload for TAI Protocol users.
type Claims struct {
	UserID    int64  `json:"user_id"`
	TGUserID  int64  `json:"tg_user_id"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

// IssueToken creates a JWT for an authenticated user.
func (m *JWTManager) IssueToken(userID, tgUserID int64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TGUserID: tgUserID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "tai-protocol",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken parses and validates a JWT, returning claims.
func (m *JWTManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// Middleware returns a Gin middleware that validates JWT from Authorization header.
func (m *JWTManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := m.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Store claims in context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("tg_user_id", claims.TGUserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
