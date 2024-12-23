package templatex

import "bufio"

var ParseQuotedString ParseFunc = func(reader *bufio.Reader) ([]string, error) {
	v, err := ReadQuotedString(reader)
	if err != nil {
		return []string{}, err
	}
	return []string{v}, nil
}

var ParseUntilWhiteSpace ParseFunc = func(reader *bufio.Reader) ([]string, error) {
	v, err := ReadUntilWhitespace(reader)
	if err != nil {
		return []string{}, err
	}
	return []string{v}, nil
}
