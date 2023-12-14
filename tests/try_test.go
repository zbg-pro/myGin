package test

import (
	"errors"
	fmt "fmt"
	"myGin/service"
	"testing"
)

func TestTry2(t *testing.T) {
	service.Try(func() {
		i := 100
		a := 0
		c := i / a

		if c > 1 {
			panic("不能大于1！！！")
		}

		fmt.Errorf("")
		errors.New("")

	}).Catch(func(err interface{}) {
		fmt.Println("Caught an error:", err)
	}).Finally(func() {
		fmt.Println("finally。。。")
	}).Run()

	fmt.Println(2323434)
}
