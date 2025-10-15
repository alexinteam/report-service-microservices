package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secretKey []byte
}

func NewManager(secretKey string) *Manager {
	return &Manager{
		secretKey: []byte(secretKey),
	}
}

func (m *Manager) GenerateToken(userID uint, username, email, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "api-gateway",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		logrus.WithError(err).Error("Ошибка генерации JWT токена")
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return tokenString, nil
}

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		logrus.WithError(err).Error("Ошибка парсинга JWT токена")
		return nil, fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("невалидный токен")
}

// RefreshToken обновляет JWT токен
func (m *Manager) RefreshToken(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("ошибка валидации токена для обновления: %w", err)
	}

	if time.Since(claims.IssuedAt.Time) > 7*24*time.Hour {
		return "", errors.New("токен слишком старый для обновления")
	}

	newToken, err := m.GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Role)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации нового токена: %w", err)
	}

	return newToken, nil
}

// ExtractTokenFromHeader извлекает токен из Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("отсутствует Authorization header")
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("неверный формат Authorization header")
	}

	token := authHeader[7:]
	if token == "" {
		return "", errors.New("отсутствует токен в Authorization header")
	}

	return token, nil
}
