package cat_pipe

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

var basicMiddleFn = func(sc *bufio.Scanner, wr *bufio.Writer, lc *LineCounts) (err error) {
	txt := sc.Text()
	if len(txt) > 0 {

		_, err = wr.WriteString(txt + "\n")
		wr.Flush()
		lc.WrittenLineCount++
	}
	return err
}

func createSkipEvenLinesMiddleFn() middleFunction{
	lineNum := 0

	return func(sc *bufio.Scanner, wr *bufio.Writer, lc *LineCounts) (err error) {

	txt := sc.Text()
	if len(txt) > 0 {
		lineNum++
		if lineNum%2==0 {
			return nil
		}

		_, err = wr.WriteString(txt + "\n")
		wr.Flush()
		lc.WrittenLineCount++
	}
	return err
	}
}

func TestPipeWithBytes(t *testing.T) {
	type args struct {
		r      io.Reader
		middle RawByteManipulator
	}
	tests := []struct {
		name    string
		args    args
		want    LineCounts
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got, err := PipeWithBytes(tt.args.r, w, tt.args.middle)
			if (err != nil) != tt.wantErr {
				t.Errorf("PipeWithBytes() error = '%v', wantErr '%v'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PipeWithBytes() = '%v', want '%v'", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("PipeWithBytes() = '%v', want '%v'", gotW, tt.wantW)
			}
		})
	}
}

func TestPipe(t *testing.T) {
	type args struct {
		r      io.Reader
		middle LineManipulator
	}
	tests := []struct {
		name    string
		args    args
		want    LineCounts
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got, err := Pipe(tt.args.r, w, tt.args.middle)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pipe() error = '%v', wantErr '%v'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipe() = '%v', want '%v'", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Pipe() = '%v', want '%v'", gotW, tt.wantW)
			}
		})
	}
}

func Test_pipe_happyPath(t *testing.T) {
	type args struct {
		r      io.Reader
		middle middleFunction
	}
	tests := []struct {
		name    string
		args    args
		want    LineCounts
		wantW   string
		wantErr bool
	}{
		{
			name: "basic - no line",
			args: args{
				strings.NewReader(""),
				basicMiddleFn,
			},
			want:    LineCounts{0, 0},
			wantW:   "",
			wantErr: false,
		},
		{
			name: "basic - one line",
			args: args{
				strings.NewReader("foo fighters\n"),
				basicMiddleFn,
			},
			want:    LineCounts{1, 1},
			wantW:   "foo fighters\n",
			wantErr: false,
		},
		{
			name: "basic - multiple lines",
			args: args{
				strings.NewReader("foo fighters\narctic monkeys\nlime cordiale\n"),
				basicMiddleFn,
			},
			want:    LineCounts{3, 3},
			wantW:   "foo fighters\narctic monkeys\nlime cordiale\n",
			wantErr: false,
		},
		{
			name: "basic - one line without newline",
			args: args{
				strings.NewReader("foo fighters"),
				basicMiddleFn,
			},
			want:    LineCounts{1, 1},
			wantW:   "foo fighters\n",
			wantErr: false,
		},
		{
			name: "basic - multiple lines without newline",
			args: args{
				strings.NewReader("foo fighters\narctic monkeys\nlime cordiale"),
				basicMiddleFn,
			},
			want:    LineCounts{3, 3},
			wantW:   "foo fighters\narctic monkeys\nlime cordiale\n",
			wantErr: false,
		},
		{
			name: "basic - special middle function",
			args: args{
				strings.NewReader("foo fighters\narctic monkeys\nlime cordiale"),
				createSkipEvenLinesMiddleFn(),
			},
			want:    LineCounts{3, 2},
			wantW:   "foo fighters\nlime cordiale\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got, err := pipe(tt.args.r, w, tt.args.middle)
			if (err != nil) != tt.wantErr {
				t.Errorf("pipe() error = '%v', wantErr '%v'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pipe() = '%v', want '%v'", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("pipe() = '%v', want '%v'", gotW, tt.wantW)
			}
		})
	}
}

type readerWithError struct {
	err string
}

func (r *readerWithError) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf(r.err)
}

func Test_pipe_errorHandling(t *testing.T) {
	type args struct {
		r      io.Reader
		w				io.Writer
		middle middleFunction
	}
	tests := []struct {
		name    string
		args    args
		want    LineCounts
		wantW   string
		wantErr error
	}{
		{
			name: "middle / writer function error",
			args: args{
				strings.NewReader("foo fighters\narctic monkeys\nlime cordiale\n"),
				&bytes.Buffer{},
				func(sc *bufio.Scanner, wr *bufio.Writer, lc *LineCounts) (err error) {
					lc.ReadLineCount = 5
					lc.WrittenLineCount = 8
					return fmt.Errorf("mayday mayday mayday")
				},
			},
			want:    LineCounts{5, 8},
			wantW:   "",
			wantErr: fmt.Errorf("mayday mayday mayday") ,
		},
		{
			name: "reader function error",
			args: args{
				&readerWithError{"reading failed here"},
				&bytes.Buffer{},
				basicMiddleFn,
			},
			want:    LineCounts{0, 0},
			wantW:   "",
			wantErr: fmt.Errorf("execution stopped with 0 lines read, 0 lines written, due to error from reader reading failed here") ,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got, err := pipe(tt.args.r, w, tt.args.middle)
			if err == nil || err.Error() != tt.wantErr.Error() {
				t.Errorf("pipe() error = '%+v', wantErr '%+v'", err.Error(), tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pipe() = '%v', want '%v'", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("pipe() = '%v', want '%v'", gotW, tt.wantW)
			}
		})
	}
}

