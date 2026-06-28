# winseq

A Windows clone of the GNU coreutils [`seq`](https://www.gnu.org/software/coreutils/manual/html_node/seq-invocation.html) command — print a sequence of numbers from the command line.

## Install

**Download a pre-built binary** from the [latest release](https://github.com/fermat-tech/winseq/releases/latest), drop `winseq.exe` anywhere on your `PATH`, and you're done — no Go toolchain required.

**Or install with Go:**

```powershell
go install github.com/fermat-tech/winseq@latest
```

**Or build from source:**

```powershell
git clone https://github.com/fermat-tech/winseq
cd winseq
go build -o winseq.exe .
```

## Usage

```
winseq [OPTION]... LAST
winseq [OPTION]... FIRST LAST
winseq [OPTION]... FIRST INCREMENT LAST
```

Prints numbers from `FIRST` to `LAST` (inclusive), stepping by `INCREMENT`. All three values may be integers or floating-point numbers. `INCREMENT` may be negative for a descending sequence. Defaults: `FIRST=1`, `INCREMENT=1`.

### Options

| Flag | Description |
|------|-------------|
| `-f FORMAT` | printf-style floating-point format (e.g. `%.2f`, `%e`, `%g`) |
| `-s STRING` | output separator — default is newline; escape sequences `\n` `\t` `\r` `\\` are interpreted |
| `-w` | equalize width by padding with leading zeroes |
| `--help` | display usage and exit |
| `--version` | output version information and exit |

## Examples

**Count to 5**
```
$ winseq 5
1
2
3
4
5
```

**Range**
```
$ winseq 3 7
3
4
5
6
7
```

**Custom step**
```
$ winseq 0 10 100
0
10
20
30
40
50
60
70
80
90
100
```

**Float sequence**
```
$ winseq 1 0.5 3
1.0
1.5
2.0
2.5
3.0
```

**Countdown**
```
$ winseq 5 -1 1
5
4
3
2
1
```

**Zero-pad to equal width (`-w`)**
```
$ winseq -w 8 12
08
09
10
11
12
```

**Comma-separated on one line (`-s`)**
```
$ winseq -s , 1 5
1,2,3,4,5
```

**Fixed decimal places (`-f`)**
```
$ winseq -f %.2f 1 0.3 2
1.00
1.30
1.60
1.90
```

**Scientific notation**
```
$ winseq -f %e 1 3
1.000000e+00
2.000000e+00
3.000000e+00
```

### Practical one-liners

Pad a batch of files with a numeric suffix:
```powershell
winseq -w 1 99 | ForEach-Object { Copy-Item template.txt "file_$_.txt" }
```

Generate an HTTP range header list:
```powershell
winseq -s "`n" 8080 8090
```

Loop over a float range in a shell pipeline:
```powershell
winseq 0.0 0.1 1.0 | ForEach-Object { Write-Host "value: $_" }
```

## Comparison with GNU seq

`winseq` is intentionally compatible with GNU `seq` for the flags it supports (`-w`, `-s`, `-f`). Differences:

| Feature | GNU `seq` | `winseq` |
|---------|-----------|----------|
| Platform | Linux / macOS | Windows (any platform) |
| `-w` zero-pad | ✓ | ✓ |
| `-s` separator | ✓ | ✓ |
| `-f` format | ✓ | ✓ |
| Float drift prevention | implementation-defined | `math/big.Rat` exact rational stepping |
| `--version` | ✓ | ✓ |
| `--help` | ✓ | ✓ |

## Notes

- Stepping uses `math/big.Rat` rational arithmetic so long float sequences (e.g. `winseq 0 0.1 1000`) do not accumulate rounding drift.
- Renaming the binary changes its name in all usage/error messages (the binary reads `os.Args[0]` at startup).
- A trailing newline is always written after the last value, matching GNU `seq` behaviour.

## Contributing

This project uses a two-branch workflow to keep `go install @latest` always resolving to a clean tagged version:

- **`dev`** — all development happens here; commit freely
- **`main`** — only ever updated by merging `dev` at release time, immediately followed by a tag

**Release checklist:**
```powershell
# 1. Finish work on dev, then merge to main
git checkout main
git merge dev --ff-only

# 2. Tag and push (tag before push so main is never untagged)
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z

# 3. Build and release
go build -ldflags "-X main.version=vX.Y.Z" -o winseq.exe .
gh release create vX.Y.Z winseq.exe --title "vX.Y.Z" --notes "..."

# 4. Switch back to dev for the next cycle
git checkout dev
```

`main` must always point to a tagged commit — any untagged commit on `main` causes `go install @latest` to report a pseudo-version (e.g. `v1.0.4-0.20260628031420-83b063bd9071`).

## License

MIT — see [LICENSE](LICENSE).
