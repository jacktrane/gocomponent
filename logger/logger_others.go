//go:build !windows
// +build !windows

package logger

import (
	"os"
	"syscall"
)

func createFile(fileName string) *os.File {
	var fileFd *os.File
	for index := 0; index < 10; index++ {
		if fd, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm); nil == err {
			fileFd.Sync()
			fileFd.Close()
			fileFd = fd

			// 下面是为了重定向标准输出到文件中，因为painc，Dup2仅能在linux运行哦，所以如果在window下注释
			syscall.Dup2(int(fileFd.Fd()), int(os.Stderr.Fd()))
			break
		}

		fileFd = nil
	}

	return fileFd
}
