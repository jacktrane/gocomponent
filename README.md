# 既然这样了，就写个包吧

## 一、redo

> 重试机，用于保障系统高可用

``` go
用例：
    type TestMachine struct {
        StatCode int 
        Num      int
        Str      string
    }

    const (
        MachineStatCodeTest1 = iota
        MachineStatCodeTest2
        MachineStatCodeTest3
        MachineStatCodeTest4
    )

    // 返回值： bool 状态/数据是否改变 error 是否有错误
    // 重试机只有在error出现时才会重试
    func (t *TestMachine) Run() (bool, error) {
        statCode := t.StatCode
        for {
            switch t.StatCode {
            case MachineStatCodeTest1:
                if t.Num < 4 {
                    t.Num++
                    logger.Errorf("%+v", t)
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
            SliceFileInterval:    3600,
            Machine:              &emptyMachine,
            PollInterval:         1,
        }
        redo := NewRedoActionConf(conf)

        var testMachine TestMachine
        redo.StableAction(&testMachine)
    }
```

## 二、logger

用例：参考 `logger/logger_test.go`

## 三、打点耗时

用例：参考 `time_cost/time_cost.go`

## 四、mcache

LRU的缓存，但相对来说会存在一些IO，谨慎使用，一般来说需要做LRU或者LFU的主要原因是担心OOM，如果对于数据量存储不多的服务其实直接设置全局map就已经够用了

## TODO 五、共享内存（shm）

进程之间通信,此处的共享内存只是实验，投入生产存在问题

# TODO 六、uuid


# 七、协程池

用例： 参考 `gopool/gopool_test.go`

# 