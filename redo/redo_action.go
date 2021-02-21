package redo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jacktrane/gocomponent/time_format"
)

type StatMachine interface {
	Run() error // 执行
	// Dump() []byte            // 执行失败时将数据dump下来
	// Load(paras []byte) error // 重新load数据执行
}

type RedoConfig struct {
	RedoFileNameWithPath string // 重试文件所处目录
	SliceFileInterval    int    // 多久切割文件，s为单位
	HoldFileNum          int    // 最大文件数量
	FileFormat           string // 文件格式
	LineLimit            int    // 文件最大行数
	PollInterval         int    // 轮询间隔
	PollRateLimit        int    // 轮询速度
	Machine              StatMachine
}

func (r RedoConfig) defaultParam() {
	// TODO 不填则放置在内存中，但这里有个问题，会不会把内存打爆了？
	// if r.RedoFileNameWithPath == "" {
	// 	r.RedoFileNameWithPath = "../log/data"
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
	logDateFormat string
	nowDate       string
	fileSliceOne  *sync.Once
	locker        *sync.RWMutex
	exitFlag      int
}

const (
	ExitSignle  = 1 // 信号退出
	ExitLogFull = 2 // log爆炸退出
)

func NewRedoActionConf(conf RedoConfig) *redoAction {
	conf.defaultParam()

	ra := redoAction{
		conf: conf,
	}

	// 保证machine一定有并且需要具备指定的状态码
	if conf.Machine == nil {
		log.Fatal("machine 不存在")
	}

	// 强行要加一个状态码
	machineValue := reflect.ValueOf(conf.Machine).FieldByName("StatCode")
	if !machineValue.IsValid() {
		log.Fatal("machine 不存在 StatCode 字段")
	}
	if machineValue.Type().Name() != "int" {
		log.Fatal("machine中StatCode字段类型不为int")
	}

	// 切割文件目前支持最低的切割粒度是小时级别
	ra.logDateFormat = time_format.FullFormatDateSimpleDay
	if ra.conf.SliceFileInterval == time_format.OneHour {
		ra.logDateFormat += "_15"
	}

	if err := ra.initFile(); err != nil {
		log.Fatal(err)
	}
	ra.setMachine()
	return &ra
}

// 保证稳定执行
func (r *redoAction) StableAction(machine StatMachine) {
	err := machine.Run()
	if err != nil { // 错误不反悔
		log.Printf("RunErr=%s failLine=%s\n", err, machine)
		err, strDumpData := r.dump(machine)
		if err != nil {
			log.Printf("dump=%s redo\n", strDumpData)
		}
		_, err = r.failFile.WriteString(strDumpData)
		if err != nil {
			log.Printf("WriteString=%s redo\n", strDumpData)
		}
	}
}

func (r *redoAction) initFile() error {
	r.locker.Lock()
	defer r.locker.Unlock()

	// 打开文件fd
	if r.conf.RedoFileNameWithPath != "" {
		if err := os.MkdirAll(path.Dir(r.conf.RedoFileNameWithPath), os.ModePerm); err != nil {
			log.Printf("mkdir %s err=%s\n", path.Dir(r.conf.RedoFileNameWithPath), err)
			return err
		}
	}

	now := time.Now()
	r.nowDate = now.Format(r.logDateFormat)
	var err error
	err, r.succFile, r.failFile = r.getLogFile(r.nowDate)
	if err != nil {
		log.Printf("getLogFile nowDate=%s err=%s\n", r.nowDate, err)
		return err
	}

	// 获取上一天的日志信息，把还没有状态机的重试的进行重试
	d, _ := time.ParseDuration("-24h")
	beforeDate := now.Add(d).Format(r.logDateFormat)
	err, beforeSuccFile, beforeFailFile := r.getLogFile(beforeDate)
	if err != nil {
		log.Printf("getLogFile nowDate=%s err=%s\n", beforeDate, err)
		return err
	}

	err, diffLines, _ := r.diff(beforeSuccFile, beforeFailFile)
	if err != nil {
		log.Printf("diff err=%s\n", err)
		return err
	}

	// 写入新日志
	if len(diffLines) != 0 {
		failLines, err := ioutil.ReadFile(r.failFile.Name())
		if err != nil {
			log.Printf("ReadFile failFile=%s err=%s\n", r.failFile.Name(), err)
			return err
		}
		arrFailLine := strings.Split(string(failLines), "\n")
		mapFailLine := make(map[string]bool)
		for _, failLine := range arrFailLine {
			mapFailLine[failLine] = true
		}

		for _, diffLine := range diffLines {
			if !mapFailLine[diffLine] {
				_, errLine := r.failFile.WriteString(diffLine + "\n")
				if errLine != nil {
					log.Printf("writeString err=%s str=%s \n", errLine, diffLine)
				}
			} else {
				log.Printf("existed str=%s\n", diffLine)
			}
		}
	}

	r.fileSliceOne.Do(func() {
		go r.checkFileSliceCond()
	})

	// 保存起来，别每次都打开，然后切换日期
	return nil
}

