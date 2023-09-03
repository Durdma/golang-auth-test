package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
)

// Необходимые константы. НА ПРОДЕ КОНСТАНТЫ ДОЛЖНЫ БЫТЬ ПЕРЕМЕЩЕНЫ В ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ
//
//	signKey - хранит специальную строку для подписи accessToken
//	AccessTokenDuration - хранит время жизни AccessToken
//	RefreshTokenDuration - хранит время жизни RefreshToken
const (
	signKey              = "key"
	AccessTokenDuration  = 2 * time.Minute
	RefreshTokenDuration = 5 * time.Minute
)

// Сущность для генерации accessToken и refreshToken
type Manager struct {
	signingKey string
}

// Функция создает сущность для генерации токенов.
func NewManager() (*Manager, error) {
	signingKey := signKey
	if signingKey == "" {
		// Возвращает ошибку связанную с тем, что спец строка для подписи пуста
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

// Функция генерирует пару accessToken и RefreshToken
// На вход фунция принимает пользовательский GUID
// Функция возвращает:
//
//	зашифрованный в base64 refreshToken;
//	захэшированный с помощью sha512 refershToken;
//	полученный accessToken
func (m *Manager) GenerateTokens(guid string) (encodedRefreshToken, hashedRefreshToken,
	accessToken string, err error) {
	accessToken, err = m.newJWT(guid)
	if err != nil {
		// Возвращает ошибку связанную с генерацией нового accessToken
		return "", "", "", err
	}

	refreshToken, err := m.newRefreshToken(accessToken[len(accessToken)-6:])
	if err != nil {
		// Возвращает ошибку связанную с генерацией нового refrehToken
		return "", "", "", err
	}

	encodedRefreshToken = encodeRefreshToken(refreshToken)
	hashedRefreshToken, err = hashRefreshToken(refreshToken)
	if err != nil {
		// Возвращает ошибку связанную с хэшированием refreshToken
		return "", "", "", err
	}

	err = nil
	return
}

// Функция генерирует новый accessToken на основе пользовательского GUID
func (m *Manager) newJWT(guid string) (string, error) {
	expiresAt := time.Now().Add(AccessTokenDuration).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Subject:   guid,
	})

	return token.SignedString([]byte(m.signingKey))
}

// Функция генерирует новый refreshToken.
// refreshToken генерируется с помощью стандартного пакета rand.
// Генератор значений в качестве сида берет время в unix формате.
// В конец зашиваются байт коды 6 последних символов accessToken
//
// Длина refreshToken 32 байта
func (m *Manager) newRefreshToken(sixCharAT string) (string, error) {
	b := make([]byte, 26)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	for i := 0; i < len(sixCharAT); i++ {
		b = append(b, byte(sixCharAT[i]))
	}

	return fmt.Sprintf("%x", b), nil
}

// Функция кодирует refreshToken в base64
func encodeRefreshToken(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// Функция декодирует refreshToken из base64 в обычный набор байт
func DecodeRefreshToken(token string) (string, error) {
	encodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	return string(encodedToken), nil
}

// Функция хэширует refreshToken в bcrypt-хэш
func hashRefreshToken(token string) (string, error) {
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)

	return string(hashedRefreshToken), err

}

// Функция сравнивает сумму полученного от пользователя refreshToken с хэш-суммой хэшированного refreshToken
func CheckRefreshToken(token, hashedToken string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(token))

	return err == nil

}

// Функция сравнивает последние 6 символов refreshToken и accessToken для подтверждения подлинности
// токенов
func CheckPairOfTokens(refreshToken, accessToken string) bool {
	decodedRT, err := DecodeRefreshToken(refreshToken)
	if err != nil {
		return false
	}

	lastSixAT := []byte{}

	for i := len(accessToken) - 3; i < len(accessToken); i++ {
		lastSixAT = append(lastSixAT, byte(accessToken[i]))
	}

	return string(lastSixAT) != decodedRT[len(decodedRT)-6:]
}
