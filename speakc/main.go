// Copyright 2013 Johan Bolmsjö
//
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// ----------------------------------------------------------------------------

var usageMessage = `usage: speakc [-h] -lang c|go speak-files

Generate serialization code from speak interface definition files.

Options:
    -h           Display this text.
    -lang        Generate code for the specified language (c|go).
    speak-files  Speak source files.

Example:

    speakc -lang c *.speak
`

// ----------------------------------------------------------------------------

type flags struct {
	help       bool
	lang       string
	speakFiles []string
}

func (f *flags) Parse() error {
	flag.BoolVar(&f.help, "h", false, "help message")
	flag.StringVar(&f.lang, "lang", "", "language to generate code for")

	err := error(nil)
	flag.Usage = func() {
		err = errors.New(usageMessage)
	}
	if flag.Parse(); err != nil {
		return err
	}
	if f.help {
		return errors.New(usageMessage)
	}

	var missing []string
	if f.lang == "" {
		missing = append(missing, "-lang")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing argument(s): %s", strings.Join(missing, ","))
	}

	if f.lang != "c" && f.lang != "go" {
		return fmt.Errorf("unsupported target language '%s'.", f.lang)
	}

	for _, arg := range flag.Args() {
		f.speakFiles = append(f.speakFiles, arg)
	}

	return nil
}

// ----------------------------------------------------------------------------

func main() {
	var f flags
	if err := f.Parse(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	parser := new(Parser)
	for _, filename := range f.speakFiles {
		if ok, errors := parser.ParseFile(filename); !ok {
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "%s\n", err)
			}
			os.Exit(1)
		}
	}
}

// ----------------------------------------------------------------------------
