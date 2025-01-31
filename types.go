package main

type Package struct {
	Name  string
	Files []File
}

type File struct {
	Name        string
	Imports     map[string]string
	StructTypes []*StructType
	Enums       []*Enum
}

type Enum struct {
	Name      string
	Type      string
	Comment   string
	String    bool
	JSON      bool
	Constants []string
}

type StructType struct {
	Name    string
	Comment string
	Mutex   *string
	New     bool
	String  bool
	Equals  bool
	JSON    bool
	Fields  []Field
}

type Field struct {
	Name    string
	Type    string
	Comment string
	Tags    map[string]Tag
	New     bool
	String  bool
	Equals  bool
	Getter  bool
	Setter  bool
}

type Tag struct {
	Key     string
	Name    string
	Options []string
}

type Generated struct {
	Name string
	Buf  Buffer
}

type GenJSON struct {
	Marshal   string
	Unmarshal string
	Type      string
	Tag       string
}

func (thiz Field) shouldImport() bool {
	return thiz.Getter || thiz.Setter || thiz.New
}
