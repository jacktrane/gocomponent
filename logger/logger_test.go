package logger

import (
	"fmt"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	// Fatalf("222222")
	// NewConfig(path.Join("..", "runtime", "log", "default.log"), 5)
	Infof("info1234567")
	// Errorf("error1234567")
	// Debugf("debug1234567")
	time.Sleep(time.Second)
	// log.Fatal("111111")
	// for {

	// }
}

func TestByte(t *testing.T) {
	arrByte := []byte{91, 73, 110, 102, 111, 93, 32, 50, 48, 50, 49, 47, 48, 50, 47, 50, 54, 32, 50, 51, 58, 49, 52, 58, 52, 49, 46, 52, 51, 56, 51, 54, 54, 32, 108, 111, 103, 103, 101, 114, 95, 116, 101, 115, 116, 46, 103, 111, 58, 49, 49, 58, 32}
	arrByte1 := []byte{91, 73, 110, 102, 111, 93, 32, 50, 48, 50, 49, 47, 48, 50, 47, 50, 54, 32, 50, 51, 58, 49, 52, 58, 52, 49, 46, 52, 51, 56, 51, 54, 54, 32, 108, 111, 103, 103, 101, 114, 95, 116, 101, 115, 116, 46, 103, 111, 58, 49, 49}
	fmt.Println(string(arrByte), "\n", string(arrByte1))
	arrByte2 := []byte{58, 32}
	fmt.Println(string(arrByte2))
}
