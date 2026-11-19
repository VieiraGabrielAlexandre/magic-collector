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
	game_sessions "magic-collection-api/internal/game_sessions"
	"magic-collection-api/internal/importer"
	"magic-collection-api/internal/mtgapi"
	"magic-collection-api/internal/tokens"
	"magic-collection-api/internal/lore"
	"encoding/json"
	"io"
	"net/http"
	"magic-collection-api/internal/reports"
	"magic-collection-api/internal/storage"
	"magic-collection-api/internal/trade"
	"magic-collection-api/internal/gamechangers"
	"magic-collection-api/internal/wishlist"
	wishlist_decks "magic-collection-api/internal/wishlist_decks"

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


	wishlistRepo := wishlist.NewRepository(db)
	wishlistSvc := wishlist.NewService(wishlistRepo, mtgClient)
	wishlistHandler := wishlist.NewHandler(wishlistSvc)

	wishlistDeckRepo := wishlist_decks.NewRepository(db)
	wishlistDeckHandler := wishlist_decks.NewHandler(wishlistDeckRepo, mtgClient, repository, deckRepo)

	gameSessionRepo := game_sessions.NewRepository(db)
	gameSessionSvc := game_sessions.NewService(gameSessionRepo, battleRepo)
	gameSessionHandler := game_sessions.NewHandler(gameSessionSvc)

	tokenRepo := tokens.NewRepository(db)
	tokenSvc := tokens.NewService(tokenRepo, mtgClient)
	tokenHandler := tokens.NewHandler(tokenSvc)

	storageRepo := storage.NewRepository(db)
	storageHandler := storage.NewHandler(storageRepo)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	tradeRepo := trade.NewRepository(db)
	tradeHandler := trade.NewHandler(tradeRepo)

	reportRepo := reports.NewRepository(db)
	reportHandler := reports.NewHandler(reportRepo).WithSessionCheck(func(token string) bool {
		_, _, err := authSvc.GetSession(token)
		return err == nil
	})

	router := gin.Default()

	// ── Rotas públicas ───────────────────────────────────────────────────
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/exchange-rate", func(c *gin.Context) {
		resp, err := http.Get("https://api.frankfurter.app/latest?from=USD&to=BRL")
		if err != nil {
			c.JSON(502, gin.H{"error": "falha ao buscar taxa de câmbio"})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(502, gin.H{"error": "resposta inválida"})
			return
		}
		c.JSON(200, data)
	})
	lore.RegisterRoutes(router)
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
	api.GET("/cards/subtypes", handler.Subtypes)
	api.GET("/cards/duplicates", handler.Duplicates)
	api.GET("/cards/export", handler.Export)
	api.GET("/cards/proxies", handler.ListProxies)
	api.POST("/cards/preview", handler.Preview)
	api.POST("/cards/refresh-prices", handler.RefreshPrices)
	api.POST("/cards/refresh-images", handler.RefreshImages)
	api.POST("/cards/suggest-decks", handler.SuggestDecks)
	api.POST("/cards/analyze-deck", handler.AnalyzeDeck)
	api.POST("/cards", handler.Create)
	api.GET("/cards/:id", handler.GetByID)
	api.PUT("/cards/:id", handler.Update)
	api.PATCH("/cards/:id/quantity", handler.UpdateQuantity)
	api.PATCH("/cards/:id/deck", handler.SetDeck)
	api.PATCH("/cards/:id/commander", handler.SetCommander)
	api.PATCH("/cards/:id/precon-original", handler.SetPreconOriginal)
	api.PATCH("/cards/:id/storage", handler.SetStorage)
	api.GET("/cards/:id/price-history", handler.GetPriceHistory)
	api.DELETE("/cards/:id", handler.Delete)

	api.GET("/decks", deckHandler.List)
	api.POST("/decks", deckHandler.Create)
	api.PUT("/decks/:id", deckHandler.Update)
	api.DELETE("/decks/:id", deckHandler.Delete)
	api.PATCH("/decks/:id/icon", deckHandler.FetchIcon)
	api.POST("/decks/:id/evaluate", deckHandler.Evaluate)
	api.GET("/decks/:id/validate", deckHandler.Validate)
	api.POST("/decks/import-precon", importerHandler.ImportPrecon)
	api.POST("/decks/import-list", importerHandler.ImportDeckList)
	api.POST("/decks/:id/import-cards", importerHandler.ImportCardsIntoDeck)

	api.GET("/battles", battleHandler.List)
	api.GET("/battles/stats", battleHandler.Stats)
	api.POST("/battles", battleHandler.Create)
	api.DELETE("/battles/:id", battleHandler.Delete)

	api.GET("/wishlist", wishlistHandler.List)
	api.POST("/wishlist", wishlistHandler.Create)
	api.GET("/wishlist/:id", wishlistHandler.GetByID)
	api.PUT("/wishlist/:id", wishlistHandler.Update)
	api.DELETE("/wishlist/:id", wishlistHandler.Delete)
	api.POST("/wishlist/:id/acquire", wishlistHandler.Acquire)

	api.GET("/wishlist-decks", wishlistDeckHandler.List)
	api.GET("/wishlist-decks/search-cards", wishlistDeckHandler.SearchCards)
	api.POST("/wishlist-decks", wishlistDeckHandler.Create)
	api.PATCH("/wishlist-decks/:id/acquire", wishlistDeckHandler.Acquire)
	api.DELETE("/wishlist-decks/:id", wishlistDeckHandler.Delete)
	api.GET("/wishlist-decks/:id/cards", wishlistDeckHandler.ListCards)
	api.POST("/wishlist-decks/:id/cards", wishlistDeckHandler.AddCard)
	api.POST("/wishlist-decks/:id/import-list", wishlistDeckHandler.ImportList)
	api.POST("/wishlist-decks/:id/cards/:card_id/acquire", wishlistDeckHandler.AcquireCard)
	api.PATCH("/wishlist-decks/:id/cards/:card_id/acquired", wishlistDeckHandler.ToggleCardAcquired)
	api.DELETE("/wishlist-decks/:id/cards/:card_id", wishlistDeckHandler.DeleteCard)

	api.GET("/tokens", tokenHandler.List)
	api.POST("/tokens/preview", tokenHandler.Preview)
	api.POST("/tokens", tokenHandler.Create)
	api.PATCH("/tokens/:id/quantity", tokenHandler.UpdateQuantity)
	api.DELETE("/tokens/:id", tokenHandler.Delete)

	api.GET("/storage", storageHandler.List)
	api.POST("/storage", storageHandler.Create)
	api.PUT("/storage/:id", storageHandler.Update)
	api.DELETE("/storage/:id", storageHandler.Delete)

	api.GET("/game-sessions", gameSessionHandler.List)
	api.POST("/game-sessions", gameSessionHandler.Create)
	api.GET("/game-sessions/:id", gameSessionHandler.GetByID)
	api.DELETE("/game-sessions/:id", gameSessionHandler.Delete)
	api.POST("/game-sessions/:id/players", gameSessionHandler.AddPlayer)
	api.PATCH("/game-sessions/:id/players/:player_id", gameSessionHandler.UpdatePlayer)
	api.DELETE("/game-sessions/:id/players/:player_id", gameSessionHandler.DeletePlayer)
	api.POST("/game-sessions/:id/reset", gameSessionHandler.Reset)
	api.POST("/game-sessions/:id/finish", gameSessionHandler.Finish)
	api.POST("/game-sessions/:id/restore", gameSessionHandler.Restore)
	api.POST("/game-sessions/:id/players/:player_id/revive", gameSessionHandler.RevivePlayer)
	api.PATCH("/game-sessions/:id/duration", gameSessionHandler.UpdateDuration)

	gcRepo := gamechangers.NewRepository(db)
	gcSvc  := gamechangers.NewService(gcRepo)
	gcHandler := gamechangers.NewHandler(gcSvc)
	api.GET("/gamechangers", gcHandler.List)
	api.POST("/gamechangers/sync", gcHandler.Sync)

	api.GET("/trade/gainers", tradeHandler.Gainers)
	api.GET("/trade/losers", tradeHandler.Losers)
	api.GET("/trade/stable", tradeHandler.Stable)
	api.GET("/trade/suggestions", tradeHandler.Suggestions)
	api.GET("/trade/price-spikes", tradeHandler.PriceSpikes)
	api.GET("/trade/portfolio-history", tradeHandler.PortfolioHistory)

	// Reports: próprio middleware que aceita SERVICE_TOKEN ou sessão normal.
	reportsGroup := router.Group("/", reportHandler.Middleware())
	reportsGroup.GET("/reports/collection", reportHandler.Get)
	reportsGroup.GET("/reports/collection/schema", reportHandler.Schema)

	router.Run(":8080")
}
