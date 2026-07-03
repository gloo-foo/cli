package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// ExitCode is the process status [run] reports: 0 on success, 1 on any error.
type ExitCode int

const (
	exitOK  ExitCode = 0
	exitErr ExitCode = 1
)

// newApp builds the urfave/cli command for spec. It replaces urfave/cli's
// default --version/-v flag with a --version-only flag, freeing single-letter
// -v for a command's own flags (e.g. grep -v) while still exposing the build
// version, and keeps exit handling in [run] rather than urfave/cli's os.Exit so
// the exit code stays testable.
func newApp(spec Spec, version Version, stdin io.Reader, stdout io.Writer, fs afero.Fs) *urf.Command {
	urf.VersionFlag = &urf.BoolFlag{Name: "version", Usage: "print version information and exit"}
	return &urf.Command{
		Name:            string(spec.Name),
		Version:         version,
		Usage:           string(spec.Summary),
		UsageText:       string(spec.Synopsis),
		HideHelpCommand: true,
		ExitErrHandler:  func(context.Context, *urf.Command, error) {},
		Flags:           spec.Flags,
		Action:          action(spec, stdin, stdout, fs),
	}
}

// action adapts spec.Build into a urfave/cli action: it builds the pipeline for
// the parsed invocation and runs it into stdout.
func action(spec Spec, stdin io.Reader, stdout io.Writer, fs afero.Fs) urf.ActionFunc {
	return func(_ context.Context, c *urf.Command) error {
		return execute(spec, Invocation{Args: c, Stdin: stdin, Fs: fs}, stdout)
	}
}

// execute builds one invocation's pipeline and runs it to stdout. A nil filter
// means the source is the whole pipeline (a source command); otherwise the
// source feeds the filter.
func execute(spec Spec, inv Invocation, stdout io.Writer) error {
	source, filter, err := spec.Build(inv)
	if err != nil {
		return err
	}
	sink := gloo.ByteWriteTo(stdout)
	if filter == nil {
		_, err = gloo.Run(source, sink)
		return err
	}
	_, err = gloo.Run(source, sink, filter)
	return err
}

// run executes spec's CLI against the injected version, arguments, and I/O,
// returning the process exit code. On error it prints "name: message" to stderr
// and reports [exitErr].
func run(spec Spec, version Version, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) ExitCode {
	app := newApp(spec, version, stdin, stdout, fs)
	app.Writer = stdout
	app.ErrWriter = stderr
	if err := app.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, "%s: %v\n", spec.Name, err)
		return exitErr
	}
	return exitOK
}

// osExit and runner are indirection seams: they let [Main] be exercised without
// terminating the test process or spawning the CLI for real.
var (
	osExit = os.Exit
	runner = run
)

// Main is a wrapper's entry point. It runs spec against the real process
// environment and exits with the resulting code, so a wrapper's main is:
//
//	func main() { cli.Main(spec, version) }
func Main(spec Spec, version Version) {
	code := runner(spec, version, os.Args, os.Stdin, os.Stdout, os.Stderr, afero.NewOsFs())
	osExit(int(code))
}
