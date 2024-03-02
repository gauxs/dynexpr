package main

import (

	// Reference the gen package to be friendly to vendoring tools,
	// as it is an indirect dependency.
	// (The temporary bootstrapping code uses it.)
	"dynexpr/codegen/test/data"

	_ "github.com/mailru/easyjson/gen"

	"fmt"
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

	// g := gen.NewGenerator("person_easyjson.go")
	// g.SetPkg("data", "dynexpr/codegen/test/data")
	// g.Add(pkg.EasyJSON_exporter_BankAccount(nil))
	// g.Add(pkg.EasyJSON_exporter_BankDetails(nil))
	// g.Add(pkg.EasyJSON_exporter_Child(nil))
	// g.Add(pkg.EasyJSON_exporter_FamilyDetail(nil))
	// g.Add(pkg.EasyJSON_exporter_Person(nil))
	// if err := g.Run(os.Stdout); err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// 	os.Exit(1)
	// }
	fmt.Println("DDD")
}
