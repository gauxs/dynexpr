package codegen

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const pkgDynexpr = "github.com/gauxs/dynexpr/pkg/v1"

// fieldTags contains parsed version of json struct field tags.
type fieldTags struct {
	name         string
	partitionKey bool
	sortKey      bool
}

type Generator struct {
	out *bytes.Buffer

	pkgName string
	pkgPath string

	rootStructNames map[string]struct{}

	// package path to local alias map for tracking imports
	imports map[string]string

	// types that encoders were already generated for
	typesSeen map[reflect.Type]bool

	// types that encoders were requested for (e.g. by encoders of other types)
	typesUnseen []reflect.Type
}

// SetPkg sets the name and path of output package.
func (g *Generator) SetPkg(name, path string) {
	g.pkgName = name
	g.pkgPath = path
}

// Add requests to generate marshaler/unmarshalers and encoding/decoding
// funcs for the type of given object.
func (g *Generator) Add(obj interface{}) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	g.addType(t)
}

// addTypes requests to generate encoding/decoding funcs for the given type.
func (g *Generator) addType(t reflect.Type) {
	if g.typesSeen[t] {
		return
	}

	for _, t1 := range g.typesUnseen {
		if t1 == t {
			return
		}
	}

	g.typesUnseen = append(g.typesUnseen, t)
}

// Run runs the generator and outputs generated code to out.
func (g *Generator) Run(out io.Writer) error {
	g.out = &bytes.Buffer{}

	for len(g.typesUnseen) > 0 {
		t := g.typesUnseen[len(g.typesUnseen)-1]
		g.typesUnseen = g.typesUnseen[:len(g.typesUnseen)-1]
		g.typesSeen[t] = true

		if err := g.genExpressionBuilder(t); err != nil {
			return err
		}
	}

	// generate root structs object builder
	for rootStructName, _ := range g.rootStructNames {
		fmt.Fprintln(g.out, "func New"+rootStructName+"_ExpressionBuilder() dynexpr.DDBItemExpressionBuilder[*"+rootStructName+"_ExpressionBuilder] {")
		fmt.Fprintln(g.out, "\treturn dynexpr.NewDDBItemExpressionBuilder(&"+rootStructName+"_ExpressionBuilder{})")
		fmt.Fprintln(g.out, "}")
		fmt.Println()
	}

	g.printHeader()
	_, err := out.Write(g.out.Bytes())
	return err
}

func (g *Generator) genExpressionBuilder(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return nil
	default:
		return g.genStructExpressionBuilder(t)
	}
}

