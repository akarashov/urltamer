// Command reset - code generator for Reset() methods based on go/types and text/template.
//
// This tool scans all Go files in the current module for struct types that have a
// specific marker comment (// generate:reset) and generates a Reset() method for each of those structs.
// The generated Reset() method will reset all fields of the struct to their zero values,
// with special handling for slices, maps, pointers, and nested structs that also have Reset() methods.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

const (
	resetMarker = "// generate:reset"
	genFileName = "reset.gen.go"
)

var funcTpl = `
func ({{.Receiver}} *{{.TypeName}}) Reset() {
	if {{.Receiver}} == nil {
		return
	}
{{range .Fields}}
	{{- if .IsPrimitive}}
	{{.Receiver}}.{{.Name}} = *new({{.TypeName}})
	{{- else if .IsSlice}}
	{{.Receiver}}.{{.Name}} = {{.Receiver}}.{{.Name}}[:0]
	{{- else if .IsMap}}
	clear({{.Receiver}}.{{.Name}})
	{{- else if .IsPointerToPrimitive}}
	if {{.Receiver}}.{{.Name}} != nil {
		*{{.Receiver}}.{{.Name}} = *new({{.TypeName}})
	}
	{{- else if .IsPointerToResettable}}
	if {{.Receiver}}.{{.Name}} != nil {
		{{.Receiver}}.{{.Name}}.Reset()
	}
	{{- else if .IsResettableValue}}
	(&{{.Receiver}}.{{.Name}}).Reset()
	{{- else if .IsPointerToNonResettable}}
	{{- end}}
{{end}}
}
`

var tmpl = template.Must(template.New("reset").Parse(funcTpl))

// fieldInfo info about a struct field, used for generating Reset() method
type fieldInfo struct {
	Name                     string // field name
	TypeName                 string // type name
	IsPrimitive              bool   // int, string, bool, float64 etc.
	IsSlice                  bool   // []T
	IsMap                    bool   // map[K]V
	IsPointerToPrimitive     bool   // *string, *int
	IsPointerToResettable    bool   // *MyStruct with Reset()
	IsResettableValue        bool   // MyStruct with Reset()
	IsPointerToNonResettable bool   // *MyStruct without Reset()
}

// structInfo info about a struct type, used for generating Reset() method
type structInfo struct {
	PkgName  string      // package name
	TypeName string      // struct type name
	Receiver string      // receiver name first letter of struct name in lowercase
	Fields   []fieldInfo // list of fields with their info
}

func main() {
	log.Println("Starting reset generator (go/types + template)...")
	pkgs, err := loadAllPackages()
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}
	pkgStructs, pkgNames := collectStructs(pkgs)
	for pkgDir, structs := range pkgStructs {
		if len(structs) > 0 {
			err := generateFile(pkgDir, pkgNames[pkgDir], structs)
			if err != nil {
				log.Printf("Failed to generate file for %s: %v", pkgDir, err)
			} else {
				log.Printf("Successfully generated %s in %s", genFileName, pkgDir)
			}
		}
	}
	log.Println("Reset generator finished.")
}

// loadAllPackages loads all Go packages in the current module using go/packages with syntax and type info.
func loadAllPackages() ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:   ".",
		Tests: false,
	}
	return packages.Load(cfg, "./...")
}

// collectStructs scans all packages and collects information about struct types that have the reset marker comment.
func collectStructs(pkgs []*packages.Package) (map[string][]structInfo, map[string]string) {
	pkgStructs := make(map[string][]structInfo)
	pkgNames := make(map[string]string)
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				log.Printf("Skipping package %s due to error: %v", pkg.PkgPath, e)
			}
			continue
		}
		var pkgDir string
		for i, fileAST := range pkg.Syntax {
			if len(pkg.GoFiles) > i {
				pkgDir = filepath.Dir(pkg.GoFiles[i])
				if strings.HasSuffix(pkg.GoFiles[i], genFileName) {
					continue
				}
			}
			processFileAST(pkg, fileAST, pkgDir, pkgStructs, pkgNames)
		}
	}
	return pkgStructs, pkgNames
}

// processFileAST inspects a single file AST and appends discovered struct info
// into the provided maps. This keeps collectStructs small and focused on
// iterating packages/files.
func processFileAST(pkg *packages.Package, fileAST *ast.File, pkgDir string, pkgStructs map[string][]structInfo, pkgNames map[string]string) {
	// Find all marked TypeSpecs in the file and convert each into structInfo.
	specs := findMarkedTypeSpecs(fileAST)
	for _, ts := range specs {
		si, ok := typeSpecToStructInfo(pkg, ts, pkgDir)
		if !ok {
			continue
		}
		log.Printf("Found marker on struct %s in %s", ts.Name.Name, pkg.PkgPath)
		pkgStructs[pkgDir] = append(pkgStructs[pkgDir], si)
		pkgNames[pkgDir] = pkg.Name
	}
}

