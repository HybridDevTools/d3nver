package util

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Exists : Check if a file exists
func Exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CopyRecursive from origin to destination
func CopyRecursive(origin, destination string) error {
	return filepath.Walk(origin,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			relPath := strings.Replace(path, origin, "", -1)
			if err := Copy(path, filepath.Join(destination, relPath)); err != nil && !os.IsExist(err) {
				return err
			}

			return nil
		})
}

// Copy a single file from origin to destination
func Copy(origin, destination string) error {
	s, err := os.Open(origin)
	if err != nil {
		return err
	}
	defer s.Close()

	if _, err = os.Stat(destination); err == nil {
		return os.ErrExist
	}

	if err = os.MkdirAll(filepath.Dir(destination), os.ModePerm); err != nil {
		return err
	}

	d, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return err
	}
	defer d.Close()

	buf := make([]byte, 1000)
	for {
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := d.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

// FileChecksum gives the checksum of a given file
func FileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	checksum := sha256.Sum256(c)
	return hex.EncodeToString(checksum[:]), nil
}
