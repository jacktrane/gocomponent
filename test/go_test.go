package test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func Test_Mkdir(t *testing.T) {
	fmt.Println(os.Mkdir("./a/b/c/d", os.ModePerm))
	fmt.Println(os.MkdirAll("./e/f/g/h", os.ModePerm))
	fmt.Println(os.Mkdir("./i", os.ModePerm))
	fmt.Println("a")
}

type A struct {
	B int
	C string
}
type H struct {
	B int
	C string
	D int
}

var a A = A{
	B: 2,
	C: "ssss",
}

func (a A) Run() error {
	fmt.Println(111111)
	return nil
}

type I interface {
	Run() error
}

func Test_Struct(t *testing.T) {
	machineValue := reflect.ValueOf(a)
	var b = H{D: 10}
	br := reflect.ValueOf(b)

	fmt.Println(machineValue.FieldByName("D").IsValid())
	fmt.Println(br.FieldByName("D").IsValid())
	fmt.Println(br.FieldByName("D").IsZero())
	fmt.Println(br.FieldByName("D").Type().Name() == "string")
	fmt.Println(br.FieldByName("D").Type().Name() == "int")
	fmt.Println(br.FieldByName("D").Interface().(int))
	fmt.Println(br.FieldByName("D").Int())
}

func Test_Inte(t *testing.T) {
	var ta = A{B: 2,
		C: "ssss"}
	inte(ta)
}

func inte(i I) {
	byteJson, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(byteJson))

}
