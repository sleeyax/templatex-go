package templatex

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
	"text/template/parse"
)

var (
	// ErrorUnsupportedNode is returned when an unsupported field or node is encountered during parsing.
	ErrorUnsupportedNode = errors.New("unsupported node")
	// ErrorInvalidNode is returned when a supported node is found during parsing but determined to be in an invalid or unsupported format.
	ErrorInvalidNode = errors.New("invalid node")
	// ErrorUnsupportedFunction is returned when an unsupported or unmapped function is encountered during parsing.
	ErrorUnsupportedFunction = errors.New("unsupported or unmapped function")
	// ErrorInputRequired is returned when no input has been provided yet but the Parse method is called.
	ErrorInputRequired = errors.New("input required")
)

type ParseFunc func(reader *bufio.Reader) ([]string, error)
type ValidateFunc any

type Func struct {
	Parse    ParseFunc
	Validate ValidateFunc
}

type FuncMap map[string]Func

type Templatex struct {
	tpl              *template.Template
	input            *bufio.Reader
	parseFuncs       map[string]ParseFunc
	leftDelimLength  int
	rightDelimLength int
}

func New(tpl *template.Template) *Templatex {
	return &Templatex{
		tpl:              tpl,
		parseFuncs:       make(map[string]ParseFunc),
		leftDelimLength:  2,
		rightDelimLength: 2,
	}
}

func (t *Templatex) Template() *template.Template {
	return t.tpl
}

func (t *Templatex) Funcs(funcMap FuncMap) *Templatex {
	for name, fn := range funcMap {
		t.tpl.Funcs(template.FuncMap{
			name: fn.Validate,
		})
		t.parseFuncs[name] = fn.Parse
	}

	return t
}

func (t *Templatex) Input(text string) *Templatex {
	t.input = bufio.NewReader(strings.NewReader(text))
	return t
}

func (t *Templatex) Delims(left, right string) *Templatex {
	t.tpl.Delims(left, right)
	t.leftDelimLength = len(left)
	t.rightDelimLength = len(right)
	return t
}

func (t *Templatex) Parse(text string, data any) (*Templatex, error) {
	if t.input == nil {
		return nil, ErrorInputRequired
	}

	if _, err := t.tpl.Parse(text); err != nil {
		return nil, err
	}

	var previousPosition int
	for _, node := range t.tpl.Tree.Root.Nodes {
		if node.Type() == parse.NodeAction {
			actionNode := node.(*parse.ActionNode)

			pipeNode := actionNode.Pipe
			if pipeNode == nil {
				return nil, ErrorInvalidNode
			}

			if len(pipeNode.Cmds) == 0 {
				return nil, ErrorInvalidNode
			}
			cmd := pipeNode.Cmds[0]

			if len(cmd.Args) == 0 {
				return nil, ErrorInvalidNode
			}
			cmdArg := cmd.Args[0]

			// Discard the bytes we've read so far from the input buffer.
			currentPosition := int(node.Position())
			var bytesRead int
			if previousPosition == 0 {
				bytesRead = currentPosition - t.leftDelimLength
			} else {
				distance := currentPosition - previousPosition
				delimLen := t.leftDelimLength + t.rightDelimLength
				bytesRead = distance - delimLen
			}
			if _, err := t.input.Discard(bytesRead); err != nil {
				return nil, err
			}
			previousPosition = currentPosition + len(cmd.String())

			if cmdArg.Type() == parse.NodeIdentifier {
				// If it's an identifier, such as {{isUUID}}, we need to parse the argument.
				// This is accomplished by calling the Parse function associated with the identifier.
				// Finally, we replace the argument node with the parsed arguments.

				identifierNode := cmdArg.(*parse.IdentifierNode)
				fn, ok := t.parseFuncs[identifierNode.Ident]
				if !ok {
					// TODO: assume this is a builtin function, which should be evaluated instead
					return nil, ErrorUnsupportedFunction
				}

				args, err := fn(t.input)
				if err != nil {
					return nil, err
				}

				var newArgNodes []parse.Node
				for _, arg := range args {
					newArgNodes = append(newArgNodes, &parse.StringNode{
						NodeType: parse.NodeString,
						Text:     arg,
					})
				}

				cmd.Args = append(cmd.Args[:1], append(newArgNodes, cmd.Args[1:]...)...)
			} else if cmdArg.Type() == parse.NodeField {
				// If it's a regular field, such as {{.Foo}} or {{.Bar.Baz}}, we need to evaluate it to determine the amount of bytes to discard in order to advance the input buffer.

				fieldNode := cmdArg.(*parse.FieldNode)

				tpl, err := template.New("field").Parse(fmt.Sprintf("{{.%s}}", strings.Join(fieldNode.Ident, ".")))
				if err != nil {
					return nil, fmt.Errorf("failed to evaluate field: %w", err)
				}

				var buffer bytes.Buffer
				if err = tpl.Execute(&buffer, data); err != nil {
					return nil, fmt.Errorf("failed to evaluate field: %w", err)
				}

				if _, err := t.input.Discard(buffer.Len()); err != nil {
					return nil, fmt.Errorf("failed to discard bytes read from field: %w", err)
				}
			} else {
				return nil, ErrorUnsupportedNode
			}
		}
	}

	return t, nil
}

func (t *Templatex) Execute(wr io.Writer, data any) error {
	return t.tpl.Execute(wr, data)
}
