### 既然这样了，就写个包吧
### 一、redo
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

### 二、logger
用例：参考 logger/logger_test.go