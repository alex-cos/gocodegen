package main

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const timeTime = "time.Time"

func generate(output string, pack *Package) ([]Generated, error) {
	if output != "" {
		return generateSingle(output, pack)
	}
	return generateMulitple(pack)
}

func generateSingle(output string, pack *Package) ([]Generated, error) {
	gen := Generated{
		Name: output,
		Buf:  &bytes.Buffer{},
	}
	structs := []*StructType{}
	enums := []*Enum{}
	imports := map[string]string{}
	for _, f := range pack.Files {
		structs = append(structs, f.StructTypes...)
		enums = append(enums, f.Enums...)
		maps.Copy(imports, f.Imports)
	}
	err := generateCommon(gen.Buf, pack, imports, structs, enums)
	if err != nil {
		return nil, err
	}
	for _, file := range pack.Files {
		for _, e := range file.Enums {
			err = generateEnum(gen.Buf, e)
			if err != nil {
				return nil, err
			}
		}
		for _, s := range file.StructTypes {
			err = generateStruct(gen.Buf, s)
			if err != nil {
				return nil, err
			}
		}
	}

	return []Generated{gen}, nil
}

func generateMulitple(pack *Package) ([]Generated, error) {
	var (
		err       error
		generated = []Generated{}
	)
	for _, file := range pack.Files {
		for _, enul := range file.Enums {
			gen := Generated{
				Name: fmt.Sprintf("gocodegen_%s.go", strings.ToLower(enul.Name)),
				Buf:  &bytes.Buffer{},
			}
			err = generateCommon(gen.Buf, pack, file.Imports, []*StructType{}, []*Enum{enul})
			if err != nil {
				return nil, err
			}
			err = generateEnum(gen.Buf, enul)
			if err != nil {
				return nil, err
			}
			generated = append(generated, gen)
		}
		for _, structType := range file.StructTypes {
			gen := Generated{
				Name: fmt.Sprintf("gocodegen_%s.go", strings.ToLower(structType.Name)),
				Buf:  &bytes.Buffer{},
			}
			err = generateCommon(gen.Buf, pack, file.Imports, []*StructType{structType}, []*Enum{})
			if err != nil {
				return nil, err
			}
			err = generateStruct(gen.Buf, structType)
			if err != nil {
				return nil, err
			}
			generated = append(generated, gen)
		}
	}

	return generated, nil
}

func generateCommon(
	writer io.Writer,
	pack *Package,
	imports map[string]string,
	structs []*StructType,
	enums []*Enum,
) error {
	name := pack.Name
	if len(enums) == 1 {
		name = enums[0].Name
	}
	if len(structs) == 1 {
		name = structs[0].Name
	}
	err := generateHeader(writer, name)
	if err != nil {
		return err
	}
	err = writeString(writer, fmt.Sprintf("package %s\n\n", pack.Name))
	if err != nil {
		return err
	}
	err = generateImports(writer, imports, structs, enums)
	if err != nil {
		return err
	}

	return nil
}

func generateEnum(w io.Writer, e *Enum) error {
	err := generateEnumStringer(w, e)
	if err != nil {
		return err
	}
	err = generateEnumJSON(w, e)
	if err != nil {
		return err
	}
	return nil
}

func generateStruct(writer io.Writer, structType *StructType) error {
	err := generateBuilder(writer, structType)
	if err != nil {
		return err
	}
	err = generateStructStringer(writer, structType)
	if err != nil {
		return err
	}
	err = generateEqual(writer, structType)
	if err != nil {
		return err
	}
	err = generateGetterSetter(writer, structType)
	if err != nil {
		return err
	}
	err = generateStructJSON(writer, structType)
	if err != nil {
		return err
	}
	return nil
}

