# gocodegen

gocodegen is a command-line tool that generates Go code by analyzing special comments in Go source files. It helps reduce boilerplate code by automatically generating getters, setters, constructors, stringers, equality methods, and JSON marshalers based on struct and enum definitions.

## Features

- Generate getter and setter methods for struct fields
- Create constructor functions (New*)
- Implement String() methods for pretty printing
- Add Equal() methods for struct comparison
- Generate JSON marshaling/unmarshaling methods
- Support for enum-like types with String() and JSON methods
- Customizable output via command-line flags

## Installation

```bash
go install github.com/alex-cos/gocodegen@latest
```

## Usage

Add special comments to your Go code to indicate what should be generated:

```go
// MyStruct gocodegen:NEW,STRING,EQUAL,JSON.
type MyStruct struct {
    Field1 string `json:"field1"`  // gocodegen:STRING,GET,SET,NEW,EQUAL.
    Field2 int    `json:"field2"`  // gocodegen:GET,SET.
}
```

Then run gocodegen:

```bash
gocodegen -output mystruct_gen.go
```

Or use the `go:generate` directive:

```go
//go:generate go run github.com/alex-cos/gocodegen -output mystruct_gen.go
```

## Available Tags

### Struct Tags

- `NEW` - Generate a constructor function
- `STRING` - Generate a String() method
- `EQUAL` - Generate an Equal() method
- `JSON` - Generate MarshalJSON and UnmarshalJSON methods

### Field Tags

- `GET` - Generate a getter method
- `SET` - Generate a setter method
- `NEW` - Include in constructor parameters
- `STRING` - Include in String() method
- `EQUAL` - Include in Equal() comparison

## Example

See the `examples/example1` directory for a complete example:

```go
// Test gocodegen:NEW,STRING,EQUAL,JSON.
type Test struct {
    Date     time.Time                 `json:"date"  toml:"Date"`  // gocodegen:STRING,GET,SET,EQUAL.
    Name     string                    `json:"name"  toml:"Name"`  // gocodegen:STRING,GET,SET,NEW,EQUAL.
    Descr    *string                   `json:"descr" toml:"Descr"` // gocodegen:STRING,GET,SET,NEW.
    // ... more fields
}
```

Running gocodegen on this file will generate:

- NewTest constructor
- String() method
- Equal() method
- MarshalJSON/UnmarshalJSON methods
- Getters and setters for tagged fields

## Output

By default, gocodegen generates files named `gocodegen_<typename>.go` for each type, or a single file if `-output` is specified.
