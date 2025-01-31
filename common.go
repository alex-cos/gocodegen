package main

import (
	"fmt"
	"io"
)

type Buffer interface {
	io.Writer
	Bytes() []byte
}

func writeString(w io.Writer, str string) error {
	if _, err := w.Write([]byte(str)); err != nil {
		return fmt.Errorf("failed to write into io.Writer: %w", err)
	}
	return nil
}

func writeBuffer(w io.Writer, buf Buffer) error {
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write into io.Writer: %w", err)
	}
	return nil
}
