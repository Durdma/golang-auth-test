package server

import (
	"fmt"
	"net/http"
	"test-auth/models"
	"test-auth/service"

	"github.com/gin-gonic/gin"
)

// func tmp(c *gin.Context) {
// 	DB, err := database.NewClient()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	postCollection := database.GetCollection(DB, "RecordsTest")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

// 	defer cancel()

// 	fmt.Println(postCollection.CountDocuments(ctx, bson.D{}))
// 	postPay := models.Record{ID: primitive.NewObjectID(),
// 		Name: "TEST",
// 		Sub:  "TEST2"}
// 	smth, err := postCollection.InsertOne(ctx, postPay)
// 	if err != nil {
// 		fmt.Println(err)
// 	} else {
// 		fmt.Println("OK", smth)
// 	}

// }

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello")
	})

	router.POST("/auth", service.GetTokens)

	router.POST("/get-jwt", func(c *gin.Context) {
		guid := c.Query("guid")
		fmt.Println(guid)
		c.String(http.StatusOK, "test guid %s", guid)
		c.JSON(http.StatusCreated, models.Session{})
	})

	router.POST("/refresh-jwt", func(c *gin.Context) {
		token := c.Query("token")
		fmt.Println(token)
		c.String(http.StatusOK, "test token %s", token)
	})

	return router

}
