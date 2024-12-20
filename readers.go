package templatex

import (
	"bufio"
	"io"
	"strings"
)

func ReadUntil(reader *bufio.Reader, delimiters []rune) (string, error) {
	var result strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		for _, delimiter := range delimiters {
			if r == delimiter {
				_ = reader.UnreadRune()
				return result.String(), nil
			}
		}

		result.WriteRune(r)
	}

	return result.String(), nil
}

func ReadUntilWhitespace(reader *bufio.Reader) (string, error) {
	return ReadUntil(reader, []rune{' ', '\t', '\n', '\r'})
}

func ReadQuotedString(reader *bufio.Reader) (string, error) {
	return ReadUntil(reader, []rune{'"', '\''})
}
