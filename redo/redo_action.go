package redo

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"time_format"
)

type StatMachine interface {
	Run() (bool, error) // 执行
	// Dump() []byte            // 执行失败时将数据dump下来
	// Load(paras []byte) error // 重新load数据执行
}

type RedoConfig struct {
	RedoFilePath      string // 重试文件所处目录
	SliceFileInterval int    // 多久切割文件，s为单位
	HoldFileNum       int    // 最大文件数量
	FileFormat        string // 文件格式
	LineLimit         int    // 文件最大行数
	PollInterval      int    // 轮询间隔
	PollRateLimit     int    // 轮询速度
	Machine           StatMachine
}

func (r RedoConfig) defaultParam() {
	// TODO 不填则放置在内存中，但这里有个问题，会不会把内存打爆了？
	// if r.RedoFilePath == "" {
	// 	r.RedoFilePath = "../log/data"
	// }
	if r.SliceFileInterval == 0 {
		r.SliceFileInterval = time_format.OneDay
	}
	if r.HoldFileNum == 0 {
		r.HoldFileNum = 10
	}
	if r.LineLimit == 0 {
		r.LineLimit = 10000000
	}
	if r.PollInterval == 0 {
		r.PollInterval = 5
	}
	if r.PollRateLimit == 0 {
		r.PollRateLimit = 3000
	}
}

// 执行者
type redoAction struct {
	conf          RedoConfig
	setMachineOne *sync.Once
	succFile      *os.File
	failFile      *os.File
}

func NewRedoActionConf(conf RedoConfig) *redoAction {
	conf.defaultParam()

	ra := redoAction{
		conf: conf,
	}

	if err := ra.initFile(); err != nil {
		log.Fatal()
	}
	ra.setMachine()
	return &ra
}

// 保证稳定执行
func (r *redoAction) StableAction(machine StatMachine) {

}

func (r *redoAction) initFile() error {
	// 打开文件fd
	if r.conf.RedoFilePath != "" {
		os.MkdirAll(r.conf.RedoFilePath, os.ModePerm)
	}

	// 保存起来，别每次都打开，然后切换日期
	return nil
}

func (r *redoAction) formatLogFile(bFailFile bool) {

}

func (r *redoAction) setMachine() {
	r.setMachineOne.Do(
		func() {
			go r.retryPoll()
		},
	)
}

func (r *redoAction) retryPoll() {
	for {
		time.Sleep(time.Second * time.Duration(r.conf.PollInterval))
		r.redo()
	}
}

// 从内存还是哪里去redo
func (r *redoAction) redo() {

}

func (r *redoAction) dump(machine StatMachine) ([]byte, error) {
	return json.Marshal(machine)
}

func (r *redoAction) load(paras []byte) (error, StatMachine) {
	tmpMachine := r.conf.Machine
	json.Unmarshal(paras, &tmpMachine)
	return nil, tmpMachine
}
