# gmd - simple markdown parser [![GoDoc](https://godoc.org/github.com/gfrey/gmd?status.svg)](http://godoc.org/github.com/gfrey/gmd) [![Report card](https://goreportcard.com/badge/github.com/gfrey/gmd)](https://goreportcard.com/report/github.com/gfrey/gmd)

A simple parser for markdown documents. The returned AST will have nodes for
sections (lines beginning with a number of `#`), code (lines indented with
whitespace), quotes (lines starting with a `>`) and text (everything else).
There is no support for text internal special formatting (like bold text,
links, ...).

This library is used in [smutje](https://github.com/gfrey/smutje) to parse the
smutje markdown (`smd` suffix) files.

