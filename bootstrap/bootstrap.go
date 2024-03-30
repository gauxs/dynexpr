// Package bootstrap implements the bootstrapping logic: generation of a .go file to
// launch the actual generator and launching the generator itself.
//
// The package may be preferred to a command-line utility if generating the expression builder
// from golang code is required.
package bootstrap

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const genPackage = "dynexpr/codegen"

type Generator struct {
	PkgPath, PkgName string
	Types            []string
	OutName          string
	LeaveTemps       bool
	NoFormat         bool
}

// writeMain creates a .go file that launches the generator if 'go run'.
func (g *Generator) writeMain() (path string, err error) {
	f, err := ioutil.TempFile(filepath.Dir(g.OutName), "dynexpr-bootstrap")
	if err != nil {
		return "", err
	}

	fmt.Fprintln(f, "// +build ignore")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "// TEMPORARY AUTOGENERATED FILE: dynexpr bootstapping code to launch")
	fmt.Fprintln(f, "// the actual generator.")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "package main")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "import (")
	fmt.Fprintln(f, `  "fmt"`)
	fmt.Fprintln(f, `  "os"`)
	fmt.Fprintln(f)
	fmt.Fprintf(f, "  %q\n", genPackage)
	if len(g.Types) > 0 {
		fmt.Fprintln(f)
		fmt.Fprintf(f, "  pkg %q\n", g.PkgPath)
	}
	fmt.Fprintln(f, ")")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "func main() {")
	fmt.Fprintf(f, "  g := codegen.NewGenerator()\n")
	fmt.Fprintf(f, "  g.SetPkg(%q, %q)\n", g.PkgName, g.PkgPath)

	sort.Strings(g.Types)
	for _, v := range g.Types {
		if !strings.HasSuffix(v, "ExpressionBuilder") {
			fmt.Fprintln(f, "  g.Add(pkg."+v+"{})")
		}
	}

	fmt.Fprintln(f, "  if err := g.Run(os.Stdout); err != nil {")
	fmt.Fprintln(f, "    fmt.Fprintln(os.Stderr, err)")
	fmt.Fprintln(f, "    os.Exit(1)")
	fmt.Fprintln(f, "  }")
	fmt.Fprintln(f, "}")

	src := f.Name()
	if err := f.Close(); err != nil {
		return src, err
	}

	dest := src + ".go"
	return dest, os.Rename(src, dest)
}

func (g *Generator) Run() error {
	path, err := g.writeMain()
	if err != nil {
		return err
	}
	if !g.LeaveTemps {
		defer os.Remove(path)
	}

	f, err := os.Create(g.OutName + ".tmp")
	if err != nil {
		return err
	}
	if !g.LeaveTemps {
		defer os.Remove(f.Name()) // will not remove after rename
	}

	execArgs := []string{"run"}
	execArgs = append(execArgs, filepath.Base(path))
	cmd := exec.Command("go", execArgs...)

	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(path)
	if err = cmd.Run(); err != nil {
		return err
	}
	f.Close()

	// move unformatted file to out path
	if g.NoFormat {
		return os.Rename(f.Name(), g.OutName)
	}

	// format file and write to out path
	in, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return err
	}
	out, err := format.Source(in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(g.OutName, out, 0644)
}