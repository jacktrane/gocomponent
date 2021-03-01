package redo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jacktrane/gocomponent/logger"
	"github.com/jacktrane/gocomponent/time_format"
)

type StatMachine interface {
	Run() (bool, error) // 执行 bool代表是否状态改变， error代表有错误需要执行
}

type RedoConfig struct {
	RedoFileNameWithPath string // 重试文件所处目录
	SliceFileInterval    int    // 多久切割文件， 仅允许 time_format.OneDay | time_format.OneHour
	HoldFileNum          int    // 最大文件数量
	LineLimit            int    // 文件最大行数
	PollInterval         int    // 轮询间隔
	PollRateLimit        int    // 轮询速度
	Machine              StatMachine
}

func (r *RedoConfig) defaultParam() {
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
	ra := redoAction{
		conf:          conf,
		locker:        &sync.RWMutex{},
		setMachineOne: &sync.Once{},
		fileSliceOne:  &sync.Once{},
	}
	ra.conf.defaultParam()

	// 保证machine一定有并且需要具备指定的状态码
	if conf.Machine == nil {
		logger.Fatal("machine 不存在")
	}

	// 强行要加一个状态码
	// machineReflect := reflect.ValueOf(conf.Machine)
	// if machineReflect.Elem().Kind() != reflect.Struct {
	// 	logger.Panic("machine 不是一个结构体")
	// }
	// statCodeField := machineReflect.Elem().FieldByName("StatCode")
	// if !statCodeField.IsValid() {
	// 	logger.Fatal("machine 不存在 StatCode 字段")
	// }
	// if statCodeField.Type().Name() != "int" {
	// 	logger.Fatal("machine中StatCode字段类型不为int")
	// }

	if ra.conf.SliceFileInterval != time_format.OneDay && ra.conf.SliceFileInterval != time_format.OneHour {
		logger.Fatalf("conf中的SliceFileInterval仅允许传%d或%d", time_format.OneDay, time_format.OneHour)
	}

	// // 限制结构体不允许有json的tag
	// machineReflectType := machineReflect.Type()
	// fieldNum := machineReflect.NumField()
	// for index := 0; index < fieldNum; index++ {
	// 	if _, existed := machineReflectType.Field(index).Tag.Lookup("json"); existed {
	// 		logger.Fatalf("字段名：%s 不允许加json tag", machineReflectType.Field(index).Name)
	// 	}
	// }

	// 切割文件目前支持最低的切割粒度是小时级别
	ra.logDateFormat = time_format.FullFormatDateSimpleDay
	if ra.conf.SliceFileInterval == time_format.OneHour {
		ra.logDateFormat += "_15"
	}

	if err := ra.initFile(); err != nil {
		logger.Fatal(err)
	}
	ra.setMachine()
	return &ra
}

// 保证稳定执行
func (r *redoAction) StableAction(machine StatMachine) {
	if _, err := machine.Run(); err != nil {
		errDump, strDumpData := r.dump(machine)
		logger.Warnf("runErr=%s failLine=%s redo", err, strDumpData)
		if errDump != nil {
			logger.Warnf("err=%s dump=%s redo", errDump, strDumpData)
		}
		_, err = r.failFile.WriteString(strDumpData)
		if err != nil {
			logger.Warnf("WriteString=%s redo\n", strDumpData)
		}
	}
}

