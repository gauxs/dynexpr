package data

import (
	"fmt"
	"os"
)

type Dynexpr_exporter_BankAccount *BankAccount
type Dynexpr_exporter_BankDetails *BankDetails
type Dynexpr_exporter_Child *Child
type Dynexpr_exporter_FamilyDetail *FamilyDetail
type Dynexpr_exporter_Person *Person

func Generate() {
	g := NewGenerator()
	g.SetPkg("data", "dynexpr/codegen/test/data")
	g.Add(Dynexpr_exporter_BankAccount(nil))
	g.Add(Dynexpr_exporter_BankDetails(nil))
	g.Add(Dynexpr_exporter_Child(nil))
	g.Add(Dynexpr_exporter_FamilyDetail(nil))
	g.Add(Dynexpr_exporter_Person(nil))
	if err := g.Run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Done")
}
