package learn

import (
	"bytes"
	"testing"
	"time"
)

// feed returns a rawReader that drains a fixed byte sequence and then blocks.
func feed(b []byte) *rawReader {
	return newRawReader(bytes.NewReader(b))
}

func TestDecode_SingleKeys(t *testing.T) {
	cases := []struct {
		in    byte
		want  string
		wantA action
	}{
		{'a', "a", actionAssign},
		{'Z', "z", actionAssign},
		{'7', "7", actionAssign},
		{' ', "space", actionAssign},
		{'\t', "tab", actionAssign},
		{'\r', "enter", actionAssign},
		{'\n', "enter", actionAssign},
		{0x03, "", actionQuit},
	}
	for _, c := range cases {
		got, act := decodeKey(c.in, feed(nil))
		if got != c.want || act != c.wantA {
			t.Errorf("decodeKey(%#x) = (%q, %d), want (%q, %d)", c.in, got, act, c.want, c.wantA)
		}
	}
}

func TestDecode_EscStandalone(t *testing.T) {
	// ESC with no follow-up bytes: must resolve to actionSkipEdge via timeout.
	got, act := decodeKey(0x1B, feed(nil))
	if act != actionSkipEdge || got != "" {
		t.Errorf("standalone ESC = (%q, %d), want (\"\", actionSkipEdge)", got, act)
	}
}

func TestDecode_Arrows(t *testing.T) {
	cases := []struct {
		seq  []byte
		want string
	}{
		{[]byte{'[', 'A'}, "up"},
		{[]byte{'[', 'B'}, "down"},
		{[]byte{'[', 'C'}, "right"},
		{[]byte{'[', 'D'}, "left"},
	}
	for _, c := range cases {
		rr := feed(c.seq[1:])
		got, act := decodeEscSeq(c.seq[0], rr)
		if got != c.want || act != actionAssign {
			t.Errorf("arrow %q = (%q, %d), want (%q, actionAssign)", c.seq, got, act, c.want)
		}
	}
}

func TestDecode_FKeys_SS3(t *testing.T) {
	cases := []struct {
		seq  []byte
		want string
	}{
		{[]byte{'O', 'P'}, "f1"},
		{[]byte{'O', 'Q'}, "f2"},
		{[]byte{'O', 'R'}, "f3"},
		{[]byte{'O', 'S'}, "f4"},
	}
	for _, c := range cases {
		rr := feed(c.seq[1:])
		got, act := decodeEscSeq(c.seq[0], rr)
		if got != c.want || act != actionAssign {
			t.Errorf("SS3 %q = (%q, %d), want (%q, actionAssign)", c.seq, got, act, c.want)
		}
	}
}

func TestDecode_FKeys_Tilde(t *testing.T) {
	cases := []struct {
		tail string
		want string
	}{
		{"15~", "f5"},
		{"17~", "f6"},
		{"18~", "f7"},
		{"19~", "f8"},
		{"20~", "f9"},
		{"21~", "f10"},
		{"23~", "f11"},
		{"24~", "f12"},
	}
	for _, c := range cases {
		rr := feed([]byte(c.tail))
		got, act := decodeEscSeq('[', rr)
		if got != c.want || act != actionAssign {
			t.Errorf("CSI %s = (%q, %d), want (%q, actionAssign)", c.tail, got, act, c.want)
		}
	}
}

func TestDecode_Unknown(t *testing.T) {
	// A random control byte with no mapping should retry.
	got, act := decodeKey(0x02, feed(nil))
	if act != actionRetry || got != "" {
		t.Errorf("byte 0x02 = (%q, %d), want (\"\", actionRetry)", got, act)
	}
}

// blockingReader never yields bytes; Read blocks until the goroutine is
// cancelled (simulates a TTY with no pending input).
type blockingReader struct{ stop chan struct{} }

func (b *blockingReader) Read(p []byte) (int, error) {
	<-b.stop
	return 0, nil
}

func TestRawReader_Timeout(t *testing.T) {
	br := &blockingReader{stop: make(chan struct{})}
	defer close(br.stop)
	rr := newRawReader(br)

	start := time.Now()
	_, ok := rr.readTimeout(20 * time.Millisecond)
	if ok {
		t.Fatalf("expected timeout, got byte")
	}
	if time.Since(start) < 15*time.Millisecond {
		t.Fatalf("readTimeout returned too early (%v)", time.Since(start))
	}
}
