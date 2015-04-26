package gomvc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// getActionsFromSourceFiles parses all controller source files and fetches
// data about action functions
func getActionsFromSourceFiles() {
	if !isDev {
		return
	}
	ActionArgs = make(map[string]map[string][]string, 0)
	curdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	files, err := ioutil.ReadDir("c/")
	if err != nil {
		panic(`
Can't find the controllers directory in debug mode.
Make sure you are running the application from the directory where it's located
Current directory: ` + curdir)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") &&
			!strings.HasSuffix(file.Name(), "_test.go") {
			getActionsFromSourceFile(file.Name())
		}
	}
	out, _ := os.Create("autogen/autogen.go")
	fmt.Fprintln(out, `
// This file has been generated automatically. Do not modify it.
package autogen

import "github.com/medvednikov/gomvc"

func init() {
	gomvc.ActionArgs = `+dump(ActionArgs)+`
}`)
	out.Close()
}

func getActionsFromSourceFile(sourceFile string) {
	b, err := ioutil.ReadFile("c/" + sourceFile)
	handle(err)
	source := string(b)
	pos := strings.Index(sourceFile, ".go")
	if pos == -1 {
		return
	}
	controllerName := capitalize(sourceFile[:pos])
	ActionArgs[controllerName] = make(map[string][]string, 0)
	r := regexp.MustCompile(
		`func \([a-zA-Z]+ \*` + controllerName + `\) (.*?)\((.*?)*\) (.*?){`)
	matches := r.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		functionName := match[1]
		if functionName == "" {
			continue
		}
		args := []string{}
		if len(match) > 2 && match[2] != "" {
			args = strings.Split(match[2], ", ")
			for i, arg := range args {
				if arg == "" {
					continue
				}
				args[i] = strings.Split(arg, " ")[0]
			}
		}
		ActionArgs[controllerName][functionName] = args
	}
}
