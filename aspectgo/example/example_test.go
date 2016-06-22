package example

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	agcli "golang.org/x/exp/aspectgo/compiler/cli"
)

var (
	GOPATH string
)

const (
	exPackage = "golang.org/x/exp/aspectgo/example"
)

func TestMain(m *testing.M) {
	flag.Parse()
	GOPATH = os.Getenv("GOPATH")
	if GOPATH == "" {
		panic(fmt.Errorf("GOPATH not set"))
	}
	os.Exit(m.Run())
}

func execAspectGo(t *testing.T, wovenGOPATH, pkg, aspectFileBasename string, recursive bool) error {
	pkgDir := filepath.Join(GOPATH, filepath.Join("src", pkg))
	aspectFilename := filepath.Join(pkgDir, aspectFileBasename)
	if recursive {
		pkg += "/..."
	}
	args := []string{"aspectgo", "-w", wovenGOPATH, "-t", pkg}
	if testing.Verbose() {
		args = append(args, "-debug=true")
	}
	args = append(args, []string{"--", aspectFilename}...)
	t.Logf("Running AspectGo with: %s", args[1:])
	exitCode := agcli.Main(args)
	if exitCode != 0 {
		return fmt.Errorf("Build failed with exit code %d", exitCode)
	}
	return nil
}

func execMainWithGOPATH(t *testing.T, gopath, pkg, mainFileBasename string) ([]byte, error) {
	if gopath == "" {
		gopath = GOPATH
	}
	pkgDir := filepath.Join(gopath, filepath.Join("src", pkg))
	mainFilename := filepath.Join(pkgDir, mainFileBasename)
	cmd := exec.Command("go", []string{"run", mainFilename}...)
	cmd.Env = []string{fmt.Sprintf("GOPATH=%s", gopath)}
	out, err := cmd.CombinedOutput()
	t.Logf("Test Result (GOPATH=%s):\n%s", gopath, string(out))
	return out, err
}

// textEx returns the output of the original test and the woven test suite if succeeds.
// the output contains stderr.
// if the woven test or aspectgo itself fails, testEx panics.
func testEx(t *testing.T, dirname, mainFileBasename, aspectFileBasename string, recursive bool) ([]byte, []byte) {
	t.Parallel()
	pkg := filepath.Join(exPackage, dirname)
	out1, err := execMainWithGOPATH(t, "", pkg, mainFileBasename)
	if err != nil {
		t.Fatal(err)
	}

	wovenGOPATH, err := ioutil.TempDir("", "agtestwovengopath")
	if err != nil {
		t.Fatal(err)
	}

	err = execAspectGo(t, wovenGOPATH, pkg, aspectFileBasename, recursive)
	if err != nil {
		t.Fatal(err)
	}

	out2, err := execMainWithGOPATH(t, wovenGOPATH, pkg, mainFileBasename)
	if err != nil {
		t.Fatal(err)
	}

	if testing.Verbose() {
		t.Logf("Keeping woven directory %s", wovenGOPATH)
	} else {
		os.RemoveAll(wovenGOPATH)
	}

	return out1, out2
}

func TestExHello(t *testing.T) {
	testEx(t, "hello", "main.go", "main_aspect.go", false)
}

func TestExHello2(t *testing.T) {
	testEx(t, "hello2", "main.go", "main_aspect.go", false)
}

func TestExHello3(t *testing.T) {
	testEx(t, "hello3", "main.go", "main_aspect.go", false)
}

func TestExReceiver(t *testing.T) {
	testEx(t, "receiver", "main.go", "main_aspect.go", false)
}

func TestExReceiver2(t *testing.T) {
	testEx(t, "receiver2", "main.go", "main_aspect.go", false)
}

func TestExMultipointcut(t *testing.T) {
	testEx(t, "multipointcut", "main.go", "main_aspect.go", false)
}

func TestExDetreplay(t *testing.T) {
	testEx(t, "detreplay", "main.go", "main_aspect.go", false)
}

func TestExRecursive(t *testing.T) {
	testEx(t, "recursive", "main.go", "main_aspect.go", true)
}
