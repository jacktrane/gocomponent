package shm

import (
	"errors"
	"fmt"
	"syscall"
)

const (
	IPC_CREATE  = 00001000
	IPC_EXCL    = 00002000
	IPC_NOWAIT  = 00004000
	IPC_DIPC    = 00010000
	IPC_OWN     = 00020000
	IPC_PRIVATE = 0
	IPC_RMID    = 0
	IPC_SET     = 1
	IPC_STAT    = 2
	IPC_INFO    = 3
	IPC_OLD     = 0
	IPC_64      = 0x0100
)

type shm1 struct {
	shmId   uintptr
	shmaddr uintptr
}

func NewShm(key int, size int) (error, *shm1) {
	shmid, _, err := syscall.Syscall(syscall.SYS_SHMGET, uintptr(key), uintptr(size), IPC_CREATE|0600)
	if err != 0 {
		fmt.Printf("syscall error, err: %v\n", err)
		return errors.New(err.Error()), nil
	}
	return nil, &shm1{shmId: shmid}
}

func (s *shm1) Attach() (error, uintptr) {
	var (
		err syscall.Errno
	)
	s.shmaddr, _, err = syscall.Syscall(syscall.SYS_SHMAT, s.shmId, 0, 0)
	if err != 0 {
		return errors.New(err.Error()), 0
	}
	return nil, s.shmaddr
}

func (s *shm1) Close() error {
	_, _, err := syscall.Syscall(syscall.SYS_SHMDT, s.shmaddr, 0, 0)
	if err != 0 {
		return errors.New(err.Error())
	}
	return nil
}
