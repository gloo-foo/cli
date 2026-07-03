// Package cli turns a gloo-foo/cmd-* pipeline command into a Unix-style
// executable.
//
// A wrapper binary declares a [Spec] — its name, flags, and a [Build] that maps
// a parsed invocation to a pipeline — and calls [Main]. This package owns
// everything that touches the framework: it constructs the urfave/cli command,
// overrides the default --version flag, wires standard input/output, executes
// the pipeline, and reports the process exit code. A wrapper therefore depends
// on this package and its cmd-* library only — never on the framework directly.
//
// The framework's byte pipeline type and source constructors are re-exported
// ([Command], [Source], [Stdin], [Files], [OperandsOrStdin]) so a wrapper builds
// its pipeline without importing the framework at all.
package cli
