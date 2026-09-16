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

// passthrough is a filter Command that emits its input unchanged, so a test can
// assert what flowed from the built Source without a real command's transform.
func passthrough() Command {
	return gloo.FuncCommand[[]byte, []byte](
		func(_ context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] { return in },
	)
}

// sentinelErr is a Build failure used to drive the error exit path.
const sentinelErr gloo.Error = "boom"

// exec runs spec with the given argv/stdin and returns the exit code plus the
// captured stdout and stderr, keeping each test to its assertion.
func exec(spec Spec, version Version, args []string, stdin io.Reader) (ExitCode, string, string) {
	var out, errOut bytes.Buffer
	code := run(spec, version, args, stdin, &out, &errOut, afero.NewMemMapFs())
	return code, out.String(), errOut.String()
}

// stdinFilter is a wrapper whose pipeline is Stdin → passthrough — the minimal
// filter shape for exercising run's success path and execute's filter branch.
func stdinFilter() Spec {
	return Spec{
		Name:     "flt",
		Summary:  "passthrough filter",
		Synopsis: "flt",
		Build: func(inv Invocation) (Source, Command, error) {
			return Stdin(inv.Stdin), passthrough(), nil
		},
	}
}

func TestRun_FilterReadsStdin(t *testing.T) {
	code, out, errOut := exec(stdinFilter(), "1.0", []string{"flt"}, strings.NewReader("alpha\nbeta\n"))
	if code != exitOK {
		t.Fatalf("code=%d, want %d (stderr=%q)", code, exitOK, errOut)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("stdout=%q, want stdin passed through", out)
	}
}

func TestRun_SourceCommandNilFilter(t *testing.T) {
	// A nil filter means the Source is the whole pipeline (a source command).
	spec := Spec{
		Name: "src",
		Build: func(inv Invocation) (Source, Command, error) {
			return Stdin(inv.Stdin), nil, nil
		},
	}
	code, out, errOut := exec(spec, "1.0", []string{"src"}, strings.NewReader("solo\n"))
	if code != exitOK {
		t.Fatalf("code=%d, want %d (stderr=%q)", code, exitOK, errOut)
	}
	if !strings.Contains(out, "solo") {
		t.Fatalf("stdout=%q, want source emitted", out)
	}
}

func TestRun_BuildErrorExitsOneToStderr(t *testing.T) {
	spec := Spec{
		Name:  "bad",
		Build: func(Invocation) (Source, Command, error) { return nil, nil, sentinelErr },
	}
	code, _, errOut := exec(spec, "1.0", []string{"bad"}, strings.NewReader(""))
	if code != exitErr {
		t.Fatalf("code=%d, want %d", code, exitErr)
	}
	if !strings.Contains(errOut, "bad: boom") {
		t.Fatalf("stderr=%q, want %q", errOut, "bad: boom")
	}
}

func TestRun_VersionFlag(t *testing.T) {
	code, out, errOut := exec(stdinFilter(), "9.9.9", []string{"flt", "--version"}, strings.NewReader(""))
	if code != exitOK {
		t.Fatalf("code=%d, want %d (stderr=%q)", code, exitOK, errOut)
	}
	if !strings.Contains(out, "9.9.9") {
		t.Fatalf("stdout=%q, want version", out)
	}
}

func TestRun_FlagsAreWired(t *testing.T) {
	// A declared flag is parseable and readable in Build.
	var seen bool
	spec := Spec{
		Name:  "fl",
		Flags: []urf.Flag{&urf.BoolFlag{Name: "on", Aliases: []string{"o"}}},
		Build: func(inv Invocation) (Source, Command, error) {
			seen = inv.Args.Bool("on")
			return Stdin(inv.Stdin), passthrough(), nil
		},
	}
	code, _, errOut := exec(spec, "1.0", []string{"fl", "-o"}, strings.NewReader("x\n"))
	if code != exitOK || !seen {
		t.Fatalf("code=%d seen=%v, want 0/true (stderr=%q)", code, seen, errOut)
	}
}

func TestMain_RunsThroughSeams(t *testing.T) {
	origExit, origRunner := osExit, runner
	t.Cleanup(func() { osExit, runner = origExit, origRunner })

	var gotSpec Name
	var gotVersion Version
	runner = func(spec Spec, version Version, _ []string, _ io.Reader, _, _ io.Writer, _ afero.Fs) ExitCode {
		gotSpec, gotVersion = spec.Name, version
		return 3
	}
	var gotCode int
	osExit = func(c int) { gotCode = c }

	Main(Spec{Name: "z"}, "7.7")

	if gotSpec != "z" || gotVersion != "7.7" || gotCode != 3 {
		t.Fatalf("spec=%q version=%q code=%d, want z/7.7/3", gotSpec, gotVersion, gotCode)
	}
}
