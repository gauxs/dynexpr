package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gauxs/dynexpr/internal/bootstrap"
	"github.com/gauxs/dynexpr/internal/parser"
)

var specifiedName = flag.String("output_filename", "", "specify the filename of the output")
var processPkg = flag.Bool("pkg", false, "process the whole package instead of just the given file")

func generate(fname string) (err error) {
	fInfo, err := os.Stat(fname)
	if err != nil {
		return err
	}

	p := parser.Parser{AllStructs: true}
	if err := p.Parse(fname, fInfo.IsDir()); err != nil {
		return fmt.Errorf("error parsing %v: %v", fname, err)
	}

	var outName string
	if fInfo.IsDir() {
		outName = filepath.Join(fname, p.PkgName+"_dynexpr.go")
	} else {
		if s := strings.TrimSuffix(fname, ".go"); s == fname {
			return errors.New("filename must end in '.go'")
		} else {
			outName = s + "_dynexpr.go"
		}
	}

	if *specifiedName != "" {
		outName = *specifiedName
	}

	// add a testcase to check if `RootStructNames` has valid entries
	g := bootstrap.Bootstraper{
		PkgPath:         p.PkgPath,
		PkgName:         p.PkgName,
		Types:           p.StructNames,
		RootStructNames: p.RootStructNames,
		OutName:         outName,
		LeaveTemps:      false,
	}

	if err := g.Run(); err != nil {
		return fmt.Errorf("bootstrap failed: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()

	files := flag.Args()

	gofile := os.Getenv("GOFILE")
	if *processPkg {
		gofile = filepath.Dir(gofile)
	}

	if len(files) == 0 && gofile != "" {
		files = []string{gofile}
	} else if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, fname := range files {
		if err := generate(fname); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
