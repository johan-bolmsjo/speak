// Copyright 2014 Johan Bolmsjö
//
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
)

// ----------------------------------------------------------------------------

func readFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ----------------------------------------------------------------------------

// Parser holds the state from parsing one or more files.
type Parser struct {
	lexer *Lexer
	item  Item
}

// Get the next item from the lexer.
// Updates p.item or returns an error from the lexer.
func (p *Parser) nextItem() error {
	p.item = p.lexer.NextItem()
	if p.item.Kind == ItemError {
		return fmt.Errorf("error:%s:%d: %v", p.lexer.Name, p.lexer.LineNumber(), p.item)
	}
	return nil
}

func (p *Parser) ParseText(name, text string) error {
	p.lexer = Lex(name, text)
	for p.item.Kind != ItemEof {
		if err := p.nextItem(); err != nil {
			return err
		}
		fmt.Printf("%v\n", p.item)
	}
	return nil
}

func (p *Parser) ParseFile(filename string) error {
	text, err := readFile(filename)
	if err != nil {
		return err
	}
	return p.ParseText(filename, text)
}

// ----------------------------------------------------------------------------
