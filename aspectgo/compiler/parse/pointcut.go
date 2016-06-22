package parse

import (
	"bytes"
	"fmt"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"golang.org/x/exp/aspectgo/aspect"
	"golang.org/x/exp/aspectgo/compiler/consts"
)

// compile the aspect and get Pointcut data
// steps:
//  * copy the aspect file to tmp.go
// * add main() to tmp.go
// * compile and run tmp.go
// * parse the output and generate Pointcut data
func (af *AspectFile) determinePointcuts(aspects []*types.Named) error {
	// TODO: do them at once
	for _, aspect := range aspects {
		dir, err := ioutil.TempDir("", "aspectgo")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		if err = locateTmpAspectFile(af.Filename, dir); err != nil {
			return err
		}
		if err = locateTmpAspectMainFile(aspect.Obj().Name(), dir); err != nil {
			return err
		}
		s, err := runTmpAspectMain(dir)
		if err != nil {
			return err
		}
		pointcut, err := parseTmpAspectMainOutput(s)
		if err != nil {
			return err
		}
		af.Pointcuts[aspect] = pointcut
	}
	return nil
}

// locate aspectFilename to dir to determine the pointcut value
// TODO: eliminate aspectStructure.Advice()
func locateTmpAspectFile(aspectFilename, dir string) error {
	tmpAspectFile := filepath.Join(dir, "aspect.go")
	cont, err := ioutil.ReadFile(aspectFilename)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(tmpAspectFile, cont, 0444); err != nil {
		return err
	}
	return nil
}

const tmpAspectMainFileTmpl = consts.AutogenFileHeader + `package main

import (
    "fmt"
    "io/ioutil"
    "os"
)

func main() {
    if len(os.Args) != 2 {
        panic(fmt.Errorf("args len mismatch: %s", os.Args))
    }
    fName := os.Args[1]

    asp := &{{.aspectStructureName}}{}
    pointcut := asp.Pointcut()

    err := ioutil.WriteFile(fName, []byte(pointcut), 0444)
    if err != nil {
        panic(err)
    }
}
`

func locateTmpAspectMainFile(aspectStructureName, dir string) error {
	var b bytes.Buffer
	t := template.New("t")
	m := map[string]string{"aspectStructureName": aspectStructureName}
	template.Must(t.Parse(tmpAspectMainFileTmpl))
	if err := t.Execute(&b, m); err != nil {
		return err
	}
	file := filepath.Join(dir, "main.go")
	if err := ioutil.WriteFile(file, b.Bytes(), 0444); err != nil {
		return err
	}
	return nil
}

func runTmpAspectMain(dir string) (string, error) {
	cmdName := "go"
	arg := []string{"run", "main.go", "aspect.go", "result.txt"}
	cmd := exec.Command(cmdName, arg...)
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return "",
			fmt.Errorf("error while executing %s %s at %s: %s: %s",
				cmdName, arg, dir, err, stderr.String())
	}
	stdoutS := stdout.String()
	if stdoutS != "" {
		log.Printf("got stdout: %s", stdoutS)
	}
	result, err := ioutil.ReadFile(filepath.Join(dir, "result.txt"))
	if err != nil {
		return "", err
	}
	resultS := string(result)
	return resultS, nil
}

func parseTmpAspectMainOutput(s string) (aspect.Pointcut, error) {
	return aspect.Pointcut(s), nil
}
