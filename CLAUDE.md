# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`clibase` (module `github.com/shouni/clibase`) is a single-file Go library that provides a common
`spf13/cobra`-based foundation for building CLI applications. Consumers import the package and call
`clibase.Execute(clibase.App{...})` instead of hand-rolling their own root command boilerplate.

The entire implementation lives in `root.go`, with tests in `root_test.go`. There is no `main`
package or subcommands in this repo — it is purely a library other projects vendor via `go get`.

## Commands

```bash
go build ./...   # compile check
go vet ./...     # static analysis
go test ./...    # run the root_test.go suite
```

There is no Makefile, linter config, or CI workflow in this repo — the above are the standard Go
toolchain commands.

## Architecture

- `App` struct is the public configuration surface: `Name`, `Version`, `SilenceUsage`,
  `SilenceErrors`, `AddFlags`, `PreRunE`, `PostRun`, `Commands`. Consumers populate this struct and
  pass it to `Execute` or `ExecuteContext`.
- `buildRootCmd(app App)` (unexported) constructs the cobra `rootCmd`: wires the `App` hooks into
  `PersistentPreRunE`/`PersistentPostRun`/`RunE`, sets `Version`/`SilenceUsage`/`SilenceErrors`,
  registers the two standard persistent flags (`--verbose`/`-V`, `--config`/`-C`), runs
  `app.AddFlags`, and attaches `app.Commands`. Both `Execute` and `ExecuteContext` go through this
  one builder — keep new `App` fields wired here rather than duplicating logic in the two entry
  points.
- `buildRootCmd` also sets `cobra.EnableTraverseRunHooks = true`. This is a global cobra switch, not
  a per-command setting: without it, cobra only runs the *closest* `PersistentPreRunE`/
  `PersistentPostRun` in the command tree, so a subcommand defining its own persistent hook would
  silently skip `app.PreRunE`/`app.PostRun`. Do not remove this — it's load-bearing for the
  documented lifecycle contract, and there's a regression test for it
  (`TestExecuteContext_RootAndChildHooksBothFire`).
- `ExecuteContext(ctx, app) error` builds the command and calls `rootCmd.ExecuteContext(ctx)`,
  returning the error instead of exiting — use this when a caller wants custom exit-code handling or
  its own cancellation context.
- `Execute(app)` is the original one-shot entry point: it derives a `context.Context` cancelled on
  SIGINT/SIGTERM via `signal.NotifyContext`, delegates to `ExecuteContext`, and calls `os.Exit(1)` on
  error (cobra already prints the error, so this file does not print it again). Command bodies can
  read `cmd.Context()` to react to the interrupt.
- `Config` / `globalConfig` / `GetConfig()` hold the values bound to the two standard flags.
  `globalConfig` is a package-level var mutated by `pflag` binding when `buildRootCmd` registers the
  flags; `GetConfig()` returns a copy so callers can't mutate package state. This is intentionally
  not mutex-guarded: writes only happen synchronously during flag parsing, before any `Run`/`RunE`
  body executes, so there's no real concurrent-write window as long as `Execute`/`ExecuteContext` is
  called once per process (the normal CLI case). Any new common flag should follow this same
  pattern: add a field to `Config`, bind it in `buildRootCmd`'s flag registration block.
- Lifecycle order for a consuming app: cobra parses flags → `app.PreRunE` (app-specific validation/
  init) → the invoked subcommand's own `PersistentPreRunE` if any (both fire, root then child, due to
  `EnableTraverseRunHooks`) → the subcommand's `Run`/`RunE` → child's `PersistentPostRun` if any →
  `app.PostRun` (cleanup, e.g. closing resources). If no subcommand/args are given, `rootCmd.RunE`
  just prints help.

## Conventions

- Doc comments in `root.go` are written in Japanese; match that style if adding comments to this file
  or `root_test.go`.
- Keep the library dependency-light — only `spf13/cobra` (and its transitive `pflag`/`mousetrap`) are
  in `go.mod`. Avoid pulling in new dependencies for what should be a minimal base package.
- The standard flags use shorthand `-V`/`-C`. If a consumer's `AddFlags` registers a flag with the
  same shorthand, cobra panics at startup — this is a known constraint, not a bug to fix reactively.
