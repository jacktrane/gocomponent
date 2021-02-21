package test

import (
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

func Test_Struct(t *testing.T) {
	machineValue := reflect.ValueOf(a)
	var b H
	br := reflect.ValueOf(b)

	fmt.Println(machineValue.FieldByName("D").IsValid())
	fmt.Println(br.FieldByName("D").IsValid())
	fmt.Println(br.FieldByName("D").IsZero())
	fmt.Println(br.FieldByName("D").Type().Name() == "string")
	fmt.Println(br.FieldByName("D").Type().Name() == "int")
}
