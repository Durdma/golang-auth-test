package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"test-auth/auth"
	"test-auth/database"
	"test-auth/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetTokens(c *gin.Context) {
	guid := c.Query("guid")
	fmt.Println(guid)
	tokenManager, err := auth.NewManager("key")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("1")
	DB, err := database.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("2")
	recordCollection := database.GetCollection(DB)
	// fmt.Println("3")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// fmt.Println("4")
	if cursor, err := recordCollection.Find(ctx, bson.M{"guid": guid}); err == nil {
		if cursor.Next(ctx) {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "This guid already has pair of tokens"})
			return
		}
	}
	// fmt.Println("5")
	accessToken, err := tokenManager.NewJWT(guid, 2*time.Minute)
	expiresAt := time.Now().Add(2 * time.Minute).Unix()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("6")
	refreshToken, err := tokenManager.NewRefreshToken()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("7")
	postRecord := models.Record{Guid: guid, RefreshToken: refreshToken, ExpiresAt: time.Now().Add(2 * time.Minute)}
	// fmt.Println("8")
	_, err = recordCollection.InsertOne(ctx, postRecord)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("9")
	// fmt.Println(accessToken)
	// fmt.Println(refreshToken)
	c.SetCookie("refreshToken", refreshToken, int(expiresAt), "/auth", "localhost", false, true)
	//fmt.Println(c.Cookie("refreshToken"))
	c.IndentedJSON(http.StatusCreated, models.Session{AccessToken: accessToken, RefreshToken: refreshToken})

}
