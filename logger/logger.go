package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jacktrane/gocomponent/basic"
	"github.com/jacktrane/gocomponent/file_util"
	"github.com/jacktrane/gocomponent/time_format"
)

const (
	PanicLevel int = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

// 等级最高 则数值最低
var printLevelArr = map[int]string{
	0: "[Panic] ",
	1: "[Fatal] ",
	2: "[Error] ",
	3: "[Warn] ",
	4: "[Info] ",
	5: "[Debug] ",
}

type LogFile struct {
	level        int
	logTime      int64
	fileName     string
	fileFd       *os.File
	holdLogSum   int
	clearLogOnce *sync.Once
}

var gLogFile LogFile

func init() {
	NewConfig("", 5)
}

func NewConfig(logFolder string, level int) {
	gLogFile.fileName = logFolder
	gLogFile.level = level
	gLogFile.holdLogSum = 2
	gLogFile.clearLogOnce = &sync.Once{}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	if logFolder != "" {
		log.SetOutput(io.MultiWriter(os.Stdout, gLogFile))
		gLogFile.fileFd = createFile(gLogFile.fileName) // 在初始化时先加个fd先
		gLogFile.clearLogOnce.Do(func() {
			gLogFile.clearLog()
		})
	}
}

func SetLevel(level int) {
	gLogFile.level = level
}

func SetHoldLogSum(holdLogSum int) {
	gLogFile.holdLogSum = holdLogSum
}

func Debugf(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Debug] ")
		// output(fmt.Sprintf(format, args...))
		output(fmt.Sprintf(format, args...))
	}
}

func Debug(v ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Debug] ")
		output(fmt.Sprint(v...))
	}
}

func Infof(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Info] ")
		output(fmt.Sprintf(format, args...))
	}
}

func Info(v ...interface{}) {
	if gLogFile.level >= InfoLevel {
		log.SetPrefix("[Info] ")
		output(fmt.Sprint(v...))
	}
}

func Warnf(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Warn] ")
		output(fmt.Sprintf(format, args...))
	}
}

func Warn(v ...interface{}) {
	if gLogFile.level >= WarnLevel {
		log.SetPrefix("[Warn] ")
		output(fmt.Sprint(v...))
	}
}

func Errorf(format string, args ...interface{}) {
	if gLogFile.level >= DebugLevel {
		log.SetPrefix("[Error] ")
		output(fmt.Sprintf(format, args...))
	}
}

func Error(v ...interface{}) {
	if gLogFile.level >= ErrorLevel {
		log.SetPrefix("[Error] ")
		output(fmt.Sprint(v...))
	}
}

func Fatalf(format string, args ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Fatal] ")
		output(fmt.Sprintf(format, args...))
		debug.PrintStack()
		os.Exit(1)
	}
}

func Fatal(v ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Fatal] ")
		output(fmt.Sprint(v...))
		debug.PrintStack()
		os.Exit(1)
	}
}

func output(v ...interface{}) {
	log.Output(3, "["+formatFuncPrefix()+"] "+fmt.Sprint(v...))
}

func formatFuncPrefix() string {
	funcName, _, _, _ := runtime.Caller(3)
	funcNameFullPatch := strings.Split(runtime.FuncForPC(funcName).Name(), "/")
	funcNameLen := len(funcNameFullPatch)
	funcPrefixWithPackage := basic.IfElseStr(funcNameLen >= 1, funcNameFullPatch[funcNameLen-1], "")
	funcPrefix := funcPrefixWithPackage[strings.Index(funcPrefixWithPackage, ".")+1:]
	return funcPrefix
}

func Panicf(format string, args ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Panic] ")
		log.Panic("[" + formatFuncPrefix() + "] " + fmt.Sprintf(format, args...))
	}
}

func Panic(v ...interface{}) {
	if gLogFile.level >= FatalLevel {
		log.SetPrefix("[Panic] ")
		log.Panic("[" + formatFuncPrefix() + "] " + fmt.Sprint(v...))
	}
}

func (lf LogFile) Write(buf []byte) (n int, err error) {
	if lf.fileName == "" {
		fmt.Printf("consol: %s", buf)
		return len(buf), nil
	}

	if gLogFile.logTime+3600 < time_format.GetNowTime().Unix() {
		gLogFile.createLogFile()
		gLogFile.logTime = time_format.GetNowTime().Unix()
	}

	if gLogFile.fileFd == nil {
		fmt.Println(gLogFile)
		return len(buf), nil
	}

	return gLogFile.fileFd.Write(buf)
}

func (lf *LogFile) createLogFile() {
	if index := strings.LastIndex(lf.fileName, "/"); index != -1 {
		os.MkdirAll(lf.fileName[0:index], os.ModePerm)
	}

	now := time_format.GetNowTime()
	err, fileModTime := file_util.GetFileModTime(lf.fileName)
	if err != nil {
		fmt.Println(err, lf.fileName)
	}

	if err != nil || now.Format(time_format.FullFormatDateSimpleDay) != fileModTime.Format(time_format.FullFormatDateSimpleDay) {
		filename := fmt.Sprintf("%s_%s.log", lf.fileName[:strings.LastIndex(lf.fileName, ".")], fileModTime.Format(time_format.FullFormatDateSimpleDay))
		if !file_util.IsExist(filename) {
			if err := os.Rename(lf.fileName, filename); err == nil {
				// go func() {
				// 	tarCmd := exec.Command("tar", "-zcf", filename+".tar.gz", filename, "--remove-files")
				// 	tarCmd.Run()

				// 	rmCmd := exec.Command("/bin/sh", "-c", "find "+logdir+` -type f -mtime +2 -exec rm {} \;`)
				// 	rmCmd.Run()
				// }()
			}
		}
	}

	lf.fileFd = createFile(lf.fileName)
}

func (lf LogFile) clearLog() {
	fileDir := path.Dir(lf.fileName)
	for {
		files, _ := ioutil.ReadDir(fileDir)
		if fileLen := len(files); fileLen > lf.holdLogSum {
			sort.SliceStable(files, func(i, j int) bool {
				return files[i].Name() < files[j].Name()
			})
			nDelNum := fileLen - lf.holdLogSum
			delNum := 0
			for _, f := range files {
				if f.Name() == path.Base(lf.fileName) {
					continue
				}
				if delNum >= nDelNum {
					break
				}

				os.Remove(path.Join(fileDir, f.Name()))
				delNum++
			}
		}

		time.Sleep(1 * time.Minute)
	}
}
