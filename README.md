# AltCalc

A calculator that you type into: the result is shown as you type, with
the functions, the variables and the history at hand. It runs on Linux,
Windows and macOS as a single binary, without cgo.

Website and documentation: https://altbins.pro/altcalc/

## Features

- **Type the expression.** `2 + 3 × sin(30)` shows `= 3.5` while it is typed;
  Enter puts it into the history. An operator typed next goes on with the
  result, a number starts a new calculation. The missing parentheses at the
  end are added. The colors tell the numbers, the operators, the functions
  and the names apart, an unknown name or an error is underlined.
- **Exact where it can be.** The integers are exact up to ±9.2·10¹⁸
  (`20!`, `2^62`), `0.1 + 0.2` is `0.3`, `sin 180°` is `0`, and a result
  put back into an expression keeps its full precision: `1/3`, Enter, `×3`
  gives `1`.
- **Percent as on a calculator.** `200 + 15%` is 230, `200 × 15%` is 30.
- **Functions.** sin, cos, tan and the inverse and hyperbolic ones; sqrt,
  cbrt, root, exp, ln, log, log2; abs, round, floor, ceil, frac; min, max,
  sum, avg, median; gcd, lcm, n!, nCr, nPr; rand. Degrees, radians or
  gradians, and `90°` in any of them. A function of one argument needs no
  parentheses: `sin 30`, `ln 2`.
- **Variables.** `rate = 0.2`, then `1500 × rate`; `ans` is the last result.
  The variables are kept between the starts.
- **For programmers.** `0xFF`, `0b1010`, `0o17`; `&`, `|`, `xor`, `~`, `<<`,
  `>>`. An integer result is also shown in hex, octal and binary.
- **History.** The calculations of the earlier starts too. A click puts the
  result into the expression, a double click the expression itself; Up and
  Down go through them from the keyboard.
- **Memory.** MC, MR, M+, M− as on a desk calculator (Ctrl+L, R, P, Q, M).
- **Keypad.** Standard, or scientific with the functions (Ctrl+1, Ctrl+2);
  "2nd" turns the keys to the inverse functions.
- **Clipboard.** Ctrl+C copies the result, Ctrl+Shift+C the calculation.
  A number pasted as `1,234.56` or `1.234,56` is read as it is meant.
- **Interface.** Light and dark themes, English and Russian, the decimal
  point or comma, the thousands separated or not; always on top; one copy
  at a time.

Help → Quick Reference (F1) lists everything that can be typed.

## Installation

Download a build for your system from the
[releases](https://github.com/ipoluianov/altcalc/releases):

- **Windows:** `altcalc.exe`.
- **macOS** (Apple Silicon): the `.dmg` image.
- **Linux** (x86-64 and ARM64): a `.deb` or `.rpm` package, or a user-level
  install (no root) that picks the build for your machine:

  ```sh
  curl -fsSL https://github.com/ipoluianov/altcalc/releases/latest/download/linux-install.sh | bash
  ```

Settings and the history are stored in `~/.altbins/.altcalc/`. Removing
the application leaves them in place.

### Windows

The exe is not signed, so on the first start SmartScreen may say "Windows
protected your PC". Click **More info**, then **Run anyway**.

A downloaded copy offers to install itself into `%USERPROFILE%\.altbins`
with the **Install** link in the status bar. It adds AltCalc to the Start
menu, the desktop and "Installed apps". When a newer version is started
from a download, the link reads **Update**. Remove AltCalc from "Installed
apps" or with the **Uninstall** link.

### macOS

Open the `.dmg` and drag AltCalc to Applications.

### Linux

AltCalc needs X11, and runs under Wayland through XWayland, which GNOME
and KDE have. glibc-based distributions only (not Alpine).

## Building

You need Go (see the version in [go.mod](go.mod)). No cgo is required.

```sh
go build .
go test ./...
```

`build/all.sh` (or `build/all.bat`) builds every platform into `bin/`,
together with the deb/rpm packages, the dmg and the installer script.

## License

[MIT](LICENSE) © Ivan Poluianov
