// Copyright 2014 Johan Bolmsjö
//
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
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
	lexer  *Lexer
	prev   Item    // Previous item from lexer (accepted).
	next   Item    // Next item from lexer (to be accepted).
	errors []error // Errors found by the lexer or parser.
	pkg    string  // Package that is being parsed.
}

func (p *Parser) ParseFile(filename string) (bool, []error) {
	text, err := readFile(filename)
	if err != nil {
		p.errors = append(p.errors, err)
		return false, p.errors
	}
	return p.ParseText(filename, text)
}

func (p *Parser) ParseText(name, text string) (bool, []error) {
	p.lexer = NewLexer(name, text)
	/* Seed the parser by fetching the first token from the lexer. */
	p.next = p.lexer.NextItem()
	p.parseRoot()
	return p.ok(), p.errors
}

// ----------------------------------------------------------------------------

// Get the next item from the lexer.
func (p *Parser) consume() {
	p.prev = p.next
	if p.next.Kind != ItemEof && p.next.Kind != ItemError {
		p.next = p.lexer.NextItem()
	}
}

// Accept the next item if it's of the specified kind.
func (p *Parser) accept(kind ItemKind) bool {
	if p.next.Kind != kind {
		return false
	}
	p.consume()
	return true
}

// Same as accept but let the supplied function do the matching.
func (p *Parser) acceptM(fn func(Item) error) bool {
	if err := fn(p.next); err != nil {
		return false
	}
	p.consume()
	return true
}

// Expect the next item to be of the specified kind, if it's not an error will
// be pushed onto the parsers error list.
func (p *Parser) expect(kind ItemKind) bool {
	if p.next.Kind != kind {
		p.itemError(p.next, fmt.Errorf("expected %s", kind))
		return false
	}
	p.consume()
	return true
}

// Same as expect but let the supplied function do the matching.
func (p *Parser) expectM(fn func(Item) error) bool {
	if err := fn(p.next); err != nil {
		p.itemError(p.next, err)
		return false
	}
	p.consume()
	return true
}

// Check the parser error state.
func (p *Parser) ok() bool {
	return len(p.errors) == 0
}

// TODO: Need to figure out how to do error handling.
func (p *Parser) itemError(item Item, postfix error) {
	err := error(nil)
	line := p.lexer.LineNumber(item)
	column := p.lexer.ColumnNumber(item)
	if item.Kind == ItemError {
		err = fmt.Errorf("%s:%d:%d: error: %v", p.lexer.Name, line, column, item)
	} else {
		if postfix == nil {
			postfix = errors.New("unexpected token")
		}
		err = fmt.Errorf("%s:%d:%d: error: at '%v', %s.", p.lexer.Name, line, column, item, postfix)
	}
	p.errors = append(p.errors, err)
}

// ----------------------------------------------------------------------------

// BigIdentifier match function.
func matchBigIdentifier(item Item) error {
	if item.Kind == ItemIdentifier {
		r := item.Value[0]
		if 'A' <= r && r <= 'Z' {
			return nil
		}
	}
	return errors.New("expected capitalized identifier")
}

// LittleIdentifier match function.
func matchLittleIdentifier(item Item) error {
	if item.Kind == ItemIdentifier {
		r := item.Value[0]
		if 'a' <= r && r <= 'z' {
			return nil
		}
	}
	return errors.New("expected uncapitalized identifier")
}

// BasicType match function.
func matchBasicType(item Item) error {
	if item.Kind > ItemBasicTypeBegin && item.Kind < ItemBasicTypeEnd {
		return nil
	}
	return errors.New("expected basic type")
}

// ----------------------------------------------------------------------------

// Top level parser.
func (p *Parser) parseRoot() {
out:
	for p.ok() {
		switch {
		case p.accept(ItemEol):
		case p.accept(ItemChoice):
			p.parseChoice()
		case p.accept(ItemEnum):
			p.parseEnum()
		case p.accept(ItemMessage):
			p.parseMessage()
		case p.accept(ItemPackage):
			p.parsePackage()
		case p.accept(ItemType):
			p.parseType()
		case p.accept(ItemEof):
			break out
		default:
			p.itemError(p.next, nil)
		}
	}
}

// ----------------------------------------------------------------------------

func (p *Parser) parseChoice() {
	if p.expectM(matchBigIdentifier) && p.expect(ItemEol) {
		for p.ok() && !p.accept(ItemEnd) {
			p.parseChoiceField()
		}
	}
}

func (p *Parser) parseChoiceIdentifier() bool {
	p.expect(ItemIdentifier)
	item0 := p.prev
	if p.accept(ItemDot) {
		// <package> . BigIdentifier
		p.expectM(matchBigIdentifier)
	} else {
		// BigIdentifier
		if err := matchBigIdentifier(item0); err != nil {
			p.itemError(item0, err)
		}
	}
	return p.ok()
}

func (p *Parser) parseChoiceField() {
	_ = p.expect(ItemNumber) && p.expect(ItemColon) && p.parseChoiceIdentifier() && p.expect(ItemEol)
}

// ----------------------------------------------------------------------------

func (p *Parser) parseEnum() {
	if p.expectM(matchBigIdentifier) && p.expect(ItemEol) {
		for p.ok() && !p.accept(ItemEnd) {
			p.parseEnumField()
		}
	}
}

func (p *Parser) parseEnumField() {
	_ = p.expect(ItemNumber) && p.expect(ItemColon) && p.expectM(matchBigIdentifier) && p.expect(ItemEol)
}

// ----------------------------------------------------------------------------

func (p *Parser) parseMessage() {
	if p.expectM(matchBigIdentifier) && p.expect(ItemEol) {
		for p.ok() && !p.accept(ItemEnd) {
			p.parseMessageField()
		}
	}
}

func (p *Parser) parseMessageField() {
	if p.expect(ItemNumber) && p.expect(ItemColon) && p.expectM(matchLittleIdentifier) {
		_ = p.parseArray() && p.parseTypeIdentifier() && p.expect(ItemEol)
	}
}

func (p *Parser) parseArray() bool {
	if p.accept(ItemLeftBracket) {
		// TODO: check that number > 0 if present.
		p.accept(ItemNumber)
		p.expect(ItemRightBracket)
	}
	return p.ok()
}

func (p *Parser) parseTypeIdentifier() bool {
	if p.acceptM(matchBasicType) {
	} else {
		// TODO: Rename this function if it's to be shared between message and choice parsing.
		p.parseChoiceIdentifier()
	}
	return p.ok()
}

// ----------------------------------------------------------------------------

func (p *Parser) parsePackage() {
	if p.expect(ItemIdentifier) {
		p.pkg = p.prev.Value
		p.expect(ItemEol)
	}
}

// ----------------------------------------------------------------------------

func (p *Parser) parseType() {
	_ = p.expectM(matchBigIdentifier) && p.parseArray() && p.expectM(matchBasicType) && p.expect(ItemEol)
}

// ----------------------------------------------------------------------------
