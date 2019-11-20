package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//CreateDirectoryWithFiles !TESTING FUNCTION
func CreateDirectoryWithFiles(files []string) (dir string, err error) {
	dir, err = ioutil.TempDir("", "test")
	if err != nil {
		return
	}

	for _, f := range files {
		absPath := filepath.Join(dir, f)

		if err = os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
			return
		}

		emptyFile, err := os.Create(absPath)
		if err != nil {
			return dir, err
		}

		_ = emptyFile.Close()
	}

	return
}

//FilesAreNotPresent !TESTING FUNCTION
func FilesAreNotPresent(dir string, files []string) (err error) {
	return filesAre(dir, files, false)
}

//FilesArePresent !TESTING FUNCTION
func FilesArePresent(dir string, files []string) (err error) {
	return filesAre(dir, files, true)
}

func filesAre(dir string, files []string, present bool) (err error) {
	filesInPath, err := filesInPath(dir)
	for _, file := range files {
		if contains(filesInPath, file) == !present {
			if present {
				return fmt.Errorf("file %s is not present", file)
			}

			return fmt.Errorf("file %s is present", file)
		}
	}

	return
}

func filesInPath(dir string) (content []string, err error) {
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				content = append(content, strings.Replace(path, fmt.Sprintf("%s%s", dir, string(filepath.Separator)), "", -1))
			}

			return nil
		})

	return
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
