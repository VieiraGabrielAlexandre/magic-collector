package main

import (
	"log"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/database"
	"magic-collection-api/internal/mtgapi"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.Open("./data/collection.db")
	if err != nil {
		log.Fatal(err)
	}

	repository := cards.NewRepository(db)
	mtgClient := mtgapi.NewClient()
	service := cards.NewService(repository, mtgClient)
	handler := cards.NewHandler(service)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/cards", handler.List)
	router.POST("/cards", handler.Create)
	router.GET("/cards/:id", handler.GetByID)
	router.DELETE("/cards/:id", handler.Delete)

	router.Run(":8080")
}
