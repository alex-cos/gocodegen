package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	keyword = "gocodegen"
)

func parsePackage(pkg *packages.Package) *Package {
	pack := &Package{
		Name:  pkg.Name,
		Files: []File{},
	}
	for _, file := range pkg.Syntax {
		f := parseFile(file)
		pack.Files = append(pack.Files, *f)
	}
	return pack
}

func parseFile(file *ast.File) *File {
	f := &File{
		Name:        file.Name.String(),
		Imports:     map[string]string{},
		StructTypes: []*StructType{},
		Enums:       []*Enum{},
	}
	f.Imports = parseImports(file)
	f.Enums = parseEnums(file)
	f.StructTypes = parseStructs(file)

	return f
}

func parseImports(file *ast.File) map[string]string {
	var imports = map[string]string{}
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok != token.IMPORT {
			continue
		}
		for _, spec := range gd.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			path := strings.Trim(importSpec.Path.Value, " \"")
			alias := filepath.Base(path)
			if importSpec.Name != nil {
				alias = strings.Trim(importSpec.Name.String(), " \"")
			}
			imports[alias] = path
		}
	}
	return imports
}

func parseEnums(file *ast.File) []*Enum {
	enums := []*Enum{}

	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok != token.TYPE {
			continue
		}
		comment := gd.Doc.Text()
		if !strings.Contains(comment, keyword) {
			continue
		}
		for _, spec := range gd.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			ident, ok := typeSpec.Type.(*ast.Ident)
			if !ok || typeSpec.Name == nil {
				continue
			}
			name := typeSpec.Name.String()
			enums = append(enums, &Enum{
				Name:      name,
				Type:      strings.Trim(ident.Name, " \""),
				Comment:   comment,
				String:    strings.Contains(comment, "STRING") || strings.Contains(comment, "JSON"),
				JSON:      strings.Contains(comment, "JSON"),
				Constants: parseConstants(file, name),
			})
			break
		}
	}
	return enums
}

func parseConstants(file *ast.File, name string) []string {
	constants := []string{}

	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			typeName := parseType(valueSpec.Type)
			if typeName != name {
				continue
			}
			for _, name := range valueSpec.Names {
				constants = append(constants, name.String())
			}
		}
	}
	return constants
}

func parseStructs(file *ast.File) []*StructType {
	structTypes := []*StructType{}

	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok != token.TYPE {
			continue
		}
		comment := gd.Doc.Text()
		if !strings.Contains(comment, keyword) {
			continue
		}
		for _, spec := range gd.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if ok {
				mStruct := StructType{
					Name:    strings.Trim(typeSpec.Name.Name, " \""),
					Comment: comment,
					Mutex:   nil,
					New:     strings.Contains(comment, "NEW"),
					String:  strings.Contains(comment, "STRING"),
					Equals:  strings.Contains(comment, "EQUAL"),
					JSON:    strings.Contains(comment, "JSON"),
					Fields:  []Field{},
				}
				fields := parseFieldList(structType.Fields)
				for _, field := range fields {
					if !strings.Contains(field.Comment, keyword) {
						continue
					}
					field.New = strings.Contains(field.Comment, "NEW")
					field.String = strings.Contains(field.Comment, "STRING")
					field.Equals = strings.Contains(field.Comment, "EQUAL")
					field.Getter = strings.Contains(field.Comment, "GET")
					field.Setter = strings.Contains(field.Comment, "SET")
					mStruct.Fields = append(mStruct.Fields, field)
				}
				structTypes = append(structTypes, &mStruct)
				break
			}
		}
	}

	return structTypes
}

func parseFieldList(fieldList *ast.FieldList) []Field {
	fields := []Field{}
	if fieldList != nil {
		for _, field := range fieldList.List {
			fType := parseType(field.Type)
			name := ""
			if len(field.Names) > 0 {
				name = field.Names[0].String()
			}
			comment := ""
			if field.Comment != nil {
				comment = field.Comment.Text()
			}
			tag := ""
			if field.Tag != nil {
				tag = field.Tag.Value
			}
			tags, err := parseTag(strings.Trim(tag, "`"))
			if err != nil {
				tags = map[string]Tag{}
			}
			fields = append(fields, Field{
				Name:    name,
				Type:    fType,
				Comment: comment,
				Tags:    tags,
				New:     false,
				String:  false,
				Equals:  false,
				Getter:  false,
				Setter:  false,
			})
		}
	}
	return fields
}

func parseTag(tag string) (map[string]Tag, error) {
	tags := map[string]Tag{}

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 {
			return nil, errors.New("bad syntax for struct tag key")
		}
		if i+1 >= len(tag) || tag[i] != ':' {
			return nil, errors.New("bad syntax for struct tag pair")
		}
		if tag[i+1] != '"' {
			return nil, errors.New("bad syntax for struct tag value")
		}

		key := tag[:i]
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, errors.New("bad syntax for struct tag value")
		}

		qvalue := tag[:i+1]
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil, errors.New("bad syntax for struct tag value")
		}

		res := strings.Split(value, ",")
		name := res[0]
		options := res[1:]
		if len(options) == 0 {
			options = nil
		}

		tags[key] = Tag{
			Key:     key,
			Name:    name,
			Options: options,
		}
	}

	return tags, nil
}

func parseType(expr ast.Expr) string {
	switch tt := expr.(type) {
	case *ast.Ident:
		return tt.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", parseType(tt.X), parseType(tt.Sel))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", parseType(tt.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", parseType(tt.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", parseType(tt.Key), parseType(tt.Value))
	case *ast.ChanType:
		return fmt.Sprintf("chan %s", parseType(tt.Value))
	case *ast.FuncType:
		return parseFuncType(tt)
	case *ast.InterfaceType:
		return parseInterfaceType(tt)
	case *ast.StructType:
		return parseStructType(tt)
	default:
		return ""
	}
}

func parseFuncType(tt *ast.FuncType) string {
	params := []string{}
	fields := parseFieldList(tt.Params)
	for _, field := range fields {
		params = append(params, field.Type)
	}
	p := strings.Join(params, ", ")
	results := []string{}
	fields = parseFieldList(tt.Results)
	for _, field := range fields {
		results = append(results, field.Type)
	}
	r := strings.Join(results, ", ")
	if len(results) > 1 {
		r = fmt.Sprintf("(%s)", r)
	}
	return fmt.Sprintf("func(%s) %s", p, r)
}

func parseInterfaceType(tt *ast.InterfaceType) string {
	fields := []string{}
	fieldList := parseFieldList(tt.Methods)
	for _, field := range fieldList {
		fields = append(fields, fmt.Sprintf("\t%s%s",
			field.Name,
			strings.Replace(field.Type, "func", "", 1)))
	}
	i := strings.Join(fields, "\n")
	if len(fieldList) > 0 {
		i = fmt.Sprintf("interface {\n%s\n}", i)
	} else {
		i = "interface{}"
	}
	return i
}

func parseStructType(tt *ast.StructType) string {
	fields := []string{}
	fieldList := parseFieldList(tt.Fields)
	for _, field := range fieldList {
		fields = append(fields, fmt.Sprintf("\t%s %s", field.Name, field.Type))
	}
	f := strings.Join(fields, "\n")
	return fmt.Sprintf("struct {\n%s\n}", f)
}
