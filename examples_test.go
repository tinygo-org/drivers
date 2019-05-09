package examples_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExamples(t *testing.T) {
	examples, err := filepath.Glob("examples/**/main.go")
	if err != nil {
		t.Fatal("Could not read exampels: ", err)
	}

	targets := []string{
		"",
	}

	if !testing.Short() {
		targets = append(targets,
			"arduino",
		)
	}

	for _, example := range examples {
		for _, target := range targets {
			t.Run(filepath.Join(target, example), func(t *testing.T) {
				buildExample(example, target, t)
			})
		}
	}

}

func buildExample(example string, target string, t *testing.T) {
	t.Parallel()

	//tmpFile, err := ioutil.TempFile("", "tinygo-example-test")
	//if err != nil {
	//	t.Fatal("could not create temporary file:", err)
	//}
	//defer os.Remove(tmpFile.Name())

	cmd := exec.Command(
		"tinygo",
		"build",
		"-size", "short",
		"-o", "/dev/null", //tmpFile.Name(),
		"-target", target,
		fmt.Sprintf("./%s", example),
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		t.Log("failed to run:", err)
		t.Log("stdout:", stdout)
		t.Log("stderr:", stderr)
		t.Fail()
	}
}
