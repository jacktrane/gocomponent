package file_util

import (
	"io"
	"os"
	"time"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func GetFileModTime(path string) (error, *time.Time) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err, nil
	}
	tTime := fileInfo.ModTime()
	return nil, &tTime
}

func GetFileContent(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