func pipeStub(input string, fn middleFunction) (string, LineCounts, error) {
	sc := bufio.NewScanner(strings.NewReader(input))
	w := &bytes.Buffer{}
	wr := bufio.NewWriter(w)
	lc := LineCounts{0,0}
	var err error
	for sc.Scan() {
		lc.ReadLineCount++
		err = fn(sc, wr, &lc)
		wr.Flush()
		if err != nil {
			return w.String(), lc, err
		}
	}
	return w.String(), lc, err
}

func Test_convertRawByteManipulator(t *testing.T) {
	type args struct {
		middle RawByteManipulator
		in []byte
	}
	tests := []struct {
		name string
		args args
		wantOut string
		wantLineCounts LineCounts
		wantErr error
	}{
		{
			name: "basic - one line",
			args: args{
				middle: func(in []byte) ([]byte, error) {
					return []byte("arctic monkeys"), nil
				},
				in: []byte("foo fighters\n"),
			},
			wantOut: "arctic monkeys\n",
			wantLineCounts: LineCounts{1,1},
			wantErr: nil,
		},
		{
			name: "basic - three line",
			args: args{
				middle: func(in []byte) ([]byte, error) {
					return in, nil
				},
				in: []byte("foo fighters\narctic monkeys\nlime cordiale\n"),
			},
			wantOut: "foo fighters\narctic monkeys\nlime cordiale\n",
			wantLineCounts: LineCounts{3,3},
			wantErr: nil,
		},
		{
			name: "skip writes",
			args: args{
				middle: func(in []byte) ([]byte, error) {
					return nil, nil
				},
				in: []byte("foo fighters\narctic monkeys\nlime cordiale\n"),
			},
			wantOut: "",
			wantLineCounts: LineCounts{3,0},
			wantErr: nil,
		},
		{
			name: "return error",
			args: args{
				middle: func(in []byte) ([]byte, error) {
					return []byte("foo fighters"), fmt.Errorf("Failed here")
				},
				in: []byte("foo fighters\narctic monkeys\nlime cordiale\n"),
			},
			wantOut: "",
			wantLineCounts: LineCounts{1,0},
			wantErr: fmt.Errorf("execution stopped with 1 lines read, 0 lines written, due to error from middle function Failed here"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, lc, err := pipeStub(string(tt.args.in), convertRawByteManipulator(tt.args.middle))
			if out != tt.wantOut {
				t.Errorf("convertRawByteManipulator() out = '%v', want '%v'", out, tt.wantOut)
			}
			if lc != tt.wantLineCounts {
				t.Errorf("convertRawByteManipulator() line counts = '%v', want '%v'", lc, tt.wantLineCounts)
			}
			if (err == nil && tt.wantErr != nil) && err.Error() != tt.wantErr.Error() {
				t.Errorf("convertRawByteManipulator() error = '%v', want '%v'", err, tt.wantErr)
			}
		})
	}
}

func Test_convertLineManipulator(t *testing.T) {
	type args struct {
		middle LineManipulator
		in string
	}
	tests := []struct {
		name string
		args args
		wantOut string
		wantLineCounts LineCounts
		wantErr error
	}{
		{
			name: "basic - one line",
			args: args{
				middle: func(in string) (string, error) {
					return "arctic monkeys", nil
				},
				in: "foo fighters\n",
			},
			wantOut: "arctic monkeys\n",
			wantLineCounts: LineCounts{1,1},
			wantErr: nil,
		},
		{
			name: "basic - three line",
			args: args{
				middle: func(in string) (string, error) {
					return in, nil
				},
				in: "foo fighters\narctic monkeys\nlime cordiale\n",
			},
			wantOut: "foo fighters\narctic monkeys\nlime cordiale\n",
			wantLineCounts: LineCounts{3,3},
			wantErr: nil,
		},
		{
			name: "skip writes",
			args: args{
				middle: func(in string) (string, error) {
					return "", nil
				},
				in: "foo fighters\narctic monkeys\nlime cordiale\n",
			},
			wantOut: "",
			wantLineCounts: LineCounts{3,0},
			wantErr: nil,
		},
		{
			name: "return error",
			args: args{
				middle: func(in string) (string, error) {
					return "foo fighters", fmt.Errorf("Failed here")
				},
				in: "foo fighters\narctic monkeys\nlime cordiale\n",
			},
			wantOut: "",
			wantLineCounts: LineCounts{1,0},
			wantErr: fmt.Errorf("execution stopped with 1 lines read, 0 lines written, due to error from middle function Failed here"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, lc, err := pipeStub(tt.args.in, convertLineManipulator(tt.args.middle))
			if out != tt.wantOut {
				t.Errorf("convertLineManipulator() = '%v', want '%v'", out, tt.wantOut)
			}
			if lc != tt.wantLineCounts {
				t.Errorf("convertLineManipulator() = '%v', want '%v'", lc, tt.wantLineCounts)
			}
			if (err == nil && tt.wantErr != nil) && err.Error() != tt.wantErr.Error() {
				t.Errorf("convertLineManipulator() = '%v', want '%v'", err, tt.wantErr)
			}
		})
	}
}
