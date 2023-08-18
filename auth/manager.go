package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//TODO эндпоинт на получение пары готов, бд поднята, необходимо реализовать второй эндпоинт
//TODO accessToken sha512 +; refreshToken в БД в bcrypt; refreshToken закодировать в base64 для передачи;
//TODO сделать использование refreshToken одноразовым, после использования генерировать новую пару токенов по факту

// Logic for access and refresh tokens generation
type TokenManager interface {
	NewJWT(guid string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

// keep signing string
type Manager struct {
	signingKey string
}

// Create new token manager with signing string
func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

// Create new access token
func (m *Manager) NewJWT(guid string, ttl time.Duration) (string, error) {
	expiresAt := time.Now().Add(ttl).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Subject:   guid,
	})

	return token.SignedString([]byte(m.signingKey))
}

// Parse access token
func (m *Manager) Parse(accessToken string) (string, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.signingKey), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("error get user claims from token")
	}

	return claims["sub"].(string), nil
}

// Generate new refresh token with base64 encoding
func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func EncodeRefreshToken(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func DecodeRefreshToken(token string) (string, error) {
	encodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	return string(encodedToken), nil
}
