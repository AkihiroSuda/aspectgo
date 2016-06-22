package pkg1

import "fmt"

type S struct {
}

func (s *S) Foo(x string) string {
	return fmt.Sprintf("pkg1, x=%s", x)
}
