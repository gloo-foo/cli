package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// collect runs a Source to completion and returns its output as a string, so a
// test can assert what the source produced.
func collect(t *testing.T, src Source) string {
	t.Helper()
	var buf bytes.Buffer
	if _, err := gloo.Run(src, gloo.ByteWriteTo(&buf)); err != nil {
		t.Fatalf("run source: %v", err)
	}
	return buf.String()
}

// invoke parses args through a bare urfave/cli command and returns the
// Invocation its action sees, so OperandsOrStdin can be tested against real
// parsed operands without standing up the full run().
func invoke(t *testing.T, args []string, stdin io.Reader, fs afero.Fs) Invocation {
	t.Helper()
	var got Invocation
	app := &urf.Command{
		Name: "x",
		Action: func(_ context.Context, c *urf.Command) error {
			got = Invocation{Args: c, Stdin: stdin, Fs: fs}
			return nil
		},
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("invoke: %v", err)
	}
	return got
}

func TestStdin_ReadsReader(t *testing.T) {
	out := collect(t, Stdin(strings.NewReader("one\ntwo\n")))
	if !strings.Contains(out, "one") || !strings.Contains(out, "two") {
		t.Fatalf("out=%q, want reader contents", out)
	}
}

func TestFiles_ReadsNamedFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "a.txt", []byte("aaa\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	out := collect(t, Files(fs, "a.txt"))
	if !strings.Contains(out, "aaa") {
		t.Fatalf("out=%q, want file contents", out)
	}
}

func TestOperandsOrStdin_UsesStdinWhenNoOperands(t *testing.T) {
	inv := invoke(t, []string{"x"}, strings.NewReader("from-stdin\n"), afero.NewMemMapFs())
	if !strings.Contains(collect(t, OperandsOrStdin(inv)), "from-stdin") {
		t.Fatal("want stdin used when no operands")
	}
}

func TestOperandsOrStdin_UsesFilesWhenOperands(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "f.txt", []byte("from-file\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	inv := invoke(t, []string{"x", "f.txt"}, strings.NewReader("ignored\n"), fs)
	out := collect(t, OperandsOrStdin(inv))
	if !strings.Contains(out, "from-file") || strings.Contains(out, "ignored") {
		t.Fatalf("out=%q, want file operand, not stdin", out)
	}
}
