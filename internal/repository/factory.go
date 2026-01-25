package repository

const (
	// PostgresType indicates a PostgreSQL repository.
	PostgresType RepositoryType = "postgres"
	// MemoryType indicates an in-memory repository.
	MemoryType RepositoryType = "memory"
	// JSONType indicates a JSON file-based repository.
	JSONType RepositoryType = "json"
)

// Repository defines the methods for interacting with the URL storage.
type RepositoryType string

// Config holds the configuration for creating a repository.
type Config struct {
	Type     RepositoryType
	DSN      string // Postgres
	Filename string // JSON
}

// NewRepository creates a new Repository based on the provided configuration.
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
