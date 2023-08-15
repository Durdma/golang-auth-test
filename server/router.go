package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	router.POST("/get-jwt", func(c *gin.Context) {
		guid := c.Query("guid")
		fmt.Println(guid)
		c.String(http.StatusOK, "test guid %s", guid)
	})

	router.POST("/refresh-jwt", func(c *gin.Context) {
		token := c.Query("token")
		fmt.Println(token)
		c.String(http.StatusOK, "test token %s", token)
	})

	return router

}
