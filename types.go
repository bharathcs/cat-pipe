package cat_pipe

import (
	"fmt"
)

// RawByteManipulator takes in byte array (without newline) and returns byte array (without newlines). Return empty array to skip, non-nil err to stop.
// out: The line to be added to the writer. If nil or of length 0, no line will be added.
// err: Error to be thrown, execution will stop here.
type RawByteManipulator = func(in []byte) (out []byte, err error)

// LineManipulator takes in string (without newline) and returns strings (without newlines), Return empty string to skip, non-nil err to stop.
// out: The line to be added to the writer. If empty, no line will be added.
// err: Error to be thrown, execution will stop here.
type LineManipulator = func(in string) (out string, err error)

// Struct identifies the progress of the pipe function moving through the reader and writer
type LineCounts struct {
	ReadLineCount    uint
	WrittenLineCount uint
}

func (l LineCounts) String() string {
	return fmt.Sprintf("%d lines read, %d lines written", l.ReadLineCount, l.WrittenLineCount)
}

func NewLineCounts(readLineCount, writtenLineCount uint) LineCounts {
	return LineCounts{ReadLineCount: readLineCount, WrittenLineCount: writtenLineCount}
}

// ReadErrors wrap any non-EOF error encountered while using the reader.
type ReadError struct {
	LineCounts LineCounts
	Err        error
}

func (e *ReadError) Error() string {
	return fmt.Sprintf("execution stopped with %s, due to error from reader %v", e.LineCounts.String(), e.Err)
}

func NewReadError(lc LineCounts, err error) *ReadError {
	return &ReadError{LineCounts: lc, Err: err}
}

// WriteError wrap any error encountered while using the writer.
type WriteError struct {
	LineCounts LineCounts
	Err        error
}

func NewWriteError(lc LineCounts, err error) *WriteError {
	return &WriteError{LineCounts: lc, Err: err}
}

func (e *WriteError) Error() string {
	return fmt.Sprintf("execution stopped with %s, due to error from writer %v", e.LineCounts.String(), e.Err)
}

// MiddleError wrap any error encountered while executing the user passed function.
type MiddleError struct {
	LineCounts LineCounts
	Err        error
}

func (e *MiddleError) Error() string {
	return fmt.Sprintf("execution stopped with %s, due to error from middle function %v", e.LineCounts.String(), e.Err)
}

func NewMiddleError(lc LineCounts, err error) *MiddleError {
	return &MiddleError{LineCounts: lc, Err: err}
}
