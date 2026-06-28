// winseq prints a sequence of numbers.
//
// Usage:
//
//	winseq [OPTION]... LAST
//	winseq [OPTION]... FIRST LAST
//	winseq [OPTION]... FIRST INCREMENT LAST
//
// winseq is a Windows clone of the GNU coreutils seq(1) command. It prints
// numbers from FIRST to LAST (inclusive) stepping by INCREMENT. All three
// values may be integers or floating-point numbers; INCREMENT may be negative
// for a descending sequence.
//
// Options:
//
//	-f FORMAT   printf-style floating-point format (e.g. %.2f, %e)
//	-s STRING   output separator; default is a newline
//	            escape sequences \n \t \r \\ are interpreted
//	-w          equalize width by padding with leading zeroes
//	--help      display usage and exit
//	--version   output version information and exit
//
// Rational arithmetic (math/big.Rat) is used for stepping so that long
// float sequences do not accumulate drift.
package main

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var progName string

func init() {
	progName = strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... LAST\n", progName)
	fmt.Fprintf(os.Stderr, "       %s [OPTION]... FIRST LAST\n", progName)
	fmt.Fprintf(os.Stderr, "       %s [OPTION]... FIRST INCREMENT LAST\n", progName)
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -f FORMAT   use printf-style floating-point FORMAT\n")
	fmt.Fprintf(os.Stderr, "  -s STRING   use STRING as output separator (default: \\n)\n")
	fmt.Fprintf(os.Stderr, "  -w          equalize width by padding with leading zeroes\n")
	fmt.Fprintf(os.Stderr, "  --help      display this help and exit\n")
	fmt.Fprintf(os.Stderr, "  --version   output version information and exit\n")
}

// decimalPlaces returns the number of decimal places in a numeric string.
func decimalPlaces(s string) int {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return len(s) - i - 1
	}
	return 0
}

// formatDefault builds a %g-style or %f-style format that shows exactly `prec`
// decimal places (0 means integer format).
func defaultFormat(prec int) string {
	if prec == 0 {
		return "%.0f"
	}
	return fmt.Sprintf("%%.%df", prec)
}

// widthOf returns the printed width of v using fmt.
func widthOf(v float64, fmt_ string) int {
	return len(fmt.Sprintf(fmt_, v))
}

// paddedFormat wraps a format string so the numeric field is zero-padded to
// at least `width` digits (before the decimal point).
func paddedFormat(baseFmt string, width int) string {
	// Insert 0<width> between % and the rest of the format verb.
	// baseFmt is like "%.2f" or "%.0f".
	// We want "%0<width>.2f".
	after := strings.TrimPrefix(baseFmt, "%")
	return fmt.Sprintf("%%0%d%s", width, after)
}

// translateFormat converts a printf-style format with %g/%e/%f to Go's fmt.
// Supports: %[flags][width][.prec](e|E|f|g|G) — passes through unknown specs.
func translateFormat(f string) string {
	// Go's fmt package handles %e %E %f %g %G natively — no translation needed.
	return f
}

func die(msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", progName, msg)
	os.Exit(1)
}

func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		die(fmt.Sprintf("invalid floating point argument: %q", s))
	}
	return v
}

func main() {
	var (
		separator = "\n"
		fmtFlag   = ""
		padZero   = false
	)

	args := os.Args[1:]

	// Parse options (stop at first non-option or "--").
	i := 0
	for i < len(args) {
		a := args[i]
		if a == "--" {
			i++
			break
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			break
		}
		switch {
		case a == "--help":
			usage()
			os.Exit(0)
		case a == "--version":
			fmt.Printf("%s 1.0.0\n", progName)
			os.Exit(0)
		case a == "-w":
			padZero = true
			i++
		case a == "-s":
			if i+1 >= len(args) {
				die("option requires an argument -- 's'")
			}
			separator = args[i+1]
			// Interpret common escape sequences.
			separator = strings.NewReplacer(
				`\n`, "\n", `\t`, "\t", `\r`, "\r", `\\`, "\\",
			).Replace(separator)
			i += 2
		case strings.HasPrefix(a, "-s"):
			separator = strings.NewReplacer(
				`\n`, "\n", `\t`, "\t", `\r`, "\r", `\\`, "\\",
			).Replace(a[2:])
			i++
		case a == "-f":
			if i+1 >= len(args) {
				die("option requires an argument -- 'f'")
			}
			fmtFlag = translateFormat(args[i+1])
			i += 2
		case strings.HasPrefix(a, "-f"):
			fmtFlag = translateFormat(a[2:])
			i++
		default:
			die(fmt.Sprintf("invalid option -- '%s'", a))
		}
	}

	rest := args[i:]
	if len(rest) == 0 || len(rest) > 3 {
		usage()
		os.Exit(1)
	}

	var firstStr, incrStr, lastStr string
	switch len(rest) {
	case 1:
		firstStr, incrStr, lastStr = "1", "1", rest[0]
	case 2:
		firstStr, incrStr, lastStr = rest[0], "1", rest[1]
	case 3:
		firstStr, incrStr, lastStr = rest[0], rest[1], rest[2]
	}

	first := parseFloat(firstStr)
	incr := parseFloat(incrStr)
	last := parseFloat(lastStr)

	if incr == 0 {
		die("invalid Zero increment value: '0'")
	}

	// Determine decimal precision from the input strings.
	prec := max(decimalPlaces(firstStr), decimalPlaces(incrStr), decimalPlaces(lastStr))

	// Choose numeric format.
	numFmt := fmtFlag
	if numFmt == "" {
		numFmt = defaultFormat(prec)
	}

	// Collect values using rational arithmetic to avoid float drift.
	// We use big.Rat for exact stepping.
	ratFirst := new(big.Rat).SetFloat64(first)
	ratIncr := new(big.Rat).SetFloat64(incr)
	ratLast := new(big.Rat).SetFloat64(last)

	var values []float64
	cur := new(big.Rat).Set(ratFirst)
	for {
		f, _ := cur.Float64()
		// Stop condition: if incr > 0, stop when cur > last; if incr < 0, stop when cur < last.
		cmp := cur.Cmp(ratLast)
		if incr > 0 && cmp > 0 {
			break
		}
		if incr < 0 && cmp < 0 {
			break
		}
		values = append(values, f)
		cur.Add(cur, ratIncr)
	}

	if len(values) == 0 {
		os.Exit(0)
	}

	// If -w, determine the widest formatted value and build a zero-padded format.
	if padZero && fmtFlag == "" {
		maxW := 0
		for _, v := range values {
			w := widthOf(v, numFmt)
			if w > maxW {
				maxW = w
			}
		}
		// The width to pad to is the integer-part width of the widest value.
		// For integers, widthOf gives the full width; for floats we want the
		// total width including decimal point and fraction.
		numFmt = paddedFormat(numFmt, maxW)
	}

	w := os.Stdout
	for idx, v := range values {
		if idx > 0 {
			fmt.Fprint(w, separator)
		}
		fmt.Fprintf(w, numFmt, v)
	}
	// GNU seq always ends with a newline (even with custom separator).
	fmt.Fprintln(w)
}

func max(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

