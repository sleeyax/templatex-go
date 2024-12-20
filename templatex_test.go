package templatex

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
	"testing"
	"text/template"
)

func TestMatchInputToTemplate(t *testing.T) {
	expected := `
			id: d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0
			static-string: "abc"
			random-number: 150
	`
	format := `
			id: {{isUUID}}
			static-string: "abc"
			random-number: {{inRange 100 200}}
	`

	tpl, err := New(template.New("test")).
		Funcs(FuncMap{
			"isUUID": {
				Parse: func(reader *bufio.Reader) ([]string, error) {
					v, err := ReadUntilWhitespaceOrEOF(reader)
					if err != nil {
						return []string{}, err
					}
					return []string{v}, nil
				},
				Validate: func(uuid string) (string, error) {
					if !isValidUUID(uuid) {
						return "", errors.New("invalid UUID")
					}

					return uuid, nil
				},
			},
			"inRange": {
				Parse: func(reader *bufio.Reader) ([]string, error) {
					v, err := ReadUntilWhitespaceOrEOF(reader)
					if err != nil {
						return []string{}, err
					}
					return []string{v}, nil
				},
				Validate: func(value string, min, max int) (any, error) {
					valueAsNumber, err := strconv.Atoi(value)
					if err != nil {
						return "", err
					}

					if valueAsNumber < min || valueAsNumber > max {
						return "", errors.New("value is not in range")
					}

					return value, nil
				},
			},
		}).
		Input(expected).
		Parse(format)

	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, nil)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	actual := buffer.String()
	if expected != actual {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}
