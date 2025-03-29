package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"text/template"
)

var templatePath = flag.String("template", "", "path to the template file to render")
var fileModeStr = flag.String("mode", "0600", "file mode for the rendered template")
var outPath = flag.String("out", "", "destination path for the rendered template")
var clearEnv = flag.Bool("clear-env", false, "used in conjunction with -exec, when specified doesn't pass any env vars to new command")
var execCommand = flag.Bool("exec", false, "execve remaining args if specified")

func renderTemplate(templatePath string) (string, bool) {
	templateError := false
	funcMap := template.FuncMap{
		"env": func(name string) string {
			return os.Getenv(name)
		},
		"envDefault": func(name string, defaultVal string) string {
			value, exists := os.LookupEnv(name)
			if !exists {
				return defaultVal
			}
			return value
		},
		"requiredEnv": func(name string) string {
			value, exists := os.LookupEnv(name)
			if !exists {
				log.Printf("ERROR: requiredEnv, env var %s is required\n", name)
				templateError = true
				return ""
			} else {
				return value
			}
		},
		"readFile": func(path string) string {
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("ERROR: readFile, error while reading %s: %s\n", path, err)
				templateError = true
				return ""
			}
			return string(data)
		},
	}

	templateText, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("ERROR: couldn't read template: %s\n", err)
	}

	tmpl, err := template.New("config").Funcs(funcMap).Parse(string(templateText))
	if err != nil {
		log.Fatalf("ERROR: couldn't parse template: %s\n", err)
	}

	templResult := new(strings.Builder)
	if err := tmpl.Execute(templResult, nil); err != nil {
		log.Fatalf("ERROR: couldn't render template: %s\n", err)
	}

	return templResult.String(), templateError
}

func main() {
	flag.Parse()

	if *templatePath == "" {
		log.Fatalln("ERROR: -template is required")
	}
	if *outPath == "" {
		log.Fatalln("ERROR: -out is required")
	}
	if *execCommand && flag.NArg() == 0 {
		log.Fatalln("ERROR: -exec is specified but no command was provided")
	}

	fileMode, err := strconv.ParseUint(*fileModeStr, 8, 16)
	if err != nil {
		log.Fatalf("ERROR: couldn't parse -mode %s to string: %s", *fileModeStr, err)
	}

	templResult, templateError := renderTemplate(*templatePath)
	if templateError {
		log.Fatalln("Errors occured, exiting ...")
	}

	if err := os.WriteFile(*outPath, []byte(templResult), fs.FileMode(fileMode)); err != nil {
		log.Fatalf("ERROR: couldn't write output file %s: %s\n", *outPath, err)
	}

	if err := os.Chmod(*outPath, fs.FileMode(fileMode)); err != nil {
		log.Fatalf("ERROR: couldn't chmod output file %s: %s\n", *outPath, err)
	}

	if !*execCommand {
		log.Printf("generated file %s from %s\n", *outPath, *templatePath)
		return
	}

	args := flag.Args()
	if !strings.HasPrefix(args[0], "/") {
		// path to executable is not absolute, resolve to full path
		execPath, err := exec.LookPath(args[0])
		if err != nil {
			log.Fatalf("ERROR: couldn't find command %s in $PATH", execPath)
		}
		// override the argument wiht the full path
		args[0] = execPath
	}

	newProcEnv := []string{}
	if !*clearEnv {
		newProcEnv = os.Environ()
	}

	if err := syscall.Exec(args[0], args, newProcEnv); err != nil {
		log.Fatalf("ERROR: couldn't exec command '%s': %s", strings.Join(args, " "), err)
	}
}
