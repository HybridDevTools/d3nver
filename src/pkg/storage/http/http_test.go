package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFailsIfFileIsNotAccessible(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	h := &HTTP{}
	_, err := h.Download(ts.URL, "")
	assert.EqualError(err, "file not found")
}

func TestDownloadsFileToSpecificLocation(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "This is a downloaded document")
	}))
	defer ts.Close()

	dir, err := ioutil.TempDir("", "test")
	assert.NoError(err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	h := &HTTP{}
	path, err := h.Download(ts.URL, dir)

	b, err := ioutil.ReadFile(path)
	assert.NoError(err)

	assert.Equal("This is a downloaded document", string(b))
}
