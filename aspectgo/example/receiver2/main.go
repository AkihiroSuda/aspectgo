package main

import (
	"fmt"
	"golang.org/x/exp/aspectgo/example/receiver2/pkg1"
	. "golang.org/x/exp/aspectgo/example/receiver2/pkg1"
	xpkg1 "golang.org/x/exp/aspectgo/example/receiver2/pkg1"
	"golang.org/x/exp/aspectgo/example/receiver2/pkg2"
	xpkg2 "golang.org/x/exp/aspectgo/example/receiver2/pkg2"
)

func main() {
	s1 := pkg1.S{}
	fmt.Println(s1.Foo("pkg1.S{}"))
	s1p := &pkg1.S{}
	fmt.Println(s1p.Foo("&pkg1.S{}"))

	s2 := pkg2.S{}
	fmt.Println(s2.Foo("pkg2.S{} (should be ignored)"))

	s := S{}
	fmt.Println(s.Foo("S{}"))

	xs1 := xpkg1.S{}
	fmt.Println(xs1.Foo("xpkg1.S{}"))

	xs2 := xpkg2.S{}
	fmt.Println(xs2.Foo("xpkg2.S{} (should be ignored)"))

	fmt.Println((&pkg1.S{}).Foo("(&pkg1.S{})"))
	fmt.Println((&pkg2.S{}).Foo("(&pkg2.S{}) (should be ignored)"))
	fmt.Println((&S{}).Foo("(&S{})"))
}
