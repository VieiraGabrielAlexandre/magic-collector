package importer

// importerService abstrai o Service para o Handler.
type importerService interface {
	ImportPrecon(input ImportPreconInput) (ImportResult, error)
	ImportDeckList(input ImportDeckListInput) (ImportResult, error)
	ImportCardsIntoDeck(deckID int64, input ImportCardsToDeckInput) (ImportResult, error)
}
