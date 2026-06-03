package main

import (
	"fmt"
	"log"
	"os"

	"magic-collection-api/internal/ai"
	"magic-collection-api/internal/battles"
	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/database"
	"magic-collection-api/internal/decks"
	"magic-collection-api/internal/importer"
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=10s&readTimeout=10s&writeTimeout=10s",
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

	aiClient := ai.NewClient(getenv("OPENAI_API_KEY", ""))

	repository := cards.NewRepository(db)
	mtgClient := mtgapi.NewClient()
	service := cards.NewService(repository, mtgClient)
	handler := cards.NewHandler(service, aiClient)

	deckRepo := decks.NewRepository(db)
	deckSvc := decks.NewService(deckRepo, mtgClient, repository, aiClient)
	deckHandler := decks.NewHandler(deckSvc)

	importerSvc := importer.NewService(deckRepo, repository, mtgClient)
	importerHandler := importer.NewHandler(importerSvc)

	battleRepo := battles.NewRepository(db)
	battleHandler := battles.NewHandler(battleRepo)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/cards", handler.List)
	router.GET("/cards/colors", handler.ListColors)
	router.GET("/cards/stats", handler.Stats)
	router.GET("/cards/export", handler.Export)
	router.POST("/cards/preview", handler.Preview)
	router.POST("/cards/refresh-prices", handler.RefreshPrices)
	router.POST("/cards/refresh-images", handler.RefreshImages)
	router.POST("/cards/suggest-decks", handler.SuggestDecks)
	router.POST("/cards", handler.Create)
	router.GET("/cards/:id", handler.GetByID)
	router.PUT("/cards/:id", handler.Update)
	router.PATCH("/cards/:id/quantity", handler.UpdateQuantity)
	router.PATCH("/cards/:id/deck", handler.SetDeck)
	router.DELETE("/cards/:id", handler.Delete)

	router.GET("/decks", deckHandler.List)
	router.POST("/decks", deckHandler.Create)
	router.PUT("/decks/:id", deckHandler.Update)
	router.DELETE("/decks/:id", deckHandler.Delete)
	router.PATCH("/decks/:id/icon", deckHandler.FetchIcon)
	router.POST("/decks/:id/evaluate", deckHandler.Evaluate)
	router.POST("/decks/import-precon", importerHandler.ImportPrecon)
	router.POST("/decks/import-list", importerHandler.ImportDeckList)
	router.POST("/decks/:id/import-cards", importerHandler.ImportCardsIntoDeck)

	router.GET("/battles", battleHandler.List)
	router.POST("/battles", battleHandler.Create)
	router.DELETE("/battles/:id", battleHandler.Delete)

	router.Run(":8080")
}
