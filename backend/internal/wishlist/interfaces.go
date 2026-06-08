package wishlist

import "magic-collection-api/internal/mtgapi"

// wishlistRepository abstrai o acesso ao banco para o Service.
type wishlistRepository interface {
	List() ([]WishlistCard, error)
	GetByID(id string) (*WishlistCard, error)
	Create(w WishlistCard) (int64, error)
	Update(id string, input WishlistUpdateInput) error
	Delete(id string) error
	Acquire(id string, input AcquireInput) (int64, error)
}

// wishlistMtgClient abstrai chamadas à Scryfall para o Service.
type wishlistMtgClient interface {
	Search(setCode, number, lang, artist string) (*mtgapi.ExternalCard, error)
}

// wishlistService abstrai o Service para o Handler.
type wishlistService interface {
	List() ([]WishlistCard, error)
	GetByID(id string) (*WishlistCard, error)
	Create(input WishlistCardInput) (int64, error)
	Update(id string, input WishlistUpdateInput) error
	Delete(id string) error
	Acquire(id string, input AcquireInput) (int64, error)
}