func generateHeader(writer io.Writer, name string) error {
	var buff bytes.Buffer

	header, err := template.New("header").Parse(headerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse header template: %w", err)
	}
	err = header.Execute(&buff, struct {
		Version  string
		DateTime string
		TypeName string
	}{
		Version:  version,
		DateTime: time.Now().Format("Mon, 02 Jan 2006 15:04:05"),
		TypeName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to execute header template: %w", err)
	}

	return writeBuffer(writer, &buff)
}

func generateImports(
	writer io.Writer,
	imports map[string]string,
	structTypes []*StructType,
	enums []*Enum,
) error {
	importPackages := []string{}
	isFmt := false
	isEnumJSON := false
	isStructJSON := false
	for _, StructType := range structTypes {
		for _, field := range StructType.Fields {
			if !field.shouldImport() {
				continue
			}
			tab := strings.Split(field.Type, ".")
			if len(tab) != 2 {
				continue
			}
			alias := tab[0]
			path, ok := imports[alias]
			if !ok {
				continue
			}
			imp := fmt.Sprintf("\t\"%s\"", path)
			if path != alias {
				imp = fmt.Sprintf("\t%s \"%s\"", alias, path)
			}
			importPackages = addImportPackages(importPackages, imp)
		}
		if StructType.String {
			isFmt = true
		}
		if StructType.JSON {
			isStructJSON = true
		}
	}
	for _, e := range enums {
		if e.JSON {
			isEnumJSON = true
		}
	}
	if isFmt || isEnumJSON {
		importPackages = addImportPackages(importPackages, "\t\"fmt\"")
	}
	if isEnumJSON || isStructJSON {
		importPackages = addImportPackages(importPackages, "\t\"encoding/json\"")
	}
	if isStructJSON {
		importPackages = addImportPackages(importPackages, "\t\"bytes\"")
	}
	sort.Strings(importPackages)
	if len(importPackages) > 0 {
		t := fmt.Sprintf("import (\n%s\n)\n\n", strings.Join(importPackages, "\n"))
		err := writeString(writer, t)
		if err != nil {
			return err
		}
	}

	return nil
}

func addImportPackages(importPackages []string, name string) []string {
	if !packageAlreadyIn(importPackages, name) {
		importPackages = append(importPackages, name)
	}
	return importPackages
}

func packageAlreadyIn(importPackages []string, name string) bool {
	return slices.Contains(importPackages, name)
}

func generateBuilder(writer io.Writer, mStruct *StructType) error {
	var buff bytes.Buffer

	caser := cases.Title(language.Und, cases.NoLower)
	if !mStruct.New {
		return nil
	}
	params := []string{}
	fields := []string{}
	for _, field := range mStruct.Fields {
		if !field.New {
			continue
		}
		fieldName := field.Name
		fieldType := field.Type
		params = append(params, fmt.Sprintf("%s %s", fieldName, fieldType))
		fields = append(fields, fmt.Sprintf("\t\t%s: %s", fieldName, fieldName))
	}

	body := fmt.Sprintf("return &%s{", mStruct.Name)
	if len(fields) > 0 {
		body += fmt.Sprintf("\n%s,\n", strings.Join(fields, ",\n"))
		body += "\t"
	}
	body += "}"
	builder, err := template.New("builder").Parse(builderTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse builder template: %w", err)
	}
	err = builder.Execute(&buff, struct {
		FuncName string
		Params   string
		TypeName string
		Returned string
		Body     string
	}{
		FuncName: "New" + caser.String(mStruct.Name),
		Params:   strings.Join(params, ", "),
		TypeName: mStruct.Name,
		Returned: "*" + mStruct.Name,
		Body:     body,
	})
	if err != nil {
		return fmt.Errorf("failed to execute builder template: %w", err)
	}

	return writeBuffer(writer, &buff)
}

func generateEnumStringer(writer io.Writer, enum *Enum) error {
	var buff bytes.Buffer

	if !enum.String {
		return nil
	}
	if len(enum.Constants) == 0 {
		return fmt.Errorf("empty enumeration constants for type '%s'", enum.Name)
	}
	stringer, err := template.New("enumstringer").Parse(enumStringerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse enumstringer template: %w", err)
	}
	err = stringer.Execute(&buff, struct {
		TypeName  string
		Constants []string
		Default   string
	}{
		TypeName:  enum.Name,
		Constants: enum.Constants,
		Default:   enum.Constants[0],
	})
	if err != nil {
		return fmt.Errorf("failed to execute enumstringer template: %w", err)
	}

	return writeBuffer(writer, &buff)
}

func generateEnumJSON(writer io.Writer, enum *Enum) error {
	var buff bytes.Buffer

	if !enum.JSON {
		return nil
	}
	stringer, err := template.New("enumjson").Parse(enumJSONTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse enumjson template: %w", err)
	}
	if len(enum.Constants) == 0 {
		return fmt.Errorf("empty enumeration constants for type '%s'", enum.Name)
	}
	err = stringer.Execute(&buff, struct {
		TypeName string
	}{
		TypeName: enum.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to execute enumjson template: %w", err)
	}

	return writeBuffer(writer, &buff)
}

func generateStructStringer(writer io.Writer, mStruct *StructType) error {
	var buff bytes.Buffer

	if !mStruct.String {
		return nil
	}
	fields := []string{}
	formats := []string{}
	for _, field := range mStruct.Fields {
		if !field.String {
			continue
		}
		fieldName := field.Name

		if field.Type == timeTime {
			fields = append(fields, fmt.Sprintf("thiz.%s.UTC().Format(\"2006/01/02 15:04:05\")", fieldName))
		} else {
			fields = append(fields, "thiz."+fieldName)
		}
		formats = append(formats, "%v")
	}
	body := fmt.Sprintf("return fmt.Sprintf(\"%s{%s}\",\n\t\t%s)",
		mStruct.Name,
		strings.Join(formats, ","),
		strings.Join(fields, ",\n\t\t"),
	)
	stringer, err := template.New("stringer").Parse(structStringerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse stringer template: %w", err)
	}
	err = stringer.Execute(&buff, struct {
		TypeName string
		Body     string
	}{
		TypeName: mStruct.Name,
		Body:     body,
	})
	if err != nil {
		return fmt.Errorf("failed to execute stringer template: %w", err)
	}
	return writeBuffer(writer, &buff)
}

func generateEqual(writer io.Writer, mStruct *StructType) error {
	var buff bytes.Buffer

	if !mStruct.Equals {
		return nil
	}
	assertions := []string{}
	for _, field := range mStruct.Fields {
		if !field.Equals {
			continue
		}
		fieldName := field.Name
		if field.Type == timeTime {
			assertions = append(assertions, fmt.Sprintf("thiz.%s.Equal(other.%s)", fieldName, fieldName))
		} else {
			assertions = append(assertions, fmt.Sprintf("thiz.%s == other.%s", fieldName, fieldName))
		}
	}
	body := "return " + strings.Join(assertions, " &&\n\t\t")
	equal, err := template.New("equal").Parse(equalTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse equal template: %w", err)
	}
	err = equal.Execute(&buff, struct {
		TypeName string
		Body     string
	}{
		TypeName: mStruct.Name,
		Body:     body,
	})
	if err != nil {
		return fmt.Errorf("failed to execute equal template: %w", err)
	}
	return writeBuffer(writer, &buff)
}

func generateGetterSetter(writer io.Writer, mStruct *StructType) error {
	var buff bytes.Buffer

	caser := cases.Title(language.Und, cases.NoLower)
	getter, err := template.New("getter").Parse(getterTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse getter template: %w", err)
	}
	setter, err := template.New("setter").Parse(setterTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse setter template: %w", err)
	}
	for _, field := range mStruct.Fields {
		fieldType := field.Type
		if field.Setter {
			buff.Reset()
			err = setter.Execute(&buff, struct {
				TypeName  string
				FuncName  string
				FieldName string
				FieldType string
			}{
				TypeName:  mStruct.Name,
				FuncName:  "Set" + caser.String(field.Name),
				FieldName: field.Name,
				FieldType: fieldType,
			})
			if err != nil {
				return fmt.Errorf("failed to execute setter template: %w", err)
			}
			err := writeBuffer(writer, &buff)
			if err != nil {
				return err
			}
		}
		if field.Getter {
			buff.Reset()
			err = getter.Execute(&buff, struct {
				TypeName  string
				FuncName  string
				FieldName string
				FieldType string
			}{
				TypeName:  mStruct.Name,
				FuncName:  "Get" + caser.String(field.Name),
				FieldName: field.Name,
				FieldType: fieldType,
			})
			if err != nil {
				return fmt.Errorf("failed to execute getter template: %w", err)
			}
			err := writeBuffer(writer, &buff)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func generateStructJSON(writer io.Writer, mStruct *StructType) error {
	var buff bytes.Buffer

	if !mStruct.JSON {
		return nil
	}
	marshal := map[string]GenJSON{}
	for _, field := range mStruct.Fields {
		tag, ok := field.Tags["json"]
		if !ok || tag.Name == "" {
			continue
		}
		marshal[field.Name] = GenJSON{
			Marshal:   field.Name,
			Unmarshal: fmt.Sprintf("thiz.%s = tmp.%s", field.Name, field.Name),
			Type:      field.Type,
			Tag:       tag.Name,
		}
		if field.Type == timeTime {
			name := field.Name + "Epoch"
			marshal[name] = GenJSON{
				Marshal:   field.Name + ".Unix()",
				Unmarshal: fmt.Sprintf("thiz.%s = time.Unix(tmp.%s, 0).UTC()", field.Name, name),
				Type:      "int64",
				Tag:       tag.Name + "Epoch",
			}
		}
	}
	structJSON, err := template.New("structJSON").Parse(structJSONTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse structJSON template: %w", err)
	}
	err = structJSON.Execute(&buff, struct {
		TypeName    string
		MarshalJSON map[string]GenJSON
	}{
		TypeName:    mStruct.Name,
		MarshalJSON: marshal,
	})
	if err != nil {
		return fmt.Errorf("failed to execute structJSON template: %w", err)
	}
	return writeBuffer(writer, &buff)
}
