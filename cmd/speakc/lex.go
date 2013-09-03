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
type itemKind int

// Types of item.
const (
	itemError itemKind = iota
	itemIdentifier
	itemNumber
	itemEol
	itemEof
	itemLeftBracket
	itemRightBracket
	itemDot
	itemColon
	itemEnd
	itemEnum
	itemMessage
	itemPackage
	itemType
	itemBasicTypeBegin
	itemBool
	itemByte
	itemInt8
	itemInt16
	itemInt32
	itemInt64
	itemUint8
	itemUint16
	itemUint32
	itemUint64
	itemFloat32
	itemFloat64
	itemString
	itemBasicTypeEnd
)

var itemKindToStr = map[itemKind]string{
	itemError:        "<error>",
	itemIdentifier:   "<identifier>",
	itemNumber:       "<number>",
	itemEol:          "<eol>",
	itemEof:          "<eof>",
	itemLeftBracket:  "[",
	itemRightBracket: "]",
	itemDot:          ".",
	itemColon:        ":",
	itemEnd:          "end",
	itemEnum:         "enum",
	itemMessage:      "message",
	itemPackage:      "package",
	itemType:         "type",
	itemBool:         "bool",
	itemByte:         "byte",
	itemInt8:         "int8",
	itemInt16:        "int16",
	itemInt32:        "int32",
	itemInt64:        "int64",
	itemUint8:        "uint8",
	itemUint16:       "uint16",
	itemUint32:       "uint32",
	itemUint64:       "uint64",
	itemFloat32:      "float32",
	itemFloat64:      "float64",
	itemString:       "string",
}

var strToItemKind = map[string]itemKind{
	"end":     itemEnd,
	"enum":    itemEnum,
	"message": itemMessage,
	"package": itemPackage,
	"type":    itemType,
	"bool":    itemBool,
	"byte":    itemByte,
	"int8":    itemInt8,
	"int16":   itemInt16,
	"int32":   itemInt32,
	"int64":   itemInt64,
	"uint8":   itemUint8,
	"uint16":  itemUint16,
	"uint32":  itemUint32,
	"uint64":  itemUint64,
	"float32": itemFloat32,
	"float64": itemFloat64,
	"string":  itemString,
}

func (kind itemKind) String() string {
	if s, ok := itemKindToStr[kind]; ok {
		return s
	}
	return fmt.Sprintf("%d", int(kind))
}

// Check if an item kind is a basic type.
func (kind itemKind) isBasicType() bool {
	if kind > itemBasicTypeBegin && kind < itemBasicTypeEnd {
		return true
	}
	return false
}

// item represents a token or text string returned from the scanner.
type item struct {
	kind  itemKind // The type of this item.
	pos   int      // The starting position, in bytes, of this item in the input string.
	value string   // The value of this item.
}

func (item item) String() string {
	switch {
	case item.kind == itemError:
		return item.value
	case item.kind == itemIdentifier || item.kind == itemNumber:
		return fmt.Sprintf("%v:%v", item.kind, item.value)
	}
	return fmt.Sprintf("%v", item.kind)
}

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	name    string    // Name of lexer for error reporting.
	input   string    // The string being scanned.
	state   stateFn   // The next lexing function to enter.
	pos     int       // Current position in input.
	start   int       // Start position of item in input.
	width   int       // Width of last rune read from input.
	lastPos int       // Position of most recent item returned by nextItem
	items   chan item // Scanned items.
}

// Returns the next rune in the input.
func (l *lexer) next() rune {
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
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// Steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// Passes a item back to the client.
func (l *lexer) emit(kind itemKind) {
	l.items <- item{kind, l.start, l.acceptStr()}
	l.start = l.pos
}

// Skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// Consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// Consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// Returns the currently accepted string.
func (l *lexer) acceptStr() string {
	return l.input[l.start:l.pos]
}

// Returns the current accept length.
func (l *lexer) acceptLen() int {
	return l.pos - l.start
}

// Reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// Returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// Creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// Runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexRoot; l.state != nil; {
		l.state = l.state(l)
	}
}

//
// State functions
//

// Top level lexer.
func lexRoot(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemEof)
			return nil
		case r == '/' && l.peek() == '/':
			l.next()
			return lexComment
		case isEol(r):
			return lexEol
		case isSpace(r):
			return lexSpace
		case r == '[':
			l.emit(itemLeftBracket)
		case r == ']':
			l.emit(itemRightBracket)
		case r == '.':
			l.emit(itemDot)
		case r == ':':
			l.emit(itemColon)
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
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexRoot
}

// Scans a run of EOL characters.
// One EOL character has already been seen.
func lexEol(l *lexer) stateFn {
	for isEol(l.peek()) {
		l.next()
	}
	l.emit(itemEol)
	return lexRoot
}

// Scans characters until EOL or EOF.
// The comment marker '//' has already been seen.
func lexComment(l *lexer) stateFn {
	for r := l.peek(); !isEol(r) && r != eof; r = l.peek() {
		l.next()
	}
	l.ignore()
	return lexRoot
}

// Scans identifiers and keywords.
func lexIdentifier(l *lexer) stateFn {
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
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}
	return lexRoot
}

// Scans a positive decimal number.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.acceptStr())
	}
	l.emit(itemNumber)
	return lexRoot
}

func (l *lexer) scanNumber() bool {
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
