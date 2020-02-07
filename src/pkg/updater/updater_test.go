package updater

import (
	"denver/pkg/storage"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	f func(string, string) (string, error)
}

func (f *fakeStorage) Download(origin, destination string) (destFile string, err error) {
	return f.f(origin, destination)
}

func TestFailsIfStorageFails(t *testing.T) {
	assert := assert.New(t)

	updater := &DefaultUpdater{storage: &fakeStorage{
		f: func(origin, destination string) (string, error) {
			return "", fmt.Errorf("this is fine")
		},
	}}
	_, err := updater.getManifestFile()
	assert.EqualError(err, "this is fine")
}

func TestCheckFailsWithInvalidManifest(t *testing.T) {
	assert := assert.New(t)

	updater := &DefaultUpdater{storage: &fakeStorage{
		f: func(origin, destination string) (destFile string, err error) {
			f, err := ioutil.TempFile(destination, fmt.Sprintf("*"))
			if err != nil {
				return
			}
			defer f.Close()
			destFile = f.Name()

			_, err = f.WriteString("ups: this is not a json :D")
			if err != nil {
				return
			}

			return
		},
	}}

	_, err := updater.getManifestFile()
	assert.EqualError(err, "invalid character 'u' looking for beginning of value")
}

func TestCheckFailsWithInvalidVersions(t *testing.T) {
	assert := assert.New(t)

	testcases := []struct {
		manifestVersion, localVersion, failingVersion string
	}{
		{"foo", "1.2.3", "foo"},
		{"1.2.3", "bar", "bar"},
		{"1234foo", "1.2.3", "1234foo"},
		{"1.2.3", "1234bar", "1234bar"},
	}

	for _, testcase := range testcases {
		updater := &DefaultUpdater{storage: getStorage(assert, &Manifest{Date: testcase.manifestVersion})}
		_, _, err := updater.CheckIsUpdated(testcase.localVersion)
		assert.EqualError(err, "improper constraint: >= ")
	}
}

// func TestCheckVersion(t *testing.T) {
// 	assert := assert.New(t)

// 	testcases := []struct {
// 		localVersion   string
// 		manifestVersion string
// 		expectation                   bool
// 	}{
// 		{"1.2.3", "1.2.4", false},
// 		{"1.2.4", "1.2.3", true},
// 		{"1.2.3", "1.2.3", true},
// 	}

// 	for _, testcase := range testcases {
// 		updater := &DefaultUpdater{storage: getStorage(assert, &Manifest{Date: testcase.manifestVersion})}
// 		_, upToDate, err := updater.CheckIsUpdated(testcase.localVersion)
// 		assert.EqualError(err, "improper constraint: >= ")
// 		assert.Equal(testcase.expectation, upToDate, fmt.Sprintf("Local %s should be newer or equal than %s", testcase.localVersion, testcase.manifestVersion))
// 	}
// }

func getStorage(a *assert.Assertions, m *Manifest) storage.Storage {
	b, err := json.Marshal(m)
	a.NoError(err)

	return &fakeStorage{
		f: func(origin, destination string) (destFile string, err error) {
			f, err := ioutil.TempFile(destination, fmt.Sprintf("*"))
			if err != nil {
				return
			}
			defer f.Close()
			destFile = f.Name()

			_, err = f.Write(b)
			if err != nil {
				return
			}

			return
		},
	}
}
