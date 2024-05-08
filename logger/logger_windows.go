//go:build windows
// +build windows

package logger

import (
	"os"
)

func createFile(fileName string) *os.File {
	var fileFd *os.File
	for index := 0; index < 10; index++ {
		if fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm); nil == err {
			fileFd.Sync()
			fileFd.Close()
			fileFd = fd

			break
		}

		fileFd = nil
	}

	return fileFd
}
