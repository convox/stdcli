package stdcli

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type lineReader struct{ lines []string }

func (r *lineReader) Read(p []byte) (int, error) {
	if len(r.lines) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.lines[0])
	r.lines = r.lines[1:]
	return n, nil
}

type failWriter struct{ err error }

func (w failWriter) Write(p []byte) (int, error) {
	return 1, w.err
}

func TestWriterWriteReturnsInputLength(t *testing.T) {
	cases := []struct {
		color   bool
		in, out string
	}{
		{false, "\x1b[36mINFO\x1b[0m pull\n", "INFO pull\n"},
		{true, "\x1b[36mINFO\x1b[0m pull\n", "\x1b[36mINFO\x1b[0m pull\n"},
		{false, "<h1>title</h1>\n", "title\n"},
		{true, "<h1>title</h1>\n", "\x1b[38;5;244mtitle\x1b[0m\n"},
		{false, "plain\n", "plain\n"},
	}

	for _, tc := range cases {
		var out bytes.Buffer
		w := &Writer{Color: tc.color, Stdout: &out, Tags: DefaultWriter.Tags}

		n, err := w.Write([]byte(tc.in))
		if err != nil || n != len(tc.in) || out.String() != tc.out {
			t.Errorf("Write(%q) color=%t: n=%d err=%v out=%q", tc.in, tc.color, n, err, out.String())
		}
	}
}

func TestWriterWriteError(t *testing.T) {
	errFail := errors.New("write failed")
	w := &Writer{Stdout: failWriter{err: errFail}, Tags: DefaultWriter.Tags}

	n, err := w.Write([]byte("\x1b[36mINFO\x1b[0m pull\n"))
	if n != 0 || err != errFail {
		t.Errorf("Write: n=%d err=%v, want 0 and %v", n, err, errFail)
	}
}

func TestWriterCopyColoredStream(t *testing.T) {
	lines := []string{"plain1\n", "\x1b[36mINFO\x1b[0m pull\n", "<h1>title</h1>\n", "plain2\n"}

	cases := []struct {
		color bool
		out   string
	}{
		{false, "plain1\nINFO pull\ntitle\nplain2\n"},
		{true, "plain1\n\x1b[36mINFO\x1b[0m pull\n\x1b[38;5;244mtitle\x1b[0m\nplain2\n"},
	}

	size := 0
	for _, l := range lines {
		size += len(l)
	}

	for _, tc := range cases {
		var out bytes.Buffer
		w := &Writer{Color: tc.color, Stdout: &out, Tags: DefaultWriter.Tags}

		n, err := io.Copy(w, &lineReader{lines: append([]string{}, lines...)})
		if err != nil || n != int64(size) || out.String() != tc.out {
			t.Errorf("Copy color=%t: n=%d err=%v out=%q", tc.color, n, err, out.String())
		}
	}
}
