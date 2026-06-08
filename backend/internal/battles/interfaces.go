package battles

type battleRepo interface {
	List() ([]Battle, error)
	Create(b BattleInput) (int64, error)
	Delete(id string) error
}
