package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var tmpl = template.Must(template.New("main").Parse(`package p1am

//go:generate go run ./internal/cmd/gen_defines

type ModuleProps struct {
	ModuleID                                 uint32
	DI, DO, AI, AO, Status, Config, DataSize byte
	Name                                     string
}

var modules = []ModuleProps{
{{.MDB -}}
}

var defaultConfig = map[uint32][]byte{
	{{range .Configs -}}
	0x{{.ID}}: // {{.Name}}
	{{index $.DefaultConfigs .Name}},
	{{end}}
}

{{range .Defines}}
const {{.Name}} = {{.Value}}{{.Comment -}}
{{end}}
`))

func findLibrary() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	for _, dir := range []string{
		"Documents/Arduino",
		"Arduino",
	} {
		dir = filepath.Join(home, dir, "libraries/P1AM/src")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	return ""
}

func definitions(path string, delim string) []string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(data), delim)
}

var (
	mdbRE    = regexp.MustCompile(`(?s)mdb\[\] = \{\s*(.+)\}`)
	configRE = regexp.MustCompile(`(?s)const char (.*?)\[\] = (.+)`)
	caseRE   = regexp.MustCompile(`(?s)case 0x([^:]+):\s+return \(char\*\)(.+)`)
	defineRE = regexp.MustCompile(`(?ms)^\s*#define (\S+)\s+(\d+|0x[0-9a-fA-F]+)(\s+.*?)?\s*$`)
)

func main() {
	base := findLibrary()
	if base == "" {
		log.Fatal("can't find Arduino library")
	}
	var data = struct {
		MDB            string
		DefaultConfigs map[string]string
		Configs        []struct {
			ID   string
			Name string
		}
		Defines []struct {
			Name    string
			Value   string
			Comment string
		}
	}{
		DefaultConfigs: make(map[string]string),
	}
	for _, line := range definitions(filepath.Join(base, "Module_List.h"), ";") {
		if matches := mdbRE.FindStringSubmatch(line); matches != nil {
			data.MDB = regexp.MustCompile(`}\s*//`).ReplaceAllString(matches[1], `}, //`)
		}
		if matches := configRE.FindStringSubmatch(line); matches != nil {
			data.DefaultConfigs[matches[1]] = matches[2]
		}
	}

	for _, line := range definitions(filepath.Join(base, "P1AM.cpp"), ";") {
		if matches := caseRE.FindStringSubmatch(line); matches != nil {
			data.Configs = append(data.Configs, struct{ ID, Name string }{
				ID:   matches[1],
				Name: matches[2],
			})
		}
	}

	for _, line := range definitions(filepath.Join(base, "defines.h"), "\n") {
		if matches := defineRE.FindStringSubmatch(line); matches != nil {
			data.Defines = append(data.Defines, struct{ Name, Value, Comment string }{
				Name:    matches[1],
				Value:   matches[2],
				Comment: matches[3],
			})
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, &data); err != nil {
		log.Fatal(err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("failed to compile %s", buf.Bytes())
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("defines.go", formatted, 0666); err != nil {
		log.Fatal(err)
	}
}
