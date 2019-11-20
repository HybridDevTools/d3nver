package bzip2

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"

	"github.com/dsnet/compress/bzip2"
)

// Bzip2 implements bzip2 compressor
type Bzip2 struct {
}

// Decompress a bzip2 file
func (b *Bzip2) Decompress(origin, destination string, filesize int) error {
	f, err := os.Open(origin)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := bzip2.NewReader(bufio.NewReader(f), nil)
	if err != nil {
		return err
	}
	defer r.Close()

	_, file := filepath.Split(origin)
	extension := filepath.Ext(file)
	file = strings.TrimSuffix(file, extension)
	destFile := filepath.Join(destination, file)

	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	tmpl := `{{ green "Progress:" }} {{counters . | blue}} {{ bar . "[" ("#" | green) ("#" | blue) ("."|white) "]" }} {{percent . | white}} {{speed . }}`
	bar := pb.ProgressBarTemplate(tmpl).Start(filesize)
	barReader := bar.NewProxyReader(r)
	if _, err := io.Copy(out, barReader); err != nil {
		return err
	}
	bar.Finish()

	return nil
}
