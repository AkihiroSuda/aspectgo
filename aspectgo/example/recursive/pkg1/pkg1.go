package pkg1

import (
	"fmt"
	"golang.org/x/exp/aspectgo/example/recursive/pkg1/pkg2"
)

func SayHello1() {
	fmt.Println("hello from pkg1")
	pkg2.SayHello2()
}
