package pkg2

import "fmt"

type S struct {
}

func (s *S) Foo(x string) string {
	return fmt.Sprintf("pkg2, x=%s", x)
}
