package repositories

type WhereClause struct {
	Field    string
	Operator string
	Value    interface{}
}

// BaseRepository defines common methods for all repositories.
type BaseRepository[T any] interface {
	Set(entity *T) error
	Delete(id string) error
	Query(
		whereClauses []WhereClause,
		sortBy string,
		descending bool,
		limit int,
		offset int,
	) ([]T, error)
}
