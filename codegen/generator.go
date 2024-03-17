package codegen

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mailru/easyjson"
)

// fieldTags contains parsed version of json struct field tags.
type fieldTags struct {
	name string

	dynExprPK   bool
	dynExprSK   bool
	omit        bool
	omitEmpty   bool
	noOmitEmpty bool
	asString    bool
	required    bool
	intern      bool
	noCopy      bool
}

type Generator struct {
	out *bytes.Buffer

	pkgName string
	pkgPath string

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
	// g.marshalers[t] = true
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

		// 	if err := g.genDecoder(t); err != nil {
		// 		return err
		// 	}
		// 	if err := g.genEncoder(t); err != nil {
		// 		return err
		// 	}

		// 	if !g.marshalers[t] {
		// 		continue
		// 	}

		// 	if err := g.genStructMarshaler(t); err != nil {
		// 		return err
		// 	}
		// 	if err := g.genStructUnmarshaler(t); err != nil {
		// 		return err
		// 	}
	}

	// TODO
	g.printHeader()
	// fmt.Println(string(g.out.Bytes()))
	_, err := out.Write(g.out.Bytes())
	return err
}

func (g *Generator) genExpressionBuilder(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return nil // g.genSliceArrayDecoder(t)
	default:
		return g.genStructExpressionBuilder(t)
	}
}

