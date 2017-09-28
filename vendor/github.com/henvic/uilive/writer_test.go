package uilive

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWriter(t *testing.T) {
	w := New()
	b := &bytes.Buffer{}
	w.Out = b
	for i := 0; i <= 3; i++ {
		fmt.Fprintf(w, "foo %d\n", i)
		w.Flush()
	}
	fmt.Fprintln(b, "bar")

	want := "foo 0\n\x1b[1A\x1b[2Kfoo 1\n\x1b[1A\x1b[2Kfoo 2\n\x1b[1A\x1b[2Kfoo 3\nbar\n"
	if b.String() != want {
		t.Fatalf("want %q, got %q", want, b.String())
	}
}
