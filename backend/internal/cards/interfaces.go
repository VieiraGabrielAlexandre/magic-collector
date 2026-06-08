package cards

import "magic-collection-api/internal/mtgapi"

// cardRepository abstrai o acesso ao banco para o Service.
type cardRepository interface {
	Create(card Card) (int64, error)
	List(params ListParams) (ListResult, error)
	GetByID(id string) (*Card, error)
	Update(id string, card Card) error
	UpdateSharedByIdentity(oldName, oldSetCode, oldCollNum, oldLang string, oldFoil bool, card Card) error
	UpdateMTGID(id, mtgID string) error
	Delete(id string) error
	SetQuantity(id string, quantity int) error
	SetDeck(id string, deckID int) error
	ListAll() ([]Card, error)
	ListAllForPriceRefresh() ([]CardForPriceRefresh, error)
	ListEmptyPricesForRefresh() ([]CardForPriceRefresh, error)
	UpdatePrice(id int, price float64) error
	UpdatePriceAndMTGID(id int, mtgID string, price float64) error
	UpdateImageURL(id int, imageURL string) error
	UpdateImageURLAndMTGID(id int, mtgID, imageURL string) error
	ListCardsWithoutColors() ([]CardForPriceRefresh, error)
	NormalizeRarities() (NormalizeRarityResult, error)
	UpdateColors(id int, colors, color string) error
	UpdateColorsAndMTGID(id int, mtgID, colors, color string) error
	ListColorCombos() ([]ColorCombo, error)
	GetStats() (CollectionStats, error)
	ListForDeckBuilder() ([]DeckBuilderCard, error)
}

// mtgAPIClient abstrai chamadas à Scryfall para o Service.
type mtgAPIClient interface {
	Search(setCode, number, lang, artist string) (*mtgapi.ExternalCard, error)
	SearchPreRelease(name, lang, artist string) (*mtgapi.ExternalCard, error)
	GetByMTGID(id string) (*mtgapi.ExternalCard, error)
}

// aiCompleter abstrai chamadas à API de IA para o Handler.
type aiCompleter interface {
	Complete(prompt string) (string, error)
}

// cardService abstrai o Service para o Handler.
type cardService interface {
	Create(input CreateCardInput) (int64, error)
	List(params ListParams) (ListResult, error)
	GetByID(id string) (map[string]any, error)
	Update(id string, input UpdateCardInput) error
	Delete(id string) error
	SetDeck(id string, deckID int) error
	NormalizeRarities() (NormalizeRarityResult, error)
	RefreshColors() (RefreshColorsResult, error)
	ListColorCombos() ([]ColorCombo, error)
	SetQuantity(id string, quantity int) error
	Preview(input PreviewCardInput) (*mtgapi.ExternalCard, error)
	GetStats() (CollectionStats, error)
	RefreshImages() (ImageRefreshResult, error)
	RefreshPrices(emptyOnly bool) (PriceRefreshResult, error)
	GetCardsForDeckBuilder() ([]DeckBuilderCard, error)
	ExportAll() ([]Card, error)
}
