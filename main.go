package main

import (
	"dynexpr/codegen/test/data"

	// Reference the gen package to be friendly to vendoring tools,
	// as it is an indirect dependency.
	// (The temporary bootstrapping code uses it.)
	_ "github.com/mailru/easyjson/gen"
)

func main() {
	// flag.Parse()

	// files := flag.Args()

	// gofile := os.Getenv("GOFILE")
	// gofile = filepath.Dir(gofile)

	// if len(files) == 0 && gofile != "" {
	// 	files = []string{gofile}
	// } else if len(files) == 0 {
	// 	flag.Usage()
	// 	os.Exit(1)
	// }

	// for _, fname := range files {
	// 	if err := codegen.Generate(fname); err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 		os.Exit(1)
	// 	}
	// }
	data.Generate()
}
