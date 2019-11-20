package compressor

import (
	"denver/pkg/util/compressor/bzip2"
	"denver/pkg/util/compressor/tar"
	"denver/pkg/util/compressor/zip"
	"path/filepath"
	"strings"
)

// Compressor represents a compressor engine
type Compressor interface {
	Decompress(origin, destination string, filesize int) error
}

// MultiCompressor supports multiple compressions
type MultiCompressor struct {
	compressors map[string]Compressor
}

// NewMultiCompressor returns a pointer to MultiCompressor
func NewMultiCompressor() *MultiCompressor {
	return &MultiCompressor{compressors: map[string]Compressor{
		".zip": &zip.Zip{},
		".bz2": &bzip2.Bzip2{},
		".tar": &tar.Tar{},
	}}
}

// Decompress a file in origin to destination
func (m *MultiCompressor) Decompress(origin, destination string, filesize int) error {
	for {
		ext := filepath.Ext(origin)
		c, ok := m.compressors[ext]
		if !ok {
			return nil
		}
		error := c.Decompress(origin, destination, filesize)
		if error != nil {
			return error
		}
		origin = strings.TrimSuffix(origin, ext)
	}
}