// findMarkedTypeSpecs returns all *ast.TypeSpec nodes in the file that are
// declared under a GenDecl which has the reset marker comment.
func findMarkedTypeSpecs(fileAST *ast.File) []*ast.TypeSpec {
	var specs []*ast.TypeSpec
	ast.Inspect(fileAST, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}
		if !hasResetMarker(genDecl.Doc) {
			return true
		}
		for _, spec := range genDecl.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				specs = append(specs, ts)
			}
		}
		return true
	})
	return specs
}

// typeSpecToStructInfo converts a TypeSpec (with known package info) into
// a structInfo. Returns ok=false when conversion is not possible.
func typeSpecToStructInfo(pkg *packages.Package, typeSpec *ast.TypeSpec, pkgDir string) (structInfo, bool) {
	obj := pkg.TypesInfo.Defs[typeSpec.Name]
	if obj == nil {
		log.Printf("Warning: could not find type info for %s", typeSpec.Name.Name)
		return structInfo{}, false
	}
	structType, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return structInfo{}, false
	}
	si := structInfo{
		PkgName:  pkg.Name,
		TypeName: typeSpec.Name.Name,
		Receiver: strings.ToLower(string(typeSpec.Name.Name[0])),
		Fields:   parseStructFields(structType),
	}
	return si, true
}

// hasResetMarker checks if the given comment group contains the reset marker comment.
func hasResetMarker(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, comment := range doc.List {
		if strings.TrimSpace(comment.Text) == resetMarker {
			return true
		}
	}
	return false
}

// parseStructFields analyzes the fields of a struct and returns a slice of fieldInfo with details about each field.
func parseStructFields(s *types.Struct) []fieldInfo {
	var fields []fieldInfo
	for field := range s.Fields() {
		if field.Embedded() {
			continue
		}
		fi := fieldInfo{
			Name: field.Name(),
		}
		analyzeFieldType(&fi, field.Type())
		fields = append(fields, fi)
	}
	return fields
}

// analyzeFieldType fills the fieldInfo based on the type of the field,
// checking if it's a primitive, slice, map, pointer, or has a Reset() method.
func analyzeFieldType(fi *fieldInfo, typ types.Type) {
	named, ok := typ.(*types.Named)
	if ok {
		fi.TypeName = named.Obj().Name()
	}
	if ptr, ok := typ.(*types.Pointer); ok {
		analyzeFieldType(fi, ptr.Elem())
		if fi.IsPrimitive {
			fi.IsPrimitive = false
			fi.IsPointerToPrimitive = true
		} else if fi.IsResettableValue {
			fi.IsResettableValue = false
			fi.IsPointerToResettable = true
		} else {
			fi.IsPointerToNonResettable = true
		}
		return
	}
	switch t := typ.Underlying().(type) {
	case *types.Basic:
		fi.IsPrimitive = true
		if fi.TypeName == "" {
			fi.TypeName = t.Name()
		}
	case *types.Slice:
		fi.IsSlice = true
	case *types.Map:
		fi.IsMap = true
	case *types.Named, *types.Struct:
		ptrType := types.NewPointer(typ)
		if hasResetMethod(ptrType) {
			fi.IsResettableValue = true
		}
		if fi.TypeName == "" && t.String() == "struct{}" {
			return
		}
	}
}

// hasResetMethod checks if the given type has a Reset() method with the correct signature.
func hasResetMethod(typ types.Type) bool {
	obj, _, _ := types.LookupFieldOrMethod(typ, true, nil, "Reset")
	if obj == nil {
		return false
	}
	meth, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	sig, ok := meth.Type().(*types.Signature)
	if !ok {
		return false
	}
	recv := sig.Recv()
	if recv == nil {
		return false
	}
	_, isPtr := recv.Type().(*types.Pointer)
	isPtrRecv := types.IsInterface(recv.Type()) || isPtr
	return isPtrRecv && sig.Params().Len() == 0 && sig.Results().Len() == 0
}

// generateFile generates the reset.gen.go file in the specified directory for the given package name and struct information.
func generateFile(dir string, pkgName string, structs []structInfo) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// Code generated by cmd/reset. DO NOT EDIT.\n\npackage %s\n\n", pkgName)
	for _, s := range structs {
		err := tmpl.Execute(&buf, s)
		if err != nil {
			return fmt.Errorf("error template rendering for %s: %w", s.TypeName, err)
		}
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("--- FORMATTING ERROR (%s) ---\n%s\n----------------------------", dir, buf.String())
		return fmt.Errorf("error code formatting for %s: %w", dir, err)
	}
	outputPath := filepath.Join(dir, genFileName)
	err = os.WriteFile(outputPath, formatted, 0644)
	if err != nil {
		return fmt.Errorf("error file write to path %s: %w", outputPath, err)
	}
	return nil
}
