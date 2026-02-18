package websocket

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type WebSocketClaims struct {
	UserID   string `json:"user_id"`
	ServerID string `json:"server_id"`
	jwt.RegisteredClaims
}

type WebSocketAuthenticator struct {
	jwtSecret string
	logger    *logrus.Logger
}

func NewWebSocketAuthenticator(jwtSecret string, logger *logrus.Logger) *WebSocketAuthenticator {
	return &WebSocketAuthenticator{
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// ValidateToken validates a WebSocket JWT token
func (a *WebSocketAuthenticator) ValidateToken(tokenString string) (*WebSocketClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &WebSocketClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		a.logger.WithError(err).Error("Failed to parse WebSocket token")
		return nil, err
	}

	if claims, ok := token.Claims.(*WebSocketClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("token expired")
		}

		a.logger.WithFields(logrus.Fields{
			"user_id":   claims.UserID,
			"server_id": claims.ServerID,
		}).Debug("WebSocket token validated successfully")

		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken generates a new WebSocket JWT token (for testing purposes)
func (a *WebSocketAuthenticator) GenerateToken(userID, serverID string, duration time.Duration) (string, error) {
	claims := WebSocketClaims{
		UserID:   userID,
		ServerID: serverID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		a.logger.WithError(err).Error("Failed to generate WebSocket token")
		return "", err
	}

	return tokenString, nil
}
