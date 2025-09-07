package ccwc

import (
	"bufio"
	"io"
	"unicode"
)

type Counts struct {
	Bytes int64
	Lines int64
	Words int64
	Chars int64
}

type Options struct {
	CountBytes bool
	CountLines bool
	CountWords bool
	CountChars bool
}

func Count(r io.Reader, opt Options) (Counts, error) {
	br := bufio.NewReader(r)
	var c Counts
	prevIsSpace := true

	for {
		rn, size, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c, err
		}

		if opt.CountBytes {
			c.Bytes += int64(size)
		}
		if opt.CountLines && rn == '\n' {
			c.Lines++
		}
		if opt.CountWords {
			isSpace := unicode.IsSpace(rn)
			if !isSpace && prevIsSpace {
				c.Words++
			}
			prevIsSpace = isSpace
		}
		if opt.CountChars {
			c.Chars++
		}
	}

	return c, nil
}