func (g *Generator) genStructExpressionBuilder(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("cannot generate encoder/decoder for %v, not a struct type", t)
	}

	// typ := g.getType(t)
	// fmt.Println("Type: " + typ)

	// get struct name
	structName := t.Name()
	// fmt.Println(structName)

	expressionBldrStructName := structName + "_ExpressionBuilder"
	fmt.Fprintln(g.out, "type "+expressionBldrStructName+" struct {")
	// Init embedded pointer fields.
	// for i := 0; i < t.NumField(); i++ {
	// 	f := t.Field(i)
	// 	if !f.Anonymous || f.Type.Kind() != reflect.Ptr {
	// 		continue
	// 	}

	// 	fmt.Fprintln(g.out, "  out."+f.Name+" = new("+g.getType(f.Type.Elem())+")")
	// }

	// get structs field names
	fs, err := getStructFields(t)
	if err != nil {
		return fmt.Errorf("cannot generate decoder for %v: %v", t, err)
	}

	for _, f := range fs {
		// fmt.Println("Field Name: " + f.Name)
		// fmt.Println("Field Type: " + f.Type.String())
		// fmt.Println("Field Type Details: " + g.getType(f.Type))
		fieldTags := parseFieldTags(f)
		// fmt.Println(fmt.Sprintf("Field Tags: %v", fieldTags))

		// if json tag has dynexpr:"partionKey" this is a partition key attribute or
		// if json tag has dynexpr:"partionKey" this is a partition key attribute
		// and we use DynamoKeyAttribute
		if fieldTags.dynExprPK || fieldTags.dynExprSK {
			fmt.Fprintln(g.out, "\t"+f.Name+"\tDynamoKeyAttribute["+f.Type.String()+"]\t")
		} else if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice { // if type is array/slice then we use DynamoListAttribute
			fmt.Fprintln(g.out, "\t"+f.Name+"\tDynamoListAttribute["+f.Type.String()+"]\t")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Array || f.Type.Elem().Kind() == reflect.Slice) {
			fmt.Fprintln(g.out, "\t"+f.Name+"\tDynamoListAttribute["+f.Type.Elem().Elem().String()+"]\t")
		} else if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Chan || f.Type.Kind() == reflect.Interface {
			return fmt.Errorf("field %s has unsupported type %s", f.Name, f.Type.Kind())
		} else { // if type is struct or native then use DynamoAttribute
			fmt.Fprintln(g.out, "\t"+f.Name+"\tDynamoAttribute["+f.Type.String()+"]\t")
		}
	}
	fmt.Fprintln(g.out, "}")

	// generate function
	fmt.Fprintln(g.out, "func (o *"+expressionBldrStructName+") BuildTree(name string) *DynamoAttribute[*"+expressionBldrStructName+"] {")
	fmt.Fprintln(g.out, "\to = &"+expressionBldrStructName+"{}")
	for _, f := range fs {
		// fmt.Println("Field Name: " + f.Name)
		// fmt.Println("Field Type: " + f.Type.String())
		// fmt.Println("Field Type Details: " + g.getType(f.Type))
		fieldTags := parseFieldTags(f)
		// fmt.Println(fmt.Sprintf("Field Tags: %v", fieldTags))

		// if json tag has dynexpr:"partionKey" this is a partition key attribute or
		// if json tag has dynexpr:"partionKey" this is a partition key attribute
		// and we use DynamoKeyAttribute
		if fieldTags.dynExprPK || fieldTags.dynExprSK {
			fmt.Fprintln(g.out, "o."+f.Name+" = *NewDynamoKeyAttribute["+f.Type.String()+"]().WithName(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice { // if type is array/slice then we use DynamoListAttribute
			fmt.Fprintln(g.out, "o."+f.Name+" = *NewDynamoListAttribute["+f.Type.String()+"]().WithName(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Array || f.Type.Elem().Kind() == reflect.Slice) {
			fmt.Fprintln(g.out, "o."+f.Name+" = *NewDynamoListAttribute["+f.Type.Elem().Elem().String()+"]().WithName(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Chan || f.Type.Kind() == reflect.Interface {
			return fmt.Errorf("field %s has unsupported type %s", f.Name, f.Type.Kind())
		} else if f.Type.Kind() == reflect.Struct {
			fmt.Fprintln(g.out, "o."+f.Name+" = *(&"+f.Type.String()+"_ExpressionBuilder{}).BuildTree(\""+fieldTags.name+"\")")
		} else if f.Type.Kind() == reflect.Pointer && (f.Type.Elem().Kind() == reflect.Struct) {
			fmt.Fprintln(g.out, "o."+f.Name+" = *(&"+f.Type.Elem().String()+"_ExpressionBuilder{}).BuildTree(\""+fieldTags.name+"\")")
		} else { // if type is struct or native then use DynamoAttribute
			fmt.Fprintln(g.out, "o."+f.Name+" = *NewDynamoAttribute["+f.Type.String()+"]().WithName(\""+fieldTags.name+"\")")
		}
	}
	fmt.Fprintln(g.out, "\t return NewDynamoAttribute[*"+expressionBldrStructName+"]().")
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

	// this is easyjson specific
	// for _, f := range fs {
	// 	g.genRequiredFieldSet(t, f)
	// }

	// get structs data type
	// add new imports and aliases
	// generate code
	// generate header

	// for _, f := range fs {
	// 	if err := g.genStructFieldDecoder(t, f); err != nil {
	// 		return err
	// 	}
	// }

	// fsMarshal, _ := json.Marshal(fs)
	// fmt.Println(string(fsMarshal))

	fmt.Println()
	return nil
}

// getType return the textual type name of given type that can be used in generated code.
func (g *Generator) getType(t reflect.Type) string {
	tjson, _ := json.Marshal(t)
	fmt.Println(string(tjson))

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
		case i == 0 && s == "-":
			ret.omit = true
		case i == 0:
			ret.name = s
		case s == "omitempty":
			ret.omitEmpty = true
		case s == "!omitempty":
			ret.noOmitEmpty = true
		case s == "string":
			ret.asString = true
		case s == "required":
			ret.required = true
		case s == "intern":
			ret.intern = true
		case s == "nocopy":
			ret.noCopy = true
		}
	}

	for _, s := range strings.Split(f.Tag.Get("dynexpr"), ",") {
		switch {
		case s == "partitionKey":
			ret.dynExprPK = true
		case s == "sortKey":
			ret.dynExprSK = true
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

func (g *Generator) genRequiredFieldSet(t reflect.Type, f reflect.StructField) {
	tags := parseFieldTags(f)

	if !tags.required {
		return
	}

	fmt.Fprintf(g.out, "var %sSet bool\n", f.Name)
}

func (g *Generator) genStructFieldDecoder(t reflect.Type, f reflect.StructField) error {
	jsonName := "" //g.fieldNamer.GetJSONFieldName(t, f)
	tags := parseFieldTags(f)

	if tags.omit {
		return nil
	}
	if tags.intern && tags.noCopy {
		return errors.New("Mutually exclusive tags are specified: 'intern' and 'nocopy'")
	}

	fmt.Fprintf(g.out, "    case %q:\n", jsonName)
	if err := g.genTypeDecoder(f.Type, "out."+f.Name, tags, 3); err != nil {
		return err
	}

	if tags.required {
		fmt.Fprintf(g.out, "%sSet = true\n", f.Name)
	}

	return nil
}

// genTypeDecoder generates decoding code for the type t, but uses unmarshaler interface if implemented by t.
func (g *Generator) genTypeDecoder(t reflect.Type, out string, tags fieldTags, indent int) error {
	ws := strings.Repeat("  ", indent)

	unmarshalerIface := reflect.TypeOf((*easyjson.Unmarshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(unmarshalerIface) {
		fmt.Fprintln(g.out, ws+"("+out+").UnmarshalEasyJSON(in)")
		return nil
	}

	unmarshalerIface = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(unmarshalerIface) {
		fmt.Fprintln(g.out, ws+"if data := in.Raw(); in.Ok() {")
		fmt.Fprintln(g.out, ws+"  in.AddError( ("+out+").UnmarshalJSON(data) )")
		fmt.Fprintln(g.out, ws+"}")
		return nil
	}

	unmarshalerIface = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	if reflect.PtrTo(t).Implements(unmarshalerIface) {
		fmt.Fprintln(g.out, ws+"if data := in.UnsafeBytes(); in.Ok() {")
		fmt.Fprintln(g.out, ws+"  in.AddError( ("+out+").UnmarshalText(data) )")
		fmt.Fprintln(g.out, ws+"}")
		return nil
	}

	var err error
	// err := g.genTypeDecoderNoCheck(t, out, tags, indent)
	// elem := t.Elem()
	fmt.Fprintln(g.out, ws+"    var "+"tmpVar"+" "+g.getType(t))
	return err
}

// genTypeDecoderNoCheck generates decoding code for the type t.
// func (g *Generator) genTypeDecoderNoCheck(t reflect.Type, out string, tags fieldTags, indent int) error {
// 	ws := strings.Repeat("  ", indent)
// 	// Check whether type is primitive, needs to be done after interface check.
// 	if dec := customDecoders[t.String()]; dec != "" {
// 		fmt.Fprintln(g.out, ws+out+" = "+dec)
// 		return nil
// 	} else if dec := primitiveStringDecoders[t.Kind()]; dec != "" && tags.asString {
// 		if tags.intern && t.Kind() == reflect.String {
// 			dec = "in.StringIntern()"
// 		}
// 		fmt.Fprintln(g.out, ws+out+" = "+g.getType(t)+"("+dec+")")
// 		return nil
// 	} else if dec := primitiveDecoders[t.Kind()]; dec != "" {
// 		if tags.intern && t.Kind() == reflect.String {
// 			dec = "in.StringIntern()"
// 		}
// 		if tags.noCopy && t.Kind() == reflect.String {
// 			dec = "in.UnsafeString()"
// 		}
// 		fmt.Fprintln(g.out, ws+out+" = "+g.getType(t)+"("+dec+")")
// 		return nil
// 	}

// 	switch t.Kind() {
// 	case reflect.Slice:
// 		tmpVar := g.uniqueVarName()
// 		elem := t.Elem()

// 		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
// 			fmt.Fprintln(g.out, ws+"if in.IsNull() {")
// 			fmt.Fprintln(g.out, ws+"  in.Skip()")
// 			fmt.Fprintln(g.out, ws+"  "+out+" = nil")
// 			fmt.Fprintln(g.out, ws+"} else {")
// 			if g.simpleBytes {
// 				fmt.Fprintln(g.out, ws+"  "+out+" = []byte(in.String())")
// 			} else {
// 				fmt.Fprintln(g.out, ws+"  "+out+" = in.Bytes()")
// 			}

// 			fmt.Fprintln(g.out, ws+"}")

// 		} else {

// 			capacity := 1
// 			if elem.Size() > 0 {
// 				capacity = minSliceBytes / int(elem.Size())
// 			}

// 			fmt.Fprintln(g.out, ws+"if in.IsNull() {")
// 			fmt.Fprintln(g.out, ws+"  in.Skip()")
// 			fmt.Fprintln(g.out, ws+"  "+out+" = nil")
// 			fmt.Fprintln(g.out, ws+"} else {")
// 			fmt.Fprintln(g.out, ws+"  in.Delim('[')")
// 			fmt.Fprintln(g.out, ws+"  if "+out+" == nil {")
// 			fmt.Fprintln(g.out, ws+"    if !in.IsDelim(']') {")
// 			fmt.Fprintln(g.out, ws+"      "+out+" = make("+g.getType(t)+", 0, "+fmt.Sprint(capacity)+")")
// 			fmt.Fprintln(g.out, ws+"    } else {")
// 			fmt.Fprintln(g.out, ws+"      "+out+" = "+g.getType(t)+"{}")
// 			fmt.Fprintln(g.out, ws+"    }")
// 			fmt.Fprintln(g.out, ws+"  } else { ")
// 			fmt.Fprintln(g.out, ws+"    "+out+" = ("+out+")[:0]")
// 			fmt.Fprintln(g.out, ws+"  }")
// 			fmt.Fprintln(g.out, ws+"  for !in.IsDelim(']') {")
// 			fmt.Fprintln(g.out, ws+"    var "+tmpVar+" "+g.getType(elem))

// 			if err := g.genTypeDecoder(elem, tmpVar, tags, indent+2); err != nil {
// 				return err
// 			}

// 			fmt.Fprintln(g.out, ws+"    "+out+" = append("+out+", "+tmpVar+")")
// 			fmt.Fprintln(g.out, ws+"    in.WantComma()")
// 			fmt.Fprintln(g.out, ws+"  }")
// 			fmt.Fprintln(g.out, ws+"  in.Delim(']')")
// 			fmt.Fprintln(g.out, ws+"}")
// 		}

// 	case reflect.Array:
// 		iterVar := g.uniqueVarName()
// 		elem := t.Elem()

// 		if elem.Kind() == reflect.Uint8 && elem.Name() == "uint8" {
// 			fmt.Fprintln(g.out, ws+"if in.IsNull() {")
// 			fmt.Fprintln(g.out, ws+"  in.Skip()")
// 			fmt.Fprintln(g.out, ws+"} else {")
// 			fmt.Fprintln(g.out, ws+"  copy("+out+"[:], in.Bytes())")
// 			fmt.Fprintln(g.out, ws+"}")

// 		} else {

// 			length := t.Len()

// 			fmt.Fprintln(g.out, ws+"if in.IsNull() {")
// 			fmt.Fprintln(g.out, ws+"  in.Skip()")
// 			fmt.Fprintln(g.out, ws+"} else {")
// 			fmt.Fprintln(g.out, ws+"  in.Delim('[')")
// 			fmt.Fprintln(g.out, ws+"  "+iterVar+" := 0")
// 			fmt.Fprintln(g.out, ws+"  for !in.IsDelim(']') {")
// 			fmt.Fprintln(g.out, ws+"    if "+iterVar+" < "+fmt.Sprint(length)+" {")

// 			if err := g.genTypeDecoder(elem, "("+out+")["+iterVar+"]", tags, indent+3); err != nil {
// 				return err
// 			}

// 			fmt.Fprintln(g.out, ws+"      "+iterVar+"++")
// 			fmt.Fprintln(g.out, ws+"    } else {")
// 			fmt.Fprintln(g.out, ws+"      in.SkipRecursive()")
// 			fmt.Fprintln(g.out, ws+"    }")
// 			fmt.Fprintln(g.out, ws+"    in.WantComma()")
// 			fmt.Fprintln(g.out, ws+"  }")
// 			fmt.Fprintln(g.out, ws+"  in.Delim(']')")
// 			fmt.Fprintln(g.out, ws+"}")
// 		}

// 	case reflect.Struct:
// 		dec := g.getDecoderName(t)
// 		g.addType(t)

// 		if len(out) > 0 && out[0] == '*' {
// 			// NOTE: In order to remove an extra reference to a pointer
// 			fmt.Fprintln(g.out, ws+dec+"(in, "+out[1:]+")")
// 		} else {
// 			fmt.Fprintln(g.out, ws+dec+"(in, &"+out+")")
// 		}

// 	case reflect.Ptr:
// 		fmt.Fprintln(g.out, ws+"if in.IsNull() {")
// 		fmt.Fprintln(g.out, ws+"  in.Skip()")
// 		fmt.Fprintln(g.out, ws+"  "+out+" = nil")
// 		fmt.Fprintln(g.out, ws+"} else {")
// 		fmt.Fprintln(g.out, ws+"  if "+out+" == nil {")
// 		fmt.Fprintln(g.out, ws+"    "+out+" = new("+g.getType(t.Elem())+")")
// 		fmt.Fprintln(g.out, ws+"  }")

// 		if err := g.genTypeDecoder(t.Elem(), "*"+out, tags, indent+1); err != nil {
// 			return err
// 		}

// 		fmt.Fprintln(g.out, ws+"}")

// 	case reflect.Map:
// 		key := t.Key()
// 		keyDec, ok := primitiveStringDecoders[key.Kind()]
// 		if !ok && !hasCustomUnmarshaler(key) {
// 			return fmt.Errorf("map type %v not supported: only string and integer keys and types implementing json.Unmarshaler are allowed", key)
// 		} // else assume the caller knows what they are doing and that the custom unmarshaler performs the translation from string or integer keys to the key type
// 		elem := t.Elem()
// 		tmpVar := g.uniqueVarName()
// 		keepEmpty := tags.required || tags.noOmitEmpty || (!g.omitEmpty && !tags.omitEmpty)

// 		fmt.Fprintln(g.out, ws+"if in.IsNull() {")
// 		fmt.Fprintln(g.out, ws+"  in.Skip()")
// 		fmt.Fprintln(g.out, ws+"} else {")
// 		fmt.Fprintln(g.out, ws+"  in.Delim('{')")
// 		if !keepEmpty {
// 			fmt.Fprintln(g.out, ws+"  if !in.IsDelim('}') {")
// 		}
// 		fmt.Fprintln(g.out, ws+"  "+out+" = make("+g.getType(t)+")")
// 		if !keepEmpty {
// 			fmt.Fprintln(g.out, ws+"  } else {")
// 			fmt.Fprintln(g.out, ws+"  "+out+" = nil")
// 			fmt.Fprintln(g.out, ws+"  }")
// 		}

// 		fmt.Fprintln(g.out, ws+"  for !in.IsDelim('}') {")
// 		// NOTE: extra check for TextUnmarshaler. It overrides default methods.
// 		if reflect.PtrTo(key).Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
// 			fmt.Fprintln(g.out, ws+"    var key "+g.getType(key))
// 			fmt.Fprintln(g.out, ws+"if data := in.UnsafeBytes(); in.Ok() {")
// 			fmt.Fprintln(g.out, ws+"  in.AddError(key.UnmarshalText(data) )")
// 			fmt.Fprintln(g.out, ws+"}")
// 		} else if keyDec != "" {
// 			fmt.Fprintln(g.out, ws+"    key := "+g.getType(key)+"("+keyDec+")")
// 		} else {
// 			fmt.Fprintln(g.out, ws+"    var key "+g.getType(key))
// 			if err := g.genTypeDecoder(key, "key", tags, indent+2); err != nil {
// 				return err
// 			}
// 		}

// 		fmt.Fprintln(g.out, ws+"    in.WantColon()")
// 		fmt.Fprintln(g.out, ws+"    var "+tmpVar+" "+g.getType(elem))

// 		if err := g.genTypeDecoder(elem, tmpVar, tags, indent+2); err != nil {
// 			return err
// 		}

// 		fmt.Fprintln(g.out, ws+"    ("+out+")[key] = "+tmpVar)
// 		fmt.Fprintln(g.out, ws+"    in.WantComma()")
// 		fmt.Fprintln(g.out, ws+"  }")
// 		fmt.Fprintln(g.out, ws+"  in.Delim('}')")
// 		fmt.Fprintln(g.out, ws+"}")

// 	case reflect.Interface:
// 		if t.NumMethod() != 0 {
// 			if g.interfaceIsEasyjsonUnmarshaller(t) {
// 				fmt.Fprintln(g.out, ws+out+".UnmarshalEasyJSON(in)")
// 			} else if g.interfaceIsJsonUnmarshaller(t) {
// 				fmt.Fprintln(g.out, ws+out+".UnmarshalJSON(in.Raw())")
// 			} else {
// 				return fmt.Errorf("interface type %v not supported: only interface{} and easyjson/json Unmarshaler are allowed", t)
// 			}
// 		} else {
// 			fmt.Fprintln(g.out, ws+"if m, ok := "+out+".(easyjson.Unmarshaler); ok {")
// 			fmt.Fprintln(g.out, ws+"m.UnmarshalEasyJSON(in)")
// 			fmt.Fprintln(g.out, ws+"} else if m, ok := "+out+".(json.Unmarshaler); ok {")
// 			fmt.Fprintln(g.out, ws+"_ = m.UnmarshalJSON(in.Raw())")
// 			fmt.Fprintln(g.out, ws+"} else {")
// 			fmt.Fprintln(g.out, ws+"  "+out+" = in.Interface()")
// 			fmt.Fprintln(g.out, ws+"}")
// 		}
// 	default:
// 		return fmt.Errorf("don't know how to decode %v", t)
// 	}
// 	return nil

// }

// printHeader prints package declaration and imports.
func (g *Generator) printHeader() {
	// if g.buildTags != "" {
	// 	fmt.Println("// +build ", g.buildTags)
	// 	fmt.Println()
	// }
	fmt.Println("// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.")
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
	// fmt.Println("")
	// fmt.Println("// suppress unused package warning")
	// fmt.Println("var (")
	// fmt.Println("   _ *json.RawMessage")
	// fmt.Println("   _ *jlexer.Lexer")
	// fmt.Println("   _ *jwriter.Writer")
	// fmt.Println("   _ easyjson.Marshaler")
	// fmt.Println(")")

	fmt.Println()
}

func NewGenerator() *Generator {
	return &Generator{
		imports:   make(map[string]string),
		typesSeen: make(map[reflect.Type]bool),
	}
}
