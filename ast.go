package gmd

import (
	"fmt"

	"github.com/pkg/errors"
)

type astNodeType int

const (
	AstSection astNodeType = iota
	AstCode
	AstText
	AstQuote
)

func (t astNodeType) String() string {
	switch t {
	case AstSection:
		return "AstSection"
	case AstCode:
		return "AstCode"
	case AstText:
		return "AstText"
	case AstQuote:
		return "AstQuote"
	default:
		return "unknown"
	}
}

type AstNode struct {
	Children []*AstNode

	Type   astNodeType
	Lines  []string
	LineNo int
	File   string
}

func newAstNode(typ astNodeType, lines []string, file string, idx int) *AstNode {
	return &AstNode{
		Type:   typ,
		Lines:  lines,
		LineNo: idx,
		File:   file,
	}
}

func (n *AstNode) addChild(child *AstNode) {
	n.Children = append(n.Children, child)
}

func (n *AstNode) Errorf(i int, msgf string, args ...interface{}) error {
	msg := fmt.Sprintf(msgf, args...)
	return errors.Errorf("%s:%d: %s", n.File, n.LineNo+i, msg)
}
