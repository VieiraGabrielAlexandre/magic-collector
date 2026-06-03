package main

import (
	"fmt"
	"log"
	"os"

	"magic-collection-api/internal/ai"
	"magic-collection-api/internal/auth"
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

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	router := gin.Default()

	// ── Rotas públicas ───────────────────────────────────────────────────
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/logout", authHandler.Logout)
	router.GET("/auth/me", authHandler.Me)

	// ── Rotas protegidas (requerem sessão válida) ────────────────────────
	api := router.Group("/", authHandler.Middleware())

	api.GET("/cards", handler.List)
	api.GET("/cards/colors", handler.ListColors)
	api.POST("/cards/refresh-colors", handler.RefreshColors)
	api.POST("/cards/normalize-rarities", handler.NormalizeRarities)
	api.GET("/cards/stats", handler.Stats)
	api.GET("/cards/export", handler.Export)
	api.POST("/cards/preview", handler.Preview)
	api.POST("/cards/refresh-prices", handler.RefreshPrices)
	api.POST("/cards/refresh-images", handler.RefreshImages)
	api.POST("/cards/suggest-decks", handler.SuggestDecks)
	api.POST("/cards", handler.Create)
	api.GET("/cards/:id", handler.GetByID)
	api.PUT("/cards/:id", handler.Update)
	api.PATCH("/cards/:id/quantity", handler.UpdateQuantity)
	api.PATCH("/cards/:id/deck", handler.SetDeck)
	api.DELETE("/cards/:id", handler.Delete)

	api.GET("/decks", deckHandler.List)
	api.POST("/decks", deckHandler.Create)
	api.PUT("/decks/:id", deckHandler.Update)
	api.DELETE("/decks/:id", deckHandler.Delete)
	api.PATCH("/decks/:id/icon", deckHandler.FetchIcon)
	api.POST("/decks/:id/evaluate", deckHandler.Evaluate)
	api.POST("/decks/import-precon", importerHandler.ImportPrecon)
	api.POST("/decks/import-list", importerHandler.ImportDeckList)
	api.POST("/decks/:id/import-cards", importerHandler.ImportCardsIntoDeck)

	api.GET("/battles", battleHandler.List)
	api.POST("/battles", battleHandler.Create)
	api.DELETE("/battles/:id", battleHandler.Delete)

	router.Run(":8080")
}
