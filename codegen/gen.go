package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type AttributeInfo struct {
	Name string
	Type token.Token
	Tags []string
}

type StructInfo struct {
	StructName     string
	AttributeInfos []AttributeInfo
}

func Codegen() {
	// Parse the source code
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "codegen/test/data/person.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	ast.Print(fset, node)
	// Extract struct names and JSON tags
	// var structs []StructInfo
	// for _, decl := range node.Decls {
	// 	if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
	// 		for _, spec := range genDecl.Specs {
	// 			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
	// 				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
	// 					structInfo := StructInfo{
	// 						Name: typeSpec.Name,
	// 					}
	// 					for _, field := range structType.Fields.List {
	// 						if field.Tag != nil {
	// 							tagValue := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json")
	// 							structInfo.JSONTags = append(structInfo.JSONTags, tagValue)
	// 						}
	// 					}
	// 					structs = append(structs, structInfo)
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// // Print the results
	// for _, s := range structs {
	// 	fmt.Println("Struct Name:", s.Name)
	// 	fmt.Println("JSON Tags:", s.JSONTags)
	// 	fmt.Println()
	// }
}
