package templatex

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"text/template"
)

type testCase struct {
	name           string
	input          string
	template       string
	mustError      bool
	mustErrorMatch error
}

type data struct {
	Foo int
	Bar int
}

func TestTemplatex_Parse(t *testing.T) {
	testCases := []testCase{
		{
			name:     "Empty",
			input:    "",
			template: "",
		},
		{
			name: "Uuids",
			input: `
				id1: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
				id2: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
				id3: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
			`,
			template: `
				id1: "{{isUUID}}"
				id2: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
				id3: "{{isUUID}}"
			`,
		},
		{
			name: "Number and UUID",
			input: `
				random-number: 150
				id1: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
			`,
			template: `
				random-number: {{inRange 100 200}}
				id1: "{{isUUID}}"
			`,
		},
		{
			name: "Simple mix",
			input: `
				id: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
				static-string: "abc"
				random-number: 150
				static-string: "def"
			`,
			template: `
				id: "{{isUUID}}"
				static-string: "abc"
				random-number: {{inRange 100 200}}
				static-string: "def"
			`,
		},
		{
			name: "Regular template with variables",
			input: `
				foo: 1
				bar: 2
			`,
			template: `
				foo: {{.Foo}}
				bar: {{.Bar}}
			`,
		},
		{
			name: "Validation template with variables",
			input: `
				foo: 1
				uuid: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
				bar: 2
			`,
			template: `
				foo: {{.Foo}}
				uuid: "{{isUUID}}"
				bar: {{.Bar}}
			`,
		},
		{
			name: "Invalid function",
			input: `
				oops: abc
			`,
			template: `
				oops: {{isNonExistent}}
			`,
			mustError: true,
		},
		{
			name: "unsupported node",
			input: `
				foo: 1
			`,
			template: `
				foo: {{- 45}}
			`,
			mustError:      true,
			mustErrorMatch: ErrorUnsupportedNode,
		},
		{
			name: "Input doesn't match template",
			input: `
				oops: 11a40eea-1a46-476c-b0e9-b301c690a115
			`,
			template: `
				my-string: {{isUUID}}
			`,
			mustError:      true,
			mustErrorMatch: ErrorInputValidation,
		},
		{
			name: "Input in wrong order",
			input: `
				a: 1
				b: 1
				c: 1
			`,
			template: `
				c: 1
				b: 1
				a: 1
			`,
			mustError:      true,
			mustErrorMatch: ErrorInputValidation,
		},
	}

	d := data{Foo: 1, Bar: 2}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := New(template.New("test")).
				Funcs(FuncMap{
					"isUUID": {
						Parse: ParseQuotedString,
						Validate: func(uuid string) (string, error) {
							if !isValidUUID(uuid) {
								return "", errors.New("invalid UUID")
							}

							return uuid, nil
						},
					},
					"inRange": {
						Parse: ParseUntilWhiteSpace,
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
				Data(d).
				Input(tc.input).
				Parse(tc.template)

			if tc.mustError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if tc.mustErrorMatch != nil && !errors.Is(err, tc.mustErrorMatch) {
					t.Fatalf("expected error %v, got %v", tc.mustErrorMatch, err)
				}
			} else {
				if err != nil {
					t.Fatalf("failed to parse template: %v", err)
				}

				var buffer bytes.Buffer
				if err = tpl.Execute(&buffer, d); err != nil {
					t.Fatalf("failed to execute template: %v", err)
				}

				if actual := buffer.String(); tc.input != actual {
					t.Fatalf("expected %s, got %s", tc.input, actual)
				}
			}
		})
	}
}

func ExampleTemplatex_Parse() {
	// Create a new Go template instance.
	tpl := template.New("example")

	// Create a new templatex instance.
	tplx, err := New(tpl).
		// Define custom parsing and validation functions.
		// The parser functions are used to extract the value from the input.
		// The validation functions are used to validate the extracted value (as you would define it on a regular `template.FuncMap` from go's `text/template` lib).
		Funcs(FuncMap{
			"isUUID": {
				// Parses the UUID from between the quotes "<UUID>".
				Parse: ParseQuotedString,
				// Validates that the parsed value is a valid UUID.
				Validate: func(uuid string) (string, error) {
					if !isValidUUID(uuid) {
						return "", errors.New("invalid UUID")
					}

					return uuid, nil
				},
			},
			"inRange": {
				// Parses the value until the first whitespace or newline character. "100 " -> "100".
				Parse: ParseUntilWhiteSpace,
				// Validates that the parsed value is an integer within the specified range.
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
		// Provide input data that should be verified using the template below.
		Input(`
			id: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
			static-string: "abc"
			invalid-string: def
			random-number: 150
		`).
		// Provide the template that should be used to verify the input data.
		// Keep in mind that it supports only a subset of the Go template syntax.
		// You'll gracefully receive an error if you use unsupported syntax.
		Parse(`
			id: "{{isUUID}}"
			static-string: "abc"
			invalid-string: def
			random-number: {{inRange 100 200}}
		`)

	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	if err = tplx.Execute(&buffer, nil); err != nil {
		panic(err)
	}

	output := buffer.String()
	output = strings.TrimSpace(strings.ReplaceAll(output, "\n\t\t\t", "\n")) // clean the output, only needed for the example.

	fmt.Println(output)

	// Output:
	// id: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
	// static-string: "abc"
	// invalid-string: def
	// random-number: 150
}
