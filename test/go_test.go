package test

import (
	"fmt"
	"os"
	"testing"
)

func Test_Mkdir(t *testing.T) {
	fmt.Println(os.Mkdir("./a/b/c/d", os.ModePerm))
	fmt.Println(os.MkdirAll("./e/f/g/h", os.ModePerm))
	fmt.Println(os.Mkdir("./i", os.ModePerm))
	fmt.Println("a")
}
