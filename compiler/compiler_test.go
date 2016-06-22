package compiler

import (
	"os"
	"testing"
)

func TestResolveTarget(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		t.Fatal("GOPATH not set")
	}

	target := "golang.org/..."
	resolved, err := resolveTarget(gopath, target)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("resolved %s", target)
	for _, r := range resolved {
		t.Logf("- %s", r)
	}
}
