package cli

import (
	"io"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// Path is one filesystem path operand a filter reads its input from.
type Path string

// Stdin builds a [Source] that reads the given reader (a wrapper passes
// inv.Stdin). It lets a filter draw its input from standard input without
// importing the framework.
func Stdin(r io.Reader) Source {
	return gloo.ByteReaderSource([]io.Reader{r})
}

// Files builds a [Source] that reads the named files in order, through fs.
func Files(fs afero.Fs, paths ...Path) Source {
	files := make([]gloo.File, len(paths))
	for i, p := range paths {
		files[i] = gloo.File(p)
	}
	return gloo.ByteFileSource(fs, files)
}

// OperandsOrStdin is the canonical Unix-filter source: the file operands when
// any are given, otherwise standard input. It centralizes the per-wrapper
// source helper that every file-or-stdin filter (cat, head, wc, …) repeats.
func OperandsOrStdin(inv Invocation) Source {
	if inv.Args.NArg() == 0 {
		return Stdin(inv.Stdin)
	}
	operands := inv.Args.Args().Slice()
	paths := make([]Path, len(operands))
	for i, o := range operands {
		paths[i] = Path(o)
	}
	return Files(inv.Fs, paths...)
}
