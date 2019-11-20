package tar

import (
	"github.com/mholt/archiver"
)

// Tar implements tar compressor
type Tar struct {
}

// Decompress a bzip2 file
func (b *Tar) Decompress(origin, destination string, filesize int) error {
	return archiver.Unarchive(origin, destination)
}
