# gloo cli

[![CI](https://github.com/gloo-foo/cli/actions/workflows/go.yml/badge.svg)](https://github.com/gloo-foo/cli/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gloo-foo/cli.svg)](https://pkg.go.dev/github.com/gloo-foo/cli)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Shared CLI runner — expose a `gloo-foo/cmd-*` command as a Unix-style executable without touching the framework.

A wrapper declares a `Spec` (name, flags, and a `Build` mapping parsed arguments to a pipeline) and calls `Main`. This package owns the urfave/cli wiring, the pipeline execution, and the process exit code, and re-exports the framework's pipeline type and source constructors — so a wrapper depends only on this package and its `cmd-*` library.
