package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/onurmicoogullari/wc-tool-go/internal/ccwc"
)

func parseFlags(args []string) (ccwc.Options, string, error) {
	fs := flag.NewFlagSet("ccwc", flag.ContinueOnError)
	var (
		c = fs.Bool("c", false, "prints the byte count")
		l = fs.Bool("l", false, "prints the line count")
		w = fs.Bool("w", false, "prints the word count")
		m = fs.Bool("m", false, "prints the char count")
	)
	if err := fs.Parse(args); err != nil {
		return ccwc.Options{}, "", err
	}

	if !(*c || *l || *w || *m) {
		*c, *l, *w = true, true, true
	}

	var filename string

	switch fs.NArg() {
	case 0:
	case 1:
		filename = fs.Arg(0)
	default:
		return ccwc.Options{}, "", fmt.Errorf("too many arguments")
	}

	return ccwc.Options{CountBytes: *c, CountLines: *l, CountWords: *w, CountChars: *m}, filename, nil
}

func main() {
	opt, filename, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	var r io.Reader = os.Stdin
	if filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	c, err := ccwc.Count(r, opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	out := make([]string, 0, 5)
	if opt.CountLines {
		out = append(out, strconv.FormatInt(c.Lines, 10))
	}
	if opt.CountWords {
		out = append(out, strconv.FormatInt(c.Words, 10))
	}
	if opt.CountBytes {
		out = append(out, strconv.FormatInt(c.Bytes, 10))
	}
	if opt.CountChars {
		out = append(out, strconv.FormatInt(c.Chars, 10))
	}
	if len(filename) > 0 {
		out = append(out, filename)
	}

	fmt.Println(strings.Join(out, "\t"))
}
