package storage

import (
	"denver/pkg/storage/http"
	natives3 "denver/pkg/storage/nativeS3"
)

// Storage represents a storage driver
type Storage interface {
	Download(origin, destination string) (string, error)
}

// MultiStorage supports multiple storage
type MultiStorage struct {
	storages map[string]Storage
}

// NewMultiStorage returns a pointer to MultiStorage
func NewMultiStorage() *MultiStorage {
	return &MultiStorage{storages: map[string]Storage{
		"http": &http.HTTP{},
		"s3":   &natives3.NativeS3{},
	}}
}

// Download a file from origin to destination
func (m *MultiStorage) Download(origin, destination string) (string, error) {

	storageType := "s3"
	s, ok := m.storages[storageType]
	if !ok {
		return "", nil
	}
	destFile, error := s.Download(origin, destination)

	return destFile, error

	//for {
	//	storageType := "s3"
	//	s, ok := m.storages[storageType]
	//	if !ok {
	//		return "", nil
	//	}
	//	destFile, error := s.Download(origin, destination)
	//	if error != nil {
	//		return destFile, error
	//	}
	//}
}
