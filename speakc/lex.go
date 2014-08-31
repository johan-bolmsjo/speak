// Copyright 2011 The Go Authors. All rights reserved.
// Copyright 2013 Johan Bolmsjö
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.go-derived file.

package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Identifies the type of item.
type ItemKind int

// Types of item.
const (
	ItemError ItemKind = iota
	ItemIdentifier
	ItemNumber
	ItemEol
	ItemEof
	ItemLeftBracket
	ItemRightBracket
	ItemDot
	ItemColon
	ItemChoice
	ItemEnd
	ItemEnum
	ItemMessage
	ItemPackage
	ItemType
	ItemBasicTypeBegin
	ItemBool
	ItemByte
	ItemInt8
	ItemInt16
	ItemInt32
	ItemInt64
	ItemUint8
	ItemUint16
	ItemUint32
	ItemUint64
	ItemFloat32
	ItemFloat64
	ItemString
	ItemBasicTypeEnd
)

var itemKindToStr = map[ItemKind]string{
	ItemError:        "<error>",
	ItemIdentifier:   "<identifier>",
	ItemNumber:       "<number>",
	ItemEol:          "<eol>",
	ItemEof:          "<eof>",
	ItemLeftBracket:  "[",
	ItemRightBracket: "]",
	ItemDot:          ".",
	ItemColon:        ":",
	ItemChoice:       "choice",
	ItemEnd:          "end",
	ItemEnum:         "enum",
	ItemMessage:      "message",
	ItemPackage:      "package",
	ItemType:         "type",
	ItemBool:         "bool",
	ItemByte:         "byte",
	ItemInt8:         "int8",
	ItemInt16:        "int16",
	ItemInt32:        "int32",
	ItemInt64:        "int64",
	ItemUint8:        "uint8",
	ItemUint16:       "uint16",
	ItemUint32:       "uint32",
	ItemUint64:       "uint64",
	ItemFloat32:      "float32",
	ItemFloat64:      "float64",
	ItemString:       "string",
}

var strToItemKind = map[string]ItemKind{
	"choice":  ItemChoice,
	"end":     ItemEnd,
	"enum":    ItemEnum,
	"message": ItemMessage,
	"package": ItemPackage,
	"type":    ItemType,
	"bool":    ItemBool,
	"byte":    ItemByte,
	"int8":    ItemInt8,
	"int16":   ItemInt16,
	"int32":   ItemInt32,
	"int64":   ItemInt64,
	"uint8":   ItemUint8,
	"uint16":  ItemUint16,
	"uint32":  ItemUint32,
	"uint64":  ItemUint64,
	"float32": ItemFloat32,
	"float64": ItemFloat64,
	"string":  ItemString,
}

func (kind ItemKind) String() string {
	if s, ok := itemKindToStr[kind]; ok {
		return s
	}
	return fmt.Sprintf("%d", int(kind))
}

// Check if an item kind is a basic type.
func (kind ItemKind) isBasicType() bool {
	if kind > ItemBasicTypeBegin && kind < ItemBasicTypeEnd {
		return true
	}
	return false
}

// item represents a token or text string returned from the scanner.
type Item struct {
	Kind  ItemKind // The type of this item.
	Value string   // The value of this item.
	Pos   int      // The starting position, in bytes, of this item in the input string.
}

func (item Item) String() string {
	if item.Kind == ItemError || item.Kind == ItemIdentifier || item.Kind == ItemNumber {
		return item.Value
	}
	return fmt.Sprintf("%v", item.Kind)
}

const eof = -1

type stateFn func(*Lexer) stateFn

type Lexer struct {
	Name  string    // Name of lexer for error reporting.
	input string    // The string being scanned.
	state stateFn   // The next lexing function to enter.
	pos   int       // Current position in input.
	start int       // Start position of item in input.
	width int       // Width of last rune read from input.
	items chan Item // Scanned items.
}

// Returns the next rune in the input.
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

// Returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// Steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// Passes a item back to the client.
func (l *Lexer) emit(kind ItemKind) {
	l.items <- Item{kind, l.acceptStr(), l.start}
	l.start = l.pos
}

// Skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// Consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// Consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// Returns the currently accepted string.
func (l *Lexer) acceptStr() string {
	return l.input[l.start:l.pos]
}

