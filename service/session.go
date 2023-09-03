// Осовная бизнес-логика API аутентификации.
package service

import (
	"net/http"
	"regexp"
	"test-auth/auth"
	"test-auth/database"
	"test-auth/models"
	"time"

	"github.com/gin-gonic/gin"
)

// Валидация пользовательского GUID, полученного из query string
//
//	Формат GUID: {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX},
//
//	где Х - [0-9] || [A-Fa-f]
//
// Функция возврщает bool, GUID прошел проверку или нет
func validGUID(guid string) bool {
	pattern := regexp.MustCompile("^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$")

	return pattern.Match([]byte(guid))
}

// Функция генерирует и шифрует токены refreshToken, accessToken по полученному GUID;
//
// Добавляет запись о сессии в MongoDB с сигнатурой models.Record из пакета models;
// Добавляет в cookie информацию о refreshToken.
func createSessionAndCookie(c *gin.Context, db *database.Database,
	tokenManager *auth.Manager, guid string) {
	encodedRefreshToken, hashedRefreshToken, accessToken, err := tokenManager.GenerateTokens(guid)
	if err != nil {
		// Возвращает ошибку связанную с генерацией токенов
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	expiresAt := int(auth.RefreshTokenDuration / time.Second) //Время жизни токена для cookie
	accessExpiresAt := int(auth.AccessTokenDuration / time.Second)

	newSession := models.Record{Guid: guid, RefreshToken: hashedRefreshToken, ExpiresAt: time.Now().Add(auth.RefreshTokenDuration)}

	_, err = db.AddToken(newSession)
	if err != nil {
		// Возвращает ошибку связанную с добавлением записи о сессии в БД
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	// Запись refreshToken'a в cookie приложения пользоваеля
	c.SetCookie("refreshToken", encodedRefreshToken, expiresAt,
		"/auth", "localhost", false, true)

	c.SetCookie("accessToken", accessToken, accessExpiresAt,
		"/auth", "localhost", false, true)

	// РАСКОМЕННТИТЬ ЕСЛИ ТЕСТИРОВАНИЕ ЧЕРЕЗ CURL или неоходимо посмотреть refreshToken, передаваемый клиенту
	//fmt.Println(encodedRefreshToken)

	// Запись accessToken'a в тело Response; Возврщает сообщние об успешном создании
	//c.JSON(http.StatusCreated, models.Response{AccessToken: accessToken})
}

// Функция обрабатывает запрос на создание пары accessToken'a и refreshToken'a по пользовательскому GUID
//
// Формирует Response, в теле которого будет храниться accessToken для приложения пользователя;
// Устанавливает в cookie значение refreshToken
func GetTokens(c *gin.Context) {
	guid := c.Query("guid")

	if !validGUID(guid) {
		// Возвращает ошибку о том, что GUID не соответствует формату
		c.JSON(http.StatusBadRequest, "invalid guid!")
		return
	}

	// Инициализируем сущность для генерации токенов
	tokenManager, err := auth.NewManager()
	if err != nil {
		// Возвращает ошибку связанную с созданием сущности для генерации токенов
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Создаем подключение к БД
	db, err := database.NewClient()
	if err != nil {
		// Возвращает ошибку связанную с созданием подключения к БД
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	createSessionAndCookie(c, db, tokenManager, guid)
}

// Функция обновляет пару acceessToken и refreshToken.
// Функция получает значение refreshToken из cookie.
// Проверяет наличие записи о сессии с таким refreshToken.
// Если такая сессия существует и токен еще не просрочен, то генерирует новую пару accessToken, refreshToken,
// иначе возвращает Response с ошибкой
func RefreshTokens(c *gin.Context) {
	userRefreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	userAccessToken, err := c.Cookie("accessToken")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if !auth.CheckPairOfTokens(userRefreshToken, userAccessToken) {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid pair of access and refresh tokens")
		return
	}

	// Инициализируем сущность для генерации токенов
	tokenManager, err := auth.NewManager()
	if err != nil {
		// Возвращает ошибку связанную с созданием сущности для генерации токенов
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Создаем подключение к БД
	db, err := database.NewClient()
	if err != nil {
		// Возвращает ошибку связанную с созданием подключения к БД
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Декодирование refreshToken из base64
	decodedRefreshToken, err := auth.DecodeRefreshToken(userRefreshToken)
	if err != nil {
		// Возвращает ошибку связанную с декодированием refreshToken
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var guid string

	// Поиск сессии с заданным refreshToken.
	// Генерация ответов при различных результатах поиска
	switch state, gguid, recordId, err := db.CheckToken(decodedRefreshToken); err {
	case nil:
		switch state {
		case "token expired!":
			err = db.DelToken(recordId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}
			c.JSON(http.StatusUnauthorized, state)
			return
		case "wrong token!":
			c.JSON(http.StatusNotFound, state)
			return
		default:
			// Удаление старой сессии при совпадении refreshToken
			err = db.DelToken(recordId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}
			guid = gguid
		}
	default:
		// Возвращает Response с ошибкой связанной с поиском записей по БД
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	createSessionAndCookie(c, db, tokenManager, guid)
}
