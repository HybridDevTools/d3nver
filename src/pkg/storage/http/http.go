package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cheggaaa/pb/v3"
)

// HTTP storage implementation
type HTTP struct{}

// Download a document and place it in destination
func (h *HTTP) Download(origin, destination string) (destFile string, err error) {

	resp, err := http.Get(origin)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		err = fmt.Errorf("cannot determine progress without Content-Length")
		return
	}
	size, err := strconv.ParseInt(contentLengthHeader, 10, 64)
	if err != nil {
		err = fmt.Errorf("bad Content-Length %q", contentLengthHeader)
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("file not found")
		return
	}

	_, file := filepath.Split(origin)
	destFile = filepath.Join(destination, file)

	f, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer f.Close()

	tmpl := `{{ green "Progress:" }} {{counters . | blue}} {{ bar . "[" ("#" | green) ("#" | blue) ("."|white) "]" }} {{percent . | white}} {{speed . }}`
	bar := pb.ProgressBarTemplate(tmpl).Start64(size)
	barReader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(f, barReader)
	bar.Finish()

	return
}
