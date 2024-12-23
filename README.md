# templatex

> [!WARNING]  
> This library was developed with a very specific use case in mind. 
> It is not a general purpose library. 
> In 99% of cases, you should stick to the default `text/template` package for templating and/or any of the many 'real' validation libraries for input validation.

This library wraps a subset of go's [text/template](https://pkg.go.dev/text/template) package to use it for input validation. 

Contains zero (0) extra dependencies and no forked code. 

## Example
Consider you are given the following arbitrary input:

```text
id: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
static-string: "abc"
invalid-string: def
random-number: 150
static-string: "def"
```

Then you can validate it as follows:

```go
package main

import (
	"bytes"
	"errors"
	"fmt"
	templatex "github.com/sleeyax/templatex-go"
	"strconv"
	"strings"
	"text/template"
)

func main() {
	// Create a new Go template instance.
	tpl := template.New("example")

	// Create a new templatex instance.
	tplx, err := templatex.New(tpl).
		// Define custom parsing and validation functions.
		// The parser functions are used to extract the value from the input.
		// The validation functions are used to validate the extracted value (as you would define it on a regular `template.FuncMap` from go's `text/template` lib).
		Funcs(templatex.FuncMap{
			"isUUID": {
				// Parses the UUID from between the quotes "<UUID>".
				Parse: templatex.ParseQuotedString,
				// Validates that the parsed value is a valid UUID.
				Validate: func(uuid string) (string, error) {
					if !isValidUUID(uuid) { // bring your own validation library/implementation; this is just an example.
						return "", errors.New("invalid UUID")
					}

					return uuid, nil
				},
			},
			"inRange": {
				// Parses the value until the first whitespace or newline character. "100 " -> "100".
				Parse: templatex.ParseUntilWhiteSpace,
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
	output = strings.TrimSpace(strings.ReplaceAll(output, "\n\t\t\t", "\n")) // clean the output (only needed for this example to work).

	fmt.Println(output)

	// Output:
	// id: "d416e1b0-97b2-4a49-8ad5-2e6b2b46eae0"
	// static-string: "abc"
	// invalid-string: def
	// random-number: 150
}
```

## Contributions
If you want to add a feature or fix, do so yourself by submitting a PR. I currently do not wish to maintain this library any further than I need to for my own use case.
