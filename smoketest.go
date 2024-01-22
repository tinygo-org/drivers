//go:build none

// Run all commands in smoketest.sh in parallel, for improved performance.
// This requires changing the -o flag to avoid a race condition between writing
// the output and reading it back to get the md5sum of the output.

package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/google/shlex"
)

var flagXtensa = flag.Bool("xtensa", true, "Enable Xtensa tests")

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./smoketest.go smoketest.txt")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}
	err := runSmokeTest(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		usage()
		os.Exit(1)
	}
}

func runSmokeTest(filename string) error {
	// Read all the lines in the file.
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := bytes.Split(data, []byte("\n"))

	// Start a number of goroutine workers.
	jobChan := make(chan *Job, len(lines))
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(jobChan)
	}

	// Create a temporary directory for the outputs of these tests.
	tmpdir, err := os.MkdirTemp("", "drivers-smoketest-*")
	if err != nil {
		return err
	}

	// Send work to the workers.
	var jobs []*Job
	for _, lineBytes := range lines {
		// Parse the line into command line parameters.
		line := string(lineBytes)
		fields, err := shlex.Split(line)
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			continue // empty line
		}

		// Replace the "output" flag, but store the original value.
		var outpath, origOutpath string
		for i := range fields {
			if fields[i] == "-o" {
				origOutpath = fields[i+1]
				ext := path.Ext(origOutpath)
				outpath = filepath.Join(tmpdir, fmt.Sprintf("output-%d%s", len(jobs), ext))
				fields[i+1] = outpath
				break
			}
		}
		if outpath == "" {
			return fmt.Errorf("could not find -o flag in command: %v", fields)
		}

		// Parse the command line parameters to get the -target flag.
		if fields[1] != "build" {
			return fmt.Errorf("unexpected subcommand: %#v", fields[1])
		}
		flagSet := flag.NewFlagSet(fields[0], flag.ContinueOnError)
		_ = flagSet.String("size", "", "")
		_ = flagSet.String("o", "", "")
		_ = flagSet.String("stack-size", "", "")
		targetFlag := flagSet.String("target", "", "")
		err = flagSet.Parse(fields[2:])
		if err != nil {
			return fmt.Errorf("failed to parse command from %s: %w", filename, err)
		}

		// Skip Xtensa tests if set in the flag.
		if !*flagXtensa && *targetFlag == "m5stack-core2" {
			continue
		}

		// Create TinyGo command (to build the driver example).
		output := &bytes.Buffer{}
		output.Write(lineBytes)
		output.Write([]byte("\n"))
		cmd := exec.Command(fields[0], fields[1:]...)
		cmd.Stdout = output
		cmd.Stderr = output

		// Submit this command for execution.
		job := &Job{
			output:      output,
			tinygoCmd:   cmd,
			outpath:     outpath,
			origOutpath: origOutpath,
			resultChan:  make(chan error),
		}
		jobChan <- job
		jobs = append(jobs, job)
	}
	close(jobChan) // stops the workers (probably not necessary)

	// Read the output from all these jobs, in order.
	for _, job := range jobs {
		result := <-job.resultChan
		os.Stdout.Write(job.output.Bytes())
		if result != nil {
			return result
		}
	}

	return nil
}

func worker(jobChan chan *Job) {
	for job := range jobChan {
		// Run the tinygo command.
		err := job.tinygoCmd.Run()
		if err != nil {
			job.resultChan <- err
		}

		// Create a md5sum, with output similar to the "md5sum" command line
		// utility.
		data, err := os.ReadFile(job.outpath)
		if err != nil {
			job.resultChan <- err
		}
		fmt.Fprintf(job.output, "%x  %s\n", md5.Sum(data), job.origOutpath)
		job.resultChan <- nil
	}
}

type Job struct {
	output      *bytes.Buffer
	tinygoCmd   *exec.Cmd
	outpath     string
	origOutpath string
	resultChan  chan error
}