// 将不一样的拿出来
func (r *redoAction) diff(succFile, failFile *os.File) (error, []string, int) {
	succLines, err := ioutil.ReadFile(succFile.Name())
	if err != nil {
		return err, nil, 0
	}

	failLines, err := ioutil.ReadFile(failFile.Name())
	if err != nil {
		return err, nil, 0
	}

	arrSuccLine := strings.Split(string(succLines), "\n")
	arrFailLine := strings.Split(string(failLines), "\n")
	mapSuccLine := make(map[string]bool)
	for _, succLine := range arrSuccLine {
		mapSuccLine[succLine] = true
	}

	arrDiffLine := make([]string, 0, len(arrFailLine)/3)
	for _, failLine := range arrFailLine {
		if !mapSuccLine[failLine] {
			arrDiffLine = append(arrDiffLine, failLine)
		}
	}

	return nil, arrDiffLine, len(arrFailLine)
}

func (r *redoAction) formatLogFile(bSuccFile bool, dateStr string) (error, *os.File) {
	status := "fail"
	if bSuccFile {
		status = "succ"
	}

	fileName := fmt.Sprintf("%s_%s_%s.log", r.conf.RedoFileNameWithPath, r.nowDate, status)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err, nil
	}
	log.Printf("fileName=%s succOpen", fileName)
	// if bSuccFile {
	// 	r.succFile = file
	// } else {
	// 	r.failFile = file
	// }
	return nil, file
}

func (r *redoAction) getLogFile(dateStr string) (error, *os.File, *os.File) {
	err, failFile := r.formatLogFile(false, dateStr)
	if err != nil {
		return err, nil, nil
	}

	err, succFile := r.formatLogFile(false, dateStr)
	if err != nil {
		failFile.Close()
		return err, nil, nil
	}

	return nil, succFile, failFile
}

func (r *redoAction) checkFileSliceCond() {
	for {
		// TODO 清除多余日志
		time.Sleep(time.Minute)
		if r.nowDate != time.Now().Format(r.logDateFormat) {
			// 抛出个panic让业务去知道吧
			if err := r.initFile(); err != nil {
				panic(err)
			}
		}
	}
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
		if r.exitFlag != 0 {
			log.Printf("exit sign=%d", r.exitFlag)
			return
		}

		time.Sleep(time.Second * time.Duration(r.conf.PollInterval))
		r.redo()
	}
}

// 从内存还是哪里去redo
func (r *redoAction) redo() {
	r.locker.Lock()
	defer r.locker.Unlock()

	err, arrFailLine, failLineNum := r.diff(r.succFile, r.failFile)
	if err != nil {
		log.Printf("diff err=%s\n", err)
		return
	}

	for _, failLine := range arrFailLine {
		// 执行
		err, statMachine := r.load(failLine)
		if err != nil { // 错误不反悔
			log.Printf("err=%s failLine=%s\n", err, failLine)
			_, err = r.succFile.WriteString(failLine + "\n")
			if err != nil {
				log.Printf("WriteString=%s redo\n", failLine)
			}
		}

		err = statMachine.Run()
		if err != nil { // 错误不反悔
			log.Printf("RunErr=%s failLine=%s\n", err, failLine)

			// 查看状态码是否改变了
			// 改变则原先执行成功塞入succlog原先的成功，塞入faillog新的失败

			continue
		}

		// 执行成功
		_, err = r.succFile.WriteString(failLine + "\n")
		if err != nil {
			log.Printf("WriteString=%s redo\n", failLine)
		}

	}

	// 查看失败行数是否超过了限制，等失败的重试成功之后再执行
	if failLineNum > r.conf.LineLimit {
		log.Printf("fail line full")
		r.exitFlag = r.exitFlag ^ ExitLogFull
		return
	}
}

const (
	logFormatDate = time_format.Year + time_format.Mon + time_format.Day + time_format.Hour + time_format.Min + time_format.Sec + ".999"
)

func (r *redoAction) dump(machine StatMachine) (error, string) {
	byteJson, err := json.Marshal(machine)
	if err != nil {
		return err, ""
	}

	strJson := string(byteJson)
	strJson = strings.Replace(strJson, "\\", "\\\\", -1)
	strJson = strings.Replace(strJson, "\n", "\\\n", -1)
	failTime := time.Now().Format(logFormatDate)
	for len(failTime) < len(logFormatDate) {
		failTime += "0"
	}

	return nil, failTime + "|" + strJson + "\n"
}

func (r *redoAction) load(para string) (error, StatMachine) {
	para = para[len(logFormatDate)+1:]
	para = strings.Replace(para, "\\\n", "\n", -1)
	para = strings.Replace(para, "\\\\", "\\", -1)

	tmpMachine := r.conf.Machine
	err := json.Unmarshal([]byte(para), &tmpMachine)
	return err, tmpMachine
}
