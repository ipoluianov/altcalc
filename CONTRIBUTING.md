# Contributing

Bug reports and ideas are welcome in
[issues](https://github.com/ipoluianov/altcalc/issues). Please include the
AltCalc version (shown in About) and your operating system.

## Building

You need Go, see the version in [go.mod](go.mod). No cgo is required, so
any platform builds for any other.

```sh
go build .
go test ./...
```

The UI library is [nui](https://github.com/ipoluianov/nui). To work on both
at once, uncomment the `replace` line at the end of `go.mod`.

## Code layout

- `calc/` - the expressions: the tokens, the parser, the numbers (exact
  integers, floats), the functions and the formatting; it has no UI and is
  covered by tests
- `forms/` - the window: the display, the keypad, the history, the dialogs;
  `forms/session.go` is what the keys do, without the UI, and is tested;
  `forms/strings.go` holds all the texts
- `config/` - the settings and the history
- `install/` - installing a downloaded copy on Windows
- `instance/` - keeps a single running copy
- `build/` - release builds and packages, `scripts/` - the Linux installer

## Pull requests

- Keep a change to one thing, and describe what it does and why.
- Every text goes into all the languages in `forms/strings.go`.
  `TestAllTranslated` fails on a missing one. A machine translation is
  fine, a native speaker can fix it later.
- Say on which systems you tried it.
