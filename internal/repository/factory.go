package repository

type RepositoryType string

const (
	PostgresType RepositoryType = "postgres"
	MemoryType   RepositoryType = "memory"
	JSONType     RepositoryType = "json"
)

type Config struct {
	Type     RepositoryType
	DSN      string // Postgres
	Filename string // JSON
}

func NewRepository(config Config) (Repository, error) {
	switch config.Type {
	case PostgresType:
		return NewPostgresRepository(config.DSN)
	case MemoryType:
		return NewMemoryRepository(), nil
	case JSONType:
		return NewJSONFileRepository(config.Filename)
	default:
		return NewMemoryRepository(), nil
	}
}
