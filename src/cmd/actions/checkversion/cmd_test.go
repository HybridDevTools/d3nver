package checkversion

import (
	"denver/pkg/notify"
	"denver/pkg/providers"
	"denver/pkg/updater"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testingVM struct {
	providers.Testing
}

func getVMProvider(provider interface{}) providers.VMProvider {
	return provider.(providers.VMProvider)
}

type mockUpdater struct {
	updater.DefaultUpdater
	manifest updater.Manifest
	updated  bool
	err      error
	notify   notify.Notify
}

func (t *mockUpdater) CheckIsUpdated(localVersion string) (manifest updater.Manifest, updated bool, err error) {
	return t.manifest, t.updated, t.err
}

type testWriter struct {
	f func(s string)
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.f(string(p))
	return len(p), nil
}

func TestCommandFailsIfUpdaterFails(t *testing.T) {
	assert := assert.New(t)
	vm := getVMProvider(&testingVM{})

	cmd := NewCheckVersion(&vm, &mockUpdater{err: fmt.Errorf("this is fine")}, (*log.Logger)(nil), notify.CliQuestion{})

	err := cmd.GetCommand().Exec()
	assert.EqualError(err, "this is fine")
}

func TestCommandReturnsAPositiveMessageIfIsUpdated(t *testing.T) {
	assert := assert.New(t)
	vm := getVMProvider(&testingVM{})
	writer := &testWriter{f: func(s string) {
		assert.True(regexp.MustCompile(`Your version [0-9]+ is up to date`).MatchString(s))
	}}

	cmd := NewCheckVersion(&vm, &mockUpdater{updated: true}, log.New(writer, "", 0), notify.CliQuestion{})

	err := cmd.GetCommand().Exec()
	assert.NoError(err)
}

func TestCommandReturnsANegativeMessageIfIsNotUpdated(t *testing.T) {
	assert := assert.New(t)
	vm := getVMProvider(&testingVM{})
	writer := &testWriter{f: func(s string) {
		assert.True(regexp.MustCompile(`Your version [0-9]+ is outdated`).MatchString(s))
	}}

	cmd := NewCheckVersion(&vm, &mockUpdater{updated: false}, log.New(writer, "", 0), notify.CliQuestion{})

	err := cmd.GetCommand().Exec()
	assert.NoError(err)
}
