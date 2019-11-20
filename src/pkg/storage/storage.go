package storage

// Storage represents a storage driver
type Storage interface {
	Download(origin, destination string) (string, error)
}
