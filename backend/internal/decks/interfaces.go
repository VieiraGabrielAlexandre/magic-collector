package decks

import (
	"magic-collection-api/internal/cards"
	"magic-collection-api/internal/mtgapi"
)

// deckRepository abstrai o acesso ao banco para o Service.
type deckRepository interface {
	List() ([]Deck, error)
	GetByID(id string) (*Deck, error)
	Create(input DeckInput) (int64, error)
	Update(id string, input DeckInput) error
	UpdateIcon(id, iconURI string) error
	UpdateEvaluation(id, evaluation string) error
	Delete(id string) error
}

// deckMtgClient abstrai chamadas à Scryfall para o Service.
type deckMtgClient interface {
	GetSetByCode(code string) (*mtgapi.SetInfo, error)
}

// deckCardRepo abstrai acesso às cartas para avaliação de deck.
type deckCardRepo interface {
	ListForEval(deckID int) ([]cards.EvalCardInfo, error)
}

// deckAiClient abstrai chamadas à API de IA para o Service.
type deckAiClient interface {
	Complete(prompt string) (string, error)
}

// deckService abstrai o Service para o Handler.
type deckService interface {
	List() ([]Deck, error)
	Create(input DeckInput) (int64, error)
	Update(id string, input DeckInput) error
	Delete(id string) error
	FetchIcon(id string) (string, error)
	EvaluateDeck(id string) (*Deck, error)
}