func (g *Generator) genStructExpressionBuilder(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("cannot generate expression builder for %v, not a struct type", t)
	}

	// get struct name
	structName := t.Name()

	expressionBldrStructName := structName + "_ExpressionBuilder"
	fmt.Fprintln(g.out, "type "+expressionBldrStructName+" struct {")

	// get structs field names
	fs, err := getStructFields(t)
	if err != nil {
		return fmt.Errorf("cannot generate decoder for %v: %v", t, err)
	}

	for _, f := range fs {
		fieldType := g.getType(f.Type)
		fieldTags := parseFieldTags(f)

		if f.Type.Kind() == reflect.Pointer {
			if f.Type.Elem().Kind() == reflect.Array || f.Type.Elem().Kind() == reflect.Slice {
				fieldType = g.getType(f.Type.Elem().Elem())
			} else if f.Type.Elem().Kind() == reflect.Struct {
				// do nothing
			}
		}

		// if json tag has dynexpr:"partionKey" this is a partition key attribute or
		// if json tag has dynexpr:"partionKey" this is a partition key attribute
		// and we use DynamoKeyAttribute
		if fieldTags.partitionKey || fieldTags.sortKey {
			fmt.Fprintln(g.out, "\t"+f.Name+"\tdynexpr.DynamoKeyAttribute["+fieldType+"]\t")
		} else if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice { // if type is array/slice then we use DynamoListAttribute
			fmt.Fprintln(g.out, "\t"+f.Name+"\tdynexpr.DynamoListAttribute["+fieldType+"]\t")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Array || f.Type.Elem().Kind() == reflect.Slice) {
			fmt.Fprintln(g.out, "\t"+f.Name+"\tdynexpr.DynamoListAttribute["+fieldType+"]\t")
		} else if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Chan || f.Type.Kind() == reflect.Interface {
			return fmt.Errorf("field %s has unsupported type %s", f.Name, f.Type.Kind())
		} else if f.Type.Kind() == reflect.Struct && !g.isExternalImport(g.getPakagePath(f.Type)) {
			fmt.Fprintln(g.out, "\t"+f.Name+"\tdynexpr.DynamoAttribute["+fieldType+"_ExpressionBuilder]\t")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Struct) && !g.isExternalImport(g.getPakagePath(f.Type)) {
			fmt.Fprintln(g.out, "\t"+f.Name+"\tdynexpr.DynamoAttribute["+fieldType+"_ExpressionBuilder]\t")
		} else { // if type is native then use DynamoAttribute
			fmt.Fprintln(g.out, "\t"+f.Name+"\tdynexpr.DynamoAttribute["+fieldType+"]\t")
		}
	}
	fmt.Fprintln(g.out, "}")

	// generate function
	fmt.Fprintln(g.out, "func (o *"+expressionBldrStructName+") BuildTree(name string) *dynexpr.DynamoAttribute[*"+expressionBldrStructName+"] {")
	fmt.Fprintln(g.out, "\to = &"+expressionBldrStructName+"{}")
	for _, f := range fs {
		fieldTags := parseFieldTags(f)
		fieldType := g.getType(f.Type) // f.Type.String()
		if f.Type.Kind() == reflect.Pointer {
			if f.Type.Elem().Kind() == reflect.Array || f.Type.Elem().Kind() == reflect.Slice {
				fieldType = g.getType(f.Type.Elem().Elem())
			} else if f.Type.Elem().Kind() == reflect.Struct && !g.isExternalImport(g.getPakagePath(f.Type)) {
				fieldType = g.getType(f.Type.Elem())
			}
		}

		// if json tag has dynexpr:"partionKey" this is a partition key attribute or
		// if json tag has dynexpr:"partionKey" this is a partition key attribute
		// and we use DynamoKeyAttribute
		if fieldTags.partitionKey || fieldTags.sortKey {
			fmt.Fprintln(g.out, "o."+f.Name+" = *dynexpr.NewDynamoKeyAttribute["+fieldType+"]().WithName(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice { // if type is array/slice then we use DynamoListAttribute
			fmt.Fprintln(g.out, "o."+f.Name+" = *dynexpr.NewDynamoListAttribute["+fieldType+"]().WithName(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Array || f.Type.Elem().Kind() == reflect.Slice) {
			fmt.Fprintln(g.out, "o."+f.Name+" = *dynexpr.NewDynamoListAttribute["+fieldType+"]().WithName(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Chan || f.Type.Kind() == reflect.Interface {
			return fmt.Errorf("field %s has unsupported type %s", f.Name, f.Type.Kind())
		} else if f.Type.Kind() == reflect.Struct && !g.isExternalImport(g.getPakagePath(f.Type)) {
			fmt.Fprintln(g.out, "o."+f.Name+" = *(&"+fieldType+"_ExpressionBuilder{}).BuildTree(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Struct) && !g.isExternalImport(g.getPakagePath(f.Type)) {
			fmt.Fprintln(g.out, "o."+f.Name+" = *(&"+fieldType+"_ExpressionBuilder{}).BuildTree(\""+fieldTags.name+"\")")
		} else { // if type is native then use DynamoAttribute
			fmt.Fprintln(g.out, "o."+f.Name+" = *dynexpr.NewDynamoAttribute["+fieldType+"]().WithName(\""+fieldTags.name+"\")")
		}
	}
	fmt.Fprintln(g.out, "\t return dynexpr.NewDynamoAttribute[*"+expressionBldrStructName+"]().")
	fmt.Fprintln(g.out, "\t\tWithAccessReference(o).")
	fmt.Fprintln(g.out, "\t\tWithName(name).")
	for idx, f := range fs {
		fmt.Fprint(g.out, "\t\tWithChildAttribute(&o."+f.Name+")")
		if idx < len(fs)-1 {
			fmt.Fprintln(g.out, ".")
		} else {
			fmt.Fprintln(g.out, "")
		}
	}
	fmt.Fprintln(g.out, "}")
	fmt.Println()

	return nil
}

func (g *Generator) isExternalImport(fieldPkgPath string) bool {
	if len(fieldPkgPath) > 0 {
		fieldPkgPathSplitted := strings.Split(fieldPkgPath, "/")
		curPkgPathSplitted := strings.Split(g.pkgPath, "/")
		if len(curPkgPathSplitted) > 0 {
			return curPkgPathSplitted[0] != fieldPkgPathSplitted[0]
		}
		return true
	}

	return false
}
func (g *Generator) getPakagePath(t reflect.Type) string {
	if t.Name() == "" {
		switch t.Kind() {
		case reflect.Ptr:
			return g.getPakagePath(t.Elem())
		case reflect.Slice:
			return g.getPakagePath(t.Elem())
		case reflect.Array:
			return g.getPakagePath(t.Elem())
		case reflect.Map:
			return g.getPakagePath(t.Elem())
		}
	}

	return t.PkgPath()
}

// getType return the textual type name of given type that can be used in generated code.
func (g *Generator) getType(t reflect.Type) string {
	if t.Name() == "" {
		switch t.Kind() {
		case reflect.Ptr:
			return "*" + g.getType(t.Elem())
		case reflect.Slice:
			return "[]" + g.getType(t.Elem())
		case reflect.Array:
			return "[" + strconv.Itoa(t.Len()) + "]" + g.getType(t.Elem())
		case reflect.Map:
			return "map[" + g.getType(t.Key()) + "]" + g.getType(t.Elem())
		}
	}

	if t.Name() == "" || t.PkgPath() == "" {
		if t.Kind() == reflect.Struct {
			// the fields of an anonymous struct can have named types,
			// and t.String() will not be sufficient because it does not
			// remove the package name when it matches g.pkgPath.
			// so we convert by hand
			nf := t.NumField()
			lines := make([]string, 0, nf)
			for i := 0; i < nf; i++ {
				f := t.Field(i)
				var line string
				if !f.Anonymous {
					line = f.Name + " "
				} // else the field is anonymous (an embedded type)
				line += g.getType(f.Type)
				t := f.Tag
				if t != "" {
					line += " " + escapeTag(t)
				}
				lines = append(lines, line)
			}
			return strings.Join([]string{"struct { ", strings.Join(lines, "; "), " }"}, "")
		}
		return t.String()
	} else if t.PkgPath() == g.pkgPath {
		return t.Name()
	}
	return g.pkgAlias(t.PkgPath()) + "." + t.Name()
}

// escape a struct field tag string back to source code
func escapeTag(tag reflect.StructTag) string {
	t := string(tag)
	if strings.ContainsRune(t, '`') {
		// there are ` in the string; we can't use ` to enclose the string
		return strconv.Quote(t)
	}
	return "`" + t + "`"
}

// pkgAlias creates and returns and import alias for a given package.
func (g *Generator) pkgAlias(pkgPath string) string {
	pkgPath = fixPkgPathVendoring(pkgPath)
	if alias := g.imports[pkgPath]; alias != "" {
		return alias
	}

	for i := 0; ; i++ {
		alias := fixAliasName(path.Base(pkgPath))
		if i > 0 {
			alias += fmt.Sprint(i)
		}

		exists := false
		for _, v := range g.imports {
			if v == alias {
				exists = true
				break
			}
		}

		if !exists {
			g.imports[pkgPath] = alias
			return alias
		}
	}
}

// fixes vendored paths
func fixPkgPathVendoring(pkgPath string) string {
	const vendor = "/vendor/"
	if i := strings.LastIndex(pkgPath, vendor); i != -1 {
		return pkgPath[i+len(vendor):]
	}
	return pkgPath
}

func fixAliasName(alias string) string {
	alias = strings.Replace(
		strings.Replace(alias, ".", "_", -1),
		"-",
		"_",
		-1,
	)

	if alias[0] == 'v' { // to void conflicting with var names, say v1
		alias = "_" + alias
	}
	return alias
}

func getStructFields(t reflect.Type) ([]reflect.StructField, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("got %v; expected a struct", t)
	}

	var efields []reflect.StructField
	var fields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tags := parseFieldTags(f)
		if !f.Anonymous || tags.name != "" {
			continue
		}

		t1 := f.Type
		if t1.Kind() == reflect.Ptr {
			t1 = t1.Elem()
		}

		if t1.Kind() == reflect.Struct {
			fs, err := getStructFields(t1)
			if err != nil {
				return nil, fmt.Errorf("error processing embedded field: %v", err)
			}
			efields = mergeStructFields(efields, fs)
		} else if (t1.Kind() >= reflect.Bool && t1.Kind() < reflect.Complex128) || t1.Kind() == reflect.String {
			if strings.Contains(f.Name, ".") || unicode.IsUpper([]rune(f.Name)[0]) {
				fields = append(fields, f)
			}
		}
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tags := parseFieldTags(f)
		if f.Anonymous && tags.name == "" {
			continue
		}

		c := []rune(f.Name)[0]
		if unicode.IsUpper(c) {
			fields = append(fields, f)
		}
	}
	return mergeStructFields(efields, fields), nil
}

// parseFieldTags parses the json field tag into a structure.
func parseFieldTags(f reflect.StructField) fieldTags {
	var ret fieldTags

	for i, s := range strings.Split(f.Tag.Get("json"), ",") {
		switch {
		case i == 0:
			ret.name = s
		}
	}

	for i, s := range strings.Split(f.Tag.Get("dynamodbav"), ",") {
		switch {
		case i == 0 && len(s) > 0:
			ret.name = s // overwriting json tag name, giving more precedence to dynamodbav
		}
	}

	for _, s := range strings.Split(f.Tag.Get("dynexpr"), ",") {
		switch {
		case s == "partitionKey":
			ret.partitionKey = true
		case s == "sortKey":
			ret.sortKey = true
		}
	}

	return ret
}

func mergeStructFields(fields1, fields2 []reflect.StructField) (fields []reflect.StructField) {
	used := map[string]bool{}
	for _, f := range fields2 {
		used[f.Name] = true
		fields = append(fields, f)
	}

	for _, f := range fields1 {
		if !used[f.Name] {
			fields = append(fields, f)
		}
	}
	return
}

// printHeader prints package declaration and imports.
func (g *Generator) printHeader() {
	fmt.Println("// Code generated by dynexpr for building expression. DO NOT EDIT.")
	fmt.Println()
	fmt.Println("package ", g.pkgName)
	fmt.Println()

	byAlias := make(map[string]string, len(g.imports))
	aliases := make([]string, 0, len(g.imports))

	for path, alias := range g.imports {
		aliases = append(aliases, alias)
		byAlias[alias] = path
	}

	sort.Strings(aliases)
	fmt.Println("import (")
	for _, alias := range aliases {
		fmt.Printf("  %s %q\n", alias, byAlias[alias])
	}

	fmt.Println(")")
	fmt.Println()
}

func NewGenerator(rootStructNames []string) *Generator {
	rStructNames := make(map[string]struct{})
	for _, structName := range rootStructNames {
		if len(structName) > 0 {
			rStructNames[structName] = struct{}{}
		}
	}

	return &Generator{
		imports: map[string]string{
			pkgDynexpr: "dynexpr",
		},
		rootStructNames: rStructNames,
		typesSeen:       make(map[reflect.Type]bool),
	}
}