func (r *redoAction) initFile() error {
	r.locker.Lock()
	defer r.locker.Unlock()

	// 打开文件fd
	if r.conf.RedoFileNameWithPath != "" {
		if err := os.MkdirAll(path.Dir(r.conf.RedoFileNameWithPath), os.ModePerm); err != nil {
			logger.Errorf("mkdir %s err=%s\n", path.Dir(r.conf.RedoFileNameWithPath), err)
			return err
		}
	}

	now := time.Now()
	r.nowDate = now.Format(r.logDateFormat)
	var err error
	err, r.succFile, r.failFile = r.getLogFile(r.nowDate)
	if err != nil {
		logger.Errorf("getLogFile nowDate=%s err=%s\n", r.nowDate, err)
		return err
	}

	// 获取上一个时间段的日志信息，把还没有状态机的重试的进行重试
	d, _ := time.ParseDuration("-24h")
	if r.conf.SliceFileInterval == time_format.OneHour {
		d, _ = time.ParseDuration("-1h")
	}

	beforeDate := now.Add(d).Format(r.logDateFormat)
	err, beforeSuccFile, beforeFailFile := r.getLogFile(beforeDate)
	if err != nil {
		logger.Errorf("getLogFile nowDate=%s err=%s\n", beforeDate, err)
		return err
	}

	err, diffLines, _ := r.diff(beforeSuccFile, beforeFailFile)
	if err != nil {
		logger.Errorf("diff err=%s\n", err)
		return err
	}

	// 写入新日志
	if len(diffLines) != 0 {
		failLines, err := ioutil.ReadFile(r.failFile.Name())
		if err != nil {
			logger.Errorf("ReadFile failFile=%s err=%s\n", r.failFile.Name(), err)
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
					logger.Warnf("writeString err=%s str=%s \n", errLine, diffLine)
				}
			} else {
				logger.Warnf("existed str=%s\n", diffLine)
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

	fileName := fmt.Sprintf("%s_%s_%s.log", r.conf.RedoFileNameWithPath, dateStr, status)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err, nil
	}
	logger.Infof("fileName=%s succOpen", fileName)
	return nil, file
}

func (r *redoAction) getLogFile(dateStr string) (error, *os.File, *os.File) {
	err, failFile := r.formatLogFile(false, dateStr)
	if err != nil {
		return err, nil, nil
	}

	err, succFile := r.formatLogFile(true, dateStr)
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
			logger.Errorf("exit sign=%d", r.exitFlag)
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
		logger.Errorf("diff err=%s\n", err)
		return
	}

	// TODO 这里可以做一下限流，防止挂掉
	for _, failLine := range arrFailLine {
		// 执行
		err, statMachine := r.load(failLine)
		if err != nil {
			logger.Errorf("致命错误，但不能阻碍重试 err=%s failLine=%s\n", err, failLine)
			_, err = r.succFile.WriteString(failLine + "\n")
			if err != nil {
				logger.Warnf("WriteString=%s redo\n", failLine)
			}
			continue
		}

		statChange, err := statMachine.Run()
		if err != nil { // 错误不返回
			logger.Warnf("RunErr=%s failLine=%s\n", err, failLine)

			if statChange {
				_, err = r.succFile.WriteString(failLine + "\n")
				if err != nil {
					logger.Warnf("WriteString=%s redo\n", failLine)
				}

				err, strDumpData := r.dump(statMachine)
				if err != nil {
					logger.Warnf("dump=%s redo\n", strDumpData)
				}
				_, err = r.failFile.WriteString(strDumpData)
				if err != nil {
					logger.Warnf("WriteString=%s redo\n", strDumpData)
				}
			}

			continue
		}

		// 执行成功
		_, err = r.succFile.WriteString(failLine + "\n")
		if err != nil {
			logger.Warnf("WriteString=%s redo\n", failLine)
		}

	}

	// 查看失败行数是否超过了限制，等失败的重试成功之后再执行
	if failLineNum > r.conf.LineLimit {
		logger.Errorf("fail line full，failLineNum=%d confLineLimit=%d", failLineNum, r.conf.LineLimit)
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

	// machineT := reflect.ValueOf(r.conf.Machine).Type()
	// machineReflect := reflect.New(machineT)
	// err := json.Unmarshal([]byte(para), machineReflect.Interface())
	// if err != nil {
	// 	return err, nil
	// }
	// return nil, machineReflect.Interface().(StatMachine)
	p := reflect.ValueOf(r.conf.Machine).Elem()
	p.Set(reflect.Zero(p.Type()))
	err := json.Unmarshal([]byte(para), r.conf.Machine)
	if err != nil {
		return err, nil
	}
	return nil, r.conf.Machine
}
