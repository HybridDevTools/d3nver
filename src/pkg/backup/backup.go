package backup

import (
	"fmt"
	"os"
	"path/filepath"
)

// Backup allow us to backup list of files
type Backup struct {
	workingDirectory string
	filesToBackup    []string
}

// NewBackup returns a pointer to Backup
func NewBackup(workingDirectory string, filesToBackup []string) *Backup {
	return &Backup{
		workingDirectory: workingDirectory,
		filesToBackup:    filesToBackup,
	}
}

// Rename all files in filesToBackup
func (b *Backup) Rename() (err error) {
	for _, element := range b.filesToBackup {
		absPath := filepath.Join(b.workingDirectory, element)
		oldElement := fmt.Sprintf("%s.old", absPath)
		if _, err := os.Stat(absPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err = os.RemoveAll(oldElement); err != nil && !os.IsNotExist(err) {
			return
		}
		if err = os.Rename(absPath, oldElement); err != nil {
			return
		}
	}

	return
}

// Rollback files to a previous version
func (b *Backup) Rollback() (err error) {
	for _, element := range b.filesToBackup {
		absPath := filepath.Join(b.workingDirectory, element)
		oldElement := fmt.Sprintf("%s.old", absPath)
		if _, err := os.Stat(oldElement); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err = os.RemoveAll(absPath); err != nil && !os.IsNotExist(err) {
			return
		}
		if err = os.Rename(oldElement, absPath); err != nil {
			return
		}
	}

	return
}

// Remove backup files
func (b *Backup) Remove() (err error) {
	for _, element := range b.filesToBackup {
		absPath := filepath.Join(b.workingDirectory, element)
		oldElement := fmt.Sprintf("%s.old", absPath)
		if _, err := os.Stat(oldElement); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err = os.RemoveAll(oldElement); err != nil {
			return
		}
	}

	return
}
