//nolint:deadcode,structcheck,unused
package main

import (
	"sync"

	"time"
)

//go:generate go run github.com/alex-cos/gocodegen -output gocodegen_test.go

// Animal gocodegen:JSON.
type Animal int

const (
	Dog Animal = 1
	Cat Animal = 2
)

// Test gocodegen:NEW,STRING,EQUAL,JSON.
type Test struct {
	Date     time.Time                 `json:"date"  toml:"Date"`  // gocodegen:STRING,GET,SET,EQUAL.
	Name     string                    `json:"name"  toml:"Name"`  // gocodegen:STRING,GET,SET,NEW,EQUAL.
	Descr    *string                   `json:"descr" toml:"Descr"` // gocodegen:STRING,GET,SET,NEW.
	ID       int                       `json:"id"    toml:"ID"`    // gocodegen:GET,EQUAL.
	array    []string                  // gocodegen:GET,SET.
	mmap     map[string]float32        // gocodegen:GET,SET.
	channel  chan int                  // gocodegen:GET,SET.
	function func(int, string) float64 // gocodegen:GET,SET.
	it       interface{}               // gocodegen:GET,SET.
	st       struct {
		a string
		b int
	} // gocodegen:GET,SET.
	hide float32
	mu   *sync.Mutex
}
