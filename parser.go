package gmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

func Parse(filename string) (*AstNode, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	return ParseString(filename, string(input))
}

func ParseString(name, template string) (*AstNode, error) {
	return parse(name, lex(template))
}

type parser struct {
	input    <-chan item
	queue    []item
	pos      int
	filename string
}

func parse(filename string, in <-chan item) (*AstNode, error) {
	p := &parser{input: in, filename: filename}
	return parseSection(p, nil)
}

func (p *parser) next() item {
	var i item
	if p.pos == len(p.queue) {
		i = <-p.input
		p.queue = append(p.queue, i)
	} else {
		i = p.queue[p.pos]
	}
	p.pos++
	return i
}

func (p *parser) peek() item {
	i := p.next()
	p.backup()
	return i
}

func (p *parser) backup() {
	p.pos--
}

func (p *parser) backupN(n int) {
	p.pos -= n
}

func (p *parser) syntaxErrorf(i item, msg string, args ...interface{}) error {
	prefix := fmt.Sprintf("%s:%d: ", p.filename, i.line)
	return errors.Errorf(prefix+msg, args...)
}

const unexpTokenErr = "unexpected token read: %s (expected %s)"

func parseTitle(p *parser) (int, *AstNode, error) {
	switch i := p.peek(); i.typ {
	case itemHash:
		prefix := i.val
		depth := len(i.val)
		p.next()
		i = p.peek()
		if i.typ == itemText {
			val := strings.TrimSpace(p.next().val)
			return depth, newAstNode(AstSection, []string{prefix + " " + val}, p.filename, i.line), nil
		}
		return -1, nil, p.syntaxErrorf(i, unexpTokenErr, i, itemText)
	default:
		return -1, nil, p.syntaxErrorf(i, unexpTokenErr, i, itemHash)
	}
}

func parseSection(p *parser, parent *AstNode) (*AstNode, error) {
	var node *AstNode
	for {
		switch i := p.peek(); i.typ {
		case itemText:
			if node == nil {
				return nil, p.syntaxErrorf(i, unexpTokenErr, i, itemHash)
			}
			text := parseText(p)
			node.addChild(text)
		case itemIndent:
			if node == nil {
				return nil, p.syntaxErrorf(i, unexpTokenErr, i, itemHash)
			}
			quote, err := parseCode(p)
			if err != nil {
				return nil, err
			}
			node.addChild(quote)
		case itemArrow:
			if node == nil {
				return nil, p.syntaxErrorf(i, unexpTokenErr, i, itemHash)
			}
			attrs, err := parseQuote(p)
			if err != nil {
				return nil, err
			}
			node.addChild(attrs)
		case itemHash:
			if node == nil {
				depth, sectionNode, err := parseTitle(p)
				switch {
				case err != nil:
					return nil, err
				case depth > 2, depth == 1 && parent != nil, depth == 2 && parent == nil:
					expDepth := 1
					if parent != nil {
						expDepth = 2
					}
					return nil, p.syntaxErrorf(i, "invalid section depth %d (expected %d)", depth, expDepth)
				default:
					node = sectionNode
				}
			} else if parent != nil {
				return node, nil
			} else {
				sectionNode, err := parseSection(p, node)
				if err != nil {
					return nil, err
				}
				node.addChild(sectionNode)
			}
		case itemEmptyLine:
			p.next() // ignore
		case itemEOF:
			return node, nil
		}
	}
}

func parseQuote(p *parser) (*AstNode, error) {
	quotes := []string{}
	line := -1
	for {
		if i := p.peek(); i.typ != itemArrow {
			return newAstNode(AstQuote, quotes, p.filename, line), nil
		} else if line == -1 {
			line = i.line
		}

		p.next()
		if i := p.peek(); i.typ == itemText {
			val := p.next().val
			quotes = append(quotes, strings.TrimSpace(val))
		}
	}
}

func parseText(p *parser) *AstNode {
	lines := []string{}
	line := -1
	for {
		if i := p.peek(); i.typ != itemText {
			return newAstNode(AstText, lines, p.filename, line)
		} else if line == -1 {
			line = i.line
		}

		lines = append(lines, strings.TrimSpace(p.next().val))
	}
}

func parseCode(p *parser) (*AstNode, error) {
	indent := ""
	lines := []string{}
	line := -1
	for {
		if i := p.peek(); i.typ != itemIndent {
			return newAstNode(AstCode, lines, p.filename, line), nil
		} else if line == -1 {
			line = i.line
		}

		ind := p.next()
		if indent == "" {
			indent = ind.val
		}

		// find the prefix belonging to code, not markup
		prefix := strings.TrimPrefix(ind.val, indent)

		txt := p.next()
		if txt.typ != itemText {
			return nil, p.syntaxErrorf(txt, unexpTokenErr, txt, itemIndent)
		}

		// only space on the right (other would be indent).
		val := strings.TrimSpace(txt.val)
		lines = append(lines, prefix+val)
	}
}
