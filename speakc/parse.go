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

type Parser struct {
}

func (p *Parser) ParseFile(filename string) error {
	text, err := readFile(filename)
	if err != nil {
		return err
	}
	lex := Lex(filename, text)
	for {
		item := lex.NextItem()
		if item.Kind == ItemError {
			return fmt.Errorf("error:%s:%d: %v", lex.Name, lex.LineNumber(), item)
		} else {
			fmt.Printf("%v\n", item)
		}
		if item.Kind == ItemEof || item.Kind == ItemError {
			break
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
