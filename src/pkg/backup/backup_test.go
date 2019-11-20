package backup

import (
	"denver/test"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackupsWhiteListedFiles(t *testing.T) {
	assert := assert.New(t)
	dir, err := test.CreateDirectoryWithFiles(
		[]string{
			"file-1",
			"file-2",
			"conf/file-3",
			"conf.old/file-5",
			"file-4.old",
			".hidden-file",
		},
	)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	assert.NoError(err)

	updater := &Backup{
		workingDirectory: dir,
		filesToBackup: []string{
			"file-1",
			"file-4",
			"conf",
			".hidden-file",
		},
	}

	err = updater.Rename()
	assert.NoError(err)

	err = test.FilesArePresent(
		dir,
		[]string{
			"file-1.old",
			"conf.old/file-3",
			".hidden-file.old",
		},
	)
	assert.NoError(err)
}

func TestRollbackImportantFiles(t *testing.T) {
	assert := assert.New(t)
	dir, err := test.CreateDirectoryWithFiles(
		[]string{
			"file-1.old",
			"file-2",
			"conf/file-3",
			"conf.old/file-4",
		},
	)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	assert.NoError(err)

	updater := &Backup{
		workingDirectory: dir,
		filesToBackup: []string{
			"file-1",
			"file-2",
			"conf",
		},
	}

	err = updater.Rollback()
	assert.NoError(err)

	err = test.FilesArePresent(
		dir,
		[]string{
			"file-1",
			"file-2",
			"conf/file-4",
		},
	)
	assert.NoError(err)
}

func TestDeletesBackupFiles(t *testing.T) {
	assert := assert.New(t)
	dir, err := test.CreateDirectoryWithFiles(
		[]string{
			"file-1.old",
			"file-2.old",
			"conf.old/file-3",
		},
	)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	assert.NoError(err)

	updater := &Backup{
		workingDirectory: dir,
		filesToBackup: []string{
			"file-1",
			"conf",
		},
	}

	err = updater.Remove()
	assert.NoError(err)

	err = test.FilesAreNotPresent(
		dir,
		[]string{
			"file-1.old",
			"conf.old/file-3",
		},
	)
	assert.NoError(err)

	err = test.FilesArePresent(
		dir,
		[]string{
			"file-2.old",
		},
	)
	assert.NoError(err)
}
