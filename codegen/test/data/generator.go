package data

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type Generator struct {
	out *bytes.Buffer

	// package path to local alias map for tracking imports
	imports map[string]string

	// types that encoders were already generated for
	typesSeen map[reflect.Type]bool

	// types that encoders were requested for (e.g. by encoders of other types)
	typesUnseen []reflect.Type
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
	// g.printHeader()
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

	// get struct name
	// get structs field names
	// get structs data type
	// add new imports and aliases
	// generate code
	// generate header
	return nil
}

func NewGenerator() *Generator {
	return &Generator{
		imports:   make(map[string]string),
		typesSeen: make(map[reflect.Type]bool),
	}
}
