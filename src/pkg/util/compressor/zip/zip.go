package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"
)

// Zip implements zip compressor
type Zip struct {
}

// Decompress a zip file
func (z *Zip) Decompress(origin, destination string, filesize int) error {
	var filenames []string

	r, err := zip.OpenReader(origin)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		fpath := filepath.Join(destination, f.Name)

		// https://snyk.io/research/zip-slip-vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			err = fmt.Errorf("%s: illegal file path", fpath)
			return err
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}

			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		tmpl := `{{ green "Progress:" }} {{counters . | blue}} {{ bar . "[" ("#" | green) ("#" | blue) ("."|white) "]" }} {{percent . | white}} {{speed . }}`
		bar := pb.ProgressBarTemplate(tmpl).Start(filesize)
		barReader := bar.NewProxyReader(rc)

		_, err = io.Copy(outFile, barReader)

		bar.Finish()

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return err
}
