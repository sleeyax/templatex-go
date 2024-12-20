package templatex

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"text/template"
	"text/template/parse"
)

type ParseFunc func(reader *bufio.Reader) ([]string, error)
type ValidateFunc any

type Func struct {
	Parse    ParseFunc
	Validate ValidateFunc
}

type FuncMap map[string]Func

type Templatex struct {
	tpl        *template.Template
	input      *bufio.Reader
	parseFuncs map[string]ParseFunc
}

func New(tpl *template.Template) *Templatex {
	return &Templatex{
		tpl:        tpl,
		parseFuncs: make(map[string]ParseFunc),
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

func (t *Templatex) Parse(text string) (*Templatex, error) {
	if t.input == nil {
		return nil, errors.New("input required")
	}

	if _, err := t.tpl.Parse(text); err != nil {
		return nil, err
	}

	for _, node := range t.tpl.Tree.Root.Nodes {
		if node.Type() == parse.NodeText {
			textNode := node.(*parse.TextNode)
			text := textNode.Text

			for i := 0; i < len(text); i++ {
				r, _, err := t.input.ReadRune()
				if err != nil {
					if err == io.EOF {
						break
					}

					return nil, err
				}

				if r != rune(text[i]) {
					return nil, errors.New("input does not match template")
				}
			}
		}

		if node.Type() == parse.NodeAction {
			actionNode := node.(*parse.ActionNode)

			pipeNode := actionNode.Pipe
			if pipeNode == nil {
				continue
			}

			if len(pipeNode.Cmds) == 0 {
				continue
			}
			cmd := pipeNode.Cmds[0]

			if len(cmd.Args) == 0 {
				continue
			}
			cmdArg := cmd.Args[0]
			if cmdArg.Type() == parse.NodeIdentifier {
				identifierNode := cmdArg.(*parse.IdentifierNode)
				fn, ok := t.parseFuncs[identifierNode.Ident]
				if !ok {
					continue
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
			}
		}
	}

	return t, nil
}

func (t *Templatex) Execute(wr io.Writer, data any) error {
	return t.tpl.Execute(wr, data)
}
