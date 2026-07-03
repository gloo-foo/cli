package cli

import (
	"io"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// Command is a byte-to-byte pipeline stage, re-exported from the framework so a
// wrapper names it through this package rather than importing the framework.
type Command = gloo.Command[[]byte, []byte]

// Source is a byte pipeline source, re-exported for the same reason.
type Source = gloo.Source[[]byte]

// File is a filesystem path a source command reads from (e.g. the root of ls or
// find), re-exported so a wrapper converts a path operand without importing the
// framework.
type File = gloo.File

// Version is the build version string a wrapper threads in from its main via
// ldflags. It is a plain alias because urfave/cli's Version field is a string
// and must stay bound to that ldflags symbol without conversion.
type Version = string

// Name is the executable's invoked name (e.g. "cat").
type Name string

// Summary is the one-line description shown in the --help header.
type Summary string

// Synopsis is the multi-line usage block shown in --help. urfave/cli indents the
// whole block three spaces, so a wrapper keeps its lines flush-left.
type Synopsis string

// Invocation is the parsed command line handed to a [Build]: the urfave/cli
// accessor for flags and operands, plus the injected standard input and
// filesystem the source is drawn from.
type Invocation struct {
	// Args accesses parsed flags (Bool/String/…) and operands (Args/NArg).
	Args *urf.Command
	// Stdin is the process's standard input, for filters that read it.
	Stdin io.Reader
	// Fs is the filesystem file operands are read through (injectable for tests).
	Fs afero.Fs
}

// Build maps one parsed [Invocation] to the pipeline to execute: the [Source]
// that feeds it and an optional filter [Command].
//
//   - A non-nil filter runs source → filter → stdout (an ordinary Unix filter).
//   - A nil filter runs source → stdout, so the Source is the whole pipeline (a
//     source command such as ls, seq, or yes).
//   - A non-nil error aborts the run with that error (e.g. a missing operand).
type Build func(inv Invocation) (Source, Command, error)

// Spec declares one wrapper executable: its identity, its flags, and how it
// builds its pipeline. It is an immutable value passed by a wrapper to [Main].
type Spec struct {
	Name     Name
	Summary  Summary
	Synopsis Synopsis
	Build    Build
	Flags    []urf.Flag
}
