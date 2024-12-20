package templatex

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

func ReadUntilWhitespaceOrEOF(reader *bufio.Reader) (string, error) {
	var result strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if r == ' ' || r == '\n' || r == '\t' {
			_ = reader.UnreadRune()
			break
		}

		result.WriteRune(r)
	}

	return result.String(), nil
}

func isValidUUID(u string) bool {
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	return uuidRegex.MatchString(u)
}
