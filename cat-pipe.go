package cat_pipe

import (
	"bufio"
	"io"
)

// Pipe reads lines from r (up to '\n), send to middle function, and write the returned string to w.
// If middle were a bash function that took in and returned a string, this should work as `cat r | middle > w`
func Pipe(r io.Reader, w io.Writer, middle LineManipulator) (result LineCounts, err error) {
	return pipe(r, w, convertLineManipulator(middle))
}

// PipeWithBytes is the same basic function as Pipe, but allows the function to work with byte arrays directly.
func PipeWithBytes(r io.Reader, w io.Writer, middle RawByteManipulator) (LineCounts, error) {
	return pipe(r, w, convertRawByteManipulator(middle))
}

type middleFunction = func(sc *bufio.Scanner, wr *bufio.Writer, lc *LineCounts) (err error)

func pipe(r io.Reader, w io.Writer, middle middleFunction) (LineCounts, error) {
	lineCounts := NewLineCounts(0, 0)

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	writer := bufio.NewWriter(w)

	for scanner.Scan() {
		lineCounts.ReadLineCount += 1

		err := middle(scanner, writer, &lineCounts)
		if err != nil {
			return lineCounts, err
		}
	}

	err := scanner.Err()
	if err != nil {
		return lineCounts, NewReadError(lineCounts, err)
	}
	return lineCounts, nil
}

func convertRawByteManipulator(middle RawByteManipulator) middleFunction {
	return func(sc *bufio.Scanner, wr *bufio.Writer, lc *LineCounts) error {
		in := sc.Bytes()

		out, err := middle(in)
		if err != nil {
			return NewMiddleError(*lc, err)
		}

		if len(out) == 0 {
			return nil
		}

		_, err = wr.Write(append(out, byte('\n')))
		lc.WrittenLineCount += 1
		if err != nil {
			return NewWriteError(*lc, err)
		}

		return nil
	}
}

func convertLineManipulator(middle LineManipulator) middleFunction {
	return func(sc *bufio.Scanner, wr *bufio.Writer, lc *LineCounts) error {
		in := sc.Text()

		out, err := middle(in)
		if err != nil {
			return NewMiddleError(*lc, err)
		}

		if len(out) == 0 {
			return nil
		}

		_, err = wr.WriteString(out + "\n")
		lc.WrittenLineCount += 1
		if err != nil {
			return NewWriteError(*lc, err)
		}
		return nil
	}

}
