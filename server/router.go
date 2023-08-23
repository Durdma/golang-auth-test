// Пакет, в котором описываются маршруты и функции-обработчики запросов
package server

import (
	"net/http"
	"test-auth/service"

	"github.com/gin-gonic/gin"
)

// Создание маршрутов и добавление к ним функций обработчиков
func NewRouter() *gin.Engine {
	router := gin.Default()

	// Тест работы роутера/сервера
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello")
	})

	// Первый энд-поинт задания.
	// Выдача пары accessToken и refreshToken для пользователя по его GUID
	router.POST("/auth/get-tokens", service.GetTokens)

	// Второй энд-поинт задания.
	// Обновление пары accessToken и refreshToken
	router.POST("/auth/refresh-tokens", service.RefreshTokens)

	return router

}
