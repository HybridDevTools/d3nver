package virtualbox

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
	"strings"
)

// Updater allow us to update a virtualbox box
type Updater struct {
	workingDirectory string
	manifestURL      string
	manifestPath     string
	boxURL           string
	boxPath          string
	storage          storage.Storage
	compressor       compressor.Compressor
	backup           *backup.Backup
	ctx              context.Context
}

// Manifest hold manifest.json info
type Manifest struct {
	Version        string
	Type           string
	ImageSize      int
	FileSize       int
	CompressedSize int
}

// NewUpdater returns a pointer to Updater
func NewUpdater(
	ctx context.Context,
	workingDirectory, manifestURL, manifestPath, boxURL, boxPath string,
	storage storage.Storage,
	compressor compressor.Compressor,
) *Updater {
	return &Updater{
		workingDirectory: workingDirectory,
		manifestURL:      manifestURL,
		manifestPath:     filepath.Join(workingDirectory, manifestPath),
		boxURL:           boxURL,
		boxPath:          filepath.Join(workingDirectory, boxPath),
		storage:          storage,
		compressor:       compressor,
		backup: backup.NewBackup(
			workingDirectory,
			[]string{
				manifestPath,
				boxPath,
			},
		),
		ctx: ctx,
	}
}

// CheckIsUpdated returns a bool with an up-to-date status
func (b *Updater) CheckIsUpdated() (updated bool, err error) {
	exists, err := util.Exists(b.manifestPath)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if !exists {
		return false, nil
	}

	path, err := b.getManifestFilePath()
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll(filepath.Dir(path))
	}()

	go func() {
		<-b.ctx.Done()
		_ = os.RemoveAll(filepath.Dir(path))
	}()

	remoteChecksum, err := util.FileChecksum(path)
	if err != nil {
		return
	}

	localChecksum, err := util.FileChecksum(b.manifestPath)
	if err != nil {
		return
	}
	return strings.Compare(localChecksum, remoteChecksum) == 0, nil
}

// Update box with latest version available
func (b *Updater) Update() (err error) {
	dir, err := ioutil.TempDir("", "box")
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	go func() {
		<-b.ctx.Done()
		_ = os.RemoveAll(dir)
	}()

	log.Println("Downloading manifest...")
	manifestFile, err := b.storage.Download(b.manifestURL, dir)
	if err != nil {
		return
	}

	log.Println("Downloading box...")
	fileName, err := b.storage.Download(b.boxURL, dir)
	if err != nil {
		return
	}

	log.Println("Decompressing...")
	manifest, err := b.getManifest(manifestFile)
	if err != nil {
		return
	}
	err = b.compressor.Decompress(fileName, dir, manifest.FileSize)
	if err != nil {
		return
	}

	err = os.Remove(fileName)
	if err != nil {
		return
	}

	log.Println("Backup current box...")
	if err = b.backup.Rename(); err != nil {
		log.Println("Unexpected error, applying rollback...")
		_ = b.backup.Rollback()
		return
	}

	log.Println("Updating current box...")
	if err = func() (err error) {
		_, file := filepath.Split(b.boxPath)
		originFile := filepath.Join(dir, file)
		err = util.Copy(originFile, b.boxPath)
		if err != nil {
			return
		}

		_, file = filepath.Split(b.manifestPath)
		originFile = filepath.Join(dir, file)
		err = util.Copy(originFile, b.manifestPath)
		if err != nil {
			return
		}

		return
	}(); err != nil {
		log.Println("Unexpected error, applying rollback...")
		_ = b.backup.Rollback()
		return
	}

	return b.backup.Remove()
}

func (b *Updater) getManifestFilePath() (path string, err error) {
	dir, err := ioutil.TempDir("", "manifest")
	if err != nil {
		return
	}

	return b.storage.Download(b.manifestURL, dir)
}

func (b *Updater) getManifest(path string) (manifest Manifest, err error) {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &manifest)
	return
}
