package updater

import (
	"context"
	"denver/pkg/backup"
	"denver/pkg/storage"
	"denver/pkg/util"
	"denver/pkg/util/compressor"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// Updater represents an updater mechanism
type Updater interface {
	CheckIsUpdated(string) (Manifest, bool, error)
	Update() error
}

// DefaultUpdater struct allow us to check for updates for our application
type DefaultUpdater struct {
	workingDirectory string
	manifestURL      string
	compressor       compressor.Compressor
	storage          storage.Storage
	backup           *backup.Backup
	ctx              context.Context
}

// Manifest stores manifestURL content
type Manifest struct {
	Date, Release, URL, FileSize string
}

// NewDefaultUpdater returns a pointer to DefaultUpdater
func NewDefaultUpdater(
	ctx context.Context,
	workingDirectory, manifestURL string,
	filesToBackup []string,
	compressor compressor.Compressor,
	storage storage.Storage,
) *DefaultUpdater {
	return &DefaultUpdater{
		workingDirectory: workingDirectory,
		manifestURL:      manifestURL,
		compressor:       compressor,
		storage:          storage,
		backup:           backup.NewBackup(workingDirectory, filesToBackup),
		ctx:              ctx,
	}
}

// CheckIsUpdated if local version is up to date with manifestURL
func (u *DefaultUpdater) CheckIsUpdated(localVersion string) (manifest Manifest, updated bool, err error) {
	updated = false
	manifest, err = u.getManifestFile()
	if err != nil {
		return
	}

	var l, m int
	l, err = strconv.Atoi(localVersion)
	if err != nil {
		return
	}

	m, err = strconv.Atoi(manifest.Date)
	if err != nil {
		return
	}

	updated = l >= m
	return
}

// Update current application
func (u *DefaultUpdater) Update() (err error) {
	var manifest Manifest
	manifest, err = u.getManifestFile()
	if err != nil {
		return
	}

	dir, err := ioutil.TempDir("", "updater")
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	go func() {
		<-u.ctx.Done()
		_ = os.RemoveAll(dir)
	}()

	log.Printf("Downloading release %s...", manifest.Release)
	fileName, err := u.storage.Download(manifest.URL, dir)
	if err != nil {
		return
	}

	log.Println("Decompressing...")
	fileSize, err := strconv.Atoi(manifest.FileSize)
	if err != nil {
		return
	}

	err = u.compressor.Decompress(fileName, dir, fileSize)
	if err != nil {
		return
	}

	err = os.Remove(fileName)
	if err != nil {
		return
	}

	// Here we will checksum the binary.

	log.Println("Backup current release...")
	if err = u.backup.Rename(); err != nil {
		log.Println("Unexpected error, applying rollback...")
		_ = u.backup.Rollback()
		return
	}

	log.Println("Updating current release...")
	//TODO: zip contains everything inside a denver directory
	if err = util.CopyRecursive(filepath.Join(dir, "denver"), u.workingDirectory); err != nil {
		log.Println("Unexpected error, applying rollback...")
		_ = u.backup.Rollback()
		return
	}

	return u.backup.Remove()
}

func (u *DefaultUpdater) getManifestFile() (manifest Manifest, err error) {
	dir, err := ioutil.TempDir("", "manifest")
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	file, err := u.storage.Download(u.manifestURL, dir)
	if err != nil {
		return
	}

	body, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &manifest)
	return
}
