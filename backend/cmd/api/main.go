package main

import (
	"fmt"
	"log"
	"os"

	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/database"
	"magic-collection-api/internal/decks"
	"magic-collection-api/internal/mtgapi"

	"github.com/gin-gonic/gin"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getenv("DB_USER", "root"),
		getenv("DB_PASSWORD", ""),
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "magic_collector"),
	)

	db, err := database.Open(dsn)
	if err != nil {
		log.Fatal(err)
	}

	repository := cards.NewRepository(db)
	mtgClient := mtgapi.NewClient()
	service := cards.NewService(repository, mtgClient)
	handler := cards.NewHandler(service)

	deckRepo := decks.NewRepository(db)
	deckHandler := decks.NewHandler(deckRepo)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/cards", handler.List)
	router.GET("/cards/export", handler.Export)
	router.POST("/cards", handler.Create)
	router.GET("/cards/:id", handler.GetByID)
	router.PUT("/cards/:id", handler.Update)
	router.PATCH("/cards/:id/deck", handler.SetDeck)
	router.DELETE("/cards/:id", handler.Delete)

	router.GET("/decks", deckHandler.List)
	router.POST("/decks", deckHandler.Create)
	router.PUT("/decks/:id", deckHandler.Update)
	router.DELETE("/decks/:id", deckHandler.Delete)

	router.Run(":8080")
}
