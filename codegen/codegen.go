package codegen

import (
	"dynexpr/parser"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mailru/easyjson/bootstrap"
)

func Generate(fname string) (err error) {
	fInfo, err := os.Stat(fname)
	if err != nil {
		return err
	}

	p := parser.Parser{}
	if err := p.Parse(fname, fInfo.IsDir()); err != nil {
		return fmt.Errorf("Error parsing %v: %v", fname, err)
	}

	var outName string
	if fInfo.IsDir() {
		outName = filepath.Join(fname, p.PkgName+"_easyjson.go")
	} else {
		if s := strings.TrimSuffix(fname, ".go"); s == fname {
			return errors.New("Filename must end in '.go'")
		} else {
			outName = s + "_easyjson.go"
		}
	}

	// if *specifiedName != "" {
	// 	outName = *specifiedName
	// }

	// var trimmedBuildTags string
	// if *buildTags != "" {
	// 	trimmedBuildTags = strings.TrimSpace(*buildTags)
	// }

	// var trimmedGenBuildFlags string
	// if *genBuildFlags != "" {
	// 	trimmedGenBuildFlags = strings.TrimSpace(*genBuildFlags)
	// }

	g := bootstrap.Generator{
		// BuildTags:                trimmedBuildTags,
		// GenBuildFlags:            trimmedGenBuildFlags,
		PkgPath: p.PkgPath,
		PkgName: p.PkgName,
		Types:   p.StructNames,
		// SnakeCase:                *snakeCase,
		// LowerCamelCase:           *lowerCamelCase,
		// NoStdMarshalers:          *noStdMarshalers,
		// DisallowUnknownFields:    *disallowUnknownFields,
		// SkipMemberNameUnescaping: *skipMemberNameUnescaping,
		// OmitEmpty:                *omitEmpty,
		// LeaveTemps:               *leaveTemps,
		OutName: outName,
		// StubsOnly:                *stubs,
		// NoFormat:                 *noformat,
		// SimpleBytes:              *simpleBytes,
	}

	if err := g.Run(); err != nil {
		return fmt.Errorf("Bootstrap failed: %v", err)
	}
	return nil
}
