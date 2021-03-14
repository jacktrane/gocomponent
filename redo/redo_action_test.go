package redo

import (
	"errors"
	"path"
	"testing"
	"time"

	"github.com/jacktrane/gocomponent/time_format"
)

type TestMachine struct {
	StatCode int `jso1n:"statcode"`
	Num      int
	Str      string
}

const (
	MachineStatCodeTest1 = iota
	MachineStatCodeTest2
	MachineStatCodeTest3
	MachineStatCodeTest4
)

func (t *TestMachine) Run() (bool, error) {
	statCode := t.StatCode
	for {
		switch t.StatCode {
		case MachineStatCodeTest1:
			if t.Num < 100 {
				t.Num++
				// logger.Errorf("%+v", t)
				return true, errors.New("errrr") // 数据 是否改变
			}

			t.StatCode = MachineStatCodeTest2
		case MachineStatCodeTest2:
			if t.Str != "a" {
				t.Str = "a"
				return statCode != t.StatCode, errors.New("errrr") // 状态 是否改变
			}

			t.StatCode = MachineStatCodeTest3
		case MachineStatCodeTest3:
			return false, nil
		}
	}
}

func TestRedo(t *testing.T) {
	var emptyMachine TestMachine
	conf := RedoConfig{
		RedoFileNameWithPath: path.Join(path.Join("..", "runtime", "log", "redo.log")),
		SliceFileInterval:    time_format.OneDay,
		Machine:              &emptyMachine,
		PollInterval:         1,
		PollRateLimit:        2,
	}
	redo := NewRedoActionConf(conf)

	var testMachine TestMachine
	redo.StableAction(&testMachine)
	time.Sleep(10 * time.Second)
}