// Returns the current accept length.
func (l *Lexer) acceptLen() int {
	return l.pos - l.start
}

// Report the line number that item was from.
func (l *Lexer) LineNumber(item Item) int {
	if item.Kind == ItemEof {
		return 1 + strings.Count(l.input, "\n")
	} else {
		line := 1 + strings.Count(l.input[:item.Pos], "\n")
		if isEol(rune(l.input[item.Pos])) {
			line++
		}
		return line
	}
}

// Report the column number that item was from.
func (l *Lexer) ColumnNumber(item Item) int {
	column := -1
	pos := item.Pos
	if item.Kind == ItemEof {
		if pos > 0 {
			pos--
			column++
		} else {
			return 0
		}
	}
	for i := pos; i >= 0; i-- {
		c := rune(l.input[i])
		if isEol(c) {
			break
		}
		if c&0x80 == 0 {
			/* utf8 start character */
			column++
		}
	}
	return column
}

// Returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{ItemError, fmt.Sprintf(format, args...), l.start}
	return nil
}

// nextItem returns the next item from the input.
func (l *Lexer) NextItem() Item {
	item := <-l.items
	return item
}

// Creates a new scanner for the input string.
func NewLexer(name, input string) *Lexer {
	l := &Lexer{
		Name:  name,
		input: input,
		items: make(chan Item),
	}
	go l.run()
	return l
}

// Runs the state machine for the lexer.
func (l *Lexer) run() {
	for l.state = lexRoot; l.state != nil; {
		l.state = l.state(l)
	}
}

//
// State functions
//

// Top level lexer.
func lexRoot(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(ItemEof)
			return nil
		case r == '/' && l.peek() == '/':
			l.next()
			return lexComment
		case isEol(r):
			return lexEol
		case isSpace(r):
			return lexSpace
		case r == '[':
			l.emit(ItemLeftBracket)
		case r == ']':
			l.emit(ItemRightBracket)
		case r == '.':
			l.emit(ItemDot)
		case r == ':':
			l.emit(ItemColon)
		case isLetter(r):
			return lexIdentifier
		case isDigit(r):
			return lexNumber
		default:
			return l.errorf("unrecognized character: %#U", r)
		}
	}
}

// Scans a run of space characters.
// One space has already been seen.
func lexSpace(l *Lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexRoot
}

// Scans a run of EOL characters.
// One EOL character has already been seen.
func lexEol(l *Lexer) stateFn {
	for isEol(l.peek()) {
		l.next()
	}
	l.emit(ItemEol)
	return lexRoot
}

// Scans characters until EOL or EOF.
// The comment marker '//' has already been seen.
func lexComment(l *Lexer) stateFn {
	for r := l.peek(); !isEol(r) && r != eof; r = l.peek() {
		l.next()
	}
	l.ignore()
	return lexRoot
}

// Scans identifiers and keywords.
func lexIdentifier(l *Lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			// Check if the scanned token is an identifier or some other item kind.
			l.backup()
			itemKind := strToItemKind[l.acceptStr()]
			switch {
			case itemKind != 0:
				l.emit(itemKind)
			default:
				l.emit(ItemIdentifier)
			}
			break Loop
		}
	}
	return lexRoot
}

// Scans a positive decimal number.
func lexNumber(l *Lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.acceptStr())
	}
	l.emit(ItemNumber)
	return lexRoot
}

func (l *Lexer) scanNumber() bool {
	l.acceptRun("0123456789")

	// The first digit must not be '0' if there are more than one digits.
	if l.acceptLen() > 1 && l.input[l.start] == '0' {
		return false
	}

	// Do some basic validation of the character that follows the last digit.
	r := l.peek()
	if isLetter(r) {
		l.next()
		return false
	}
	return true
}

// Reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// Reports whether r is an EOL character.
func isEol(r rune) bool {
	return r == '\r' || r == '\n'
}

// Reports whether r is a digit.
func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}

// Reports whether r is a letter.
func isLetter(r rune) bool {
	return ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}

// Reports whether r is an alphabetic, digit.
func isAlphaNumeric(r rune) bool {
	return isLetter(r) || isDigit(r)
}
