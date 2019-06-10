// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package ftp

import (
	"fmt"
	"log"
)

type Logger interface {
	Print(message interface{})
	Printf(format string, v ...interface{})
	PrintCommand(command string, params ...interface{})
	PrintResponse(code int, message interface{})
}

// Use an instance of this to log in a standard format
type StdLogger struct{}

var _ Logger = new(StdLogger)

func (logger *StdLogger) Print(message interface{}) {
	log.Printf("%s", message)
}

func (logger *StdLogger) Printf(format string, v ...interface{}) {
	logger.Print(fmt.Sprintf(format, v...))
}

func (logger *StdLogger) PrintCommand(command string, params ...interface{}) {
	if command == "PASS" {
		log.Printf("> PASS ****")
	} else {
		log.Printf("> %s %s", command, params)
	}
}

func (logger *StdLogger) PrintResponse(code int, message interface{}) {
	log.Printf("< %d %s", code, message)
}

// Silent logger, produces no output
type DiscardLogger struct{}

var _ Logger = new(DiscardLogger)

func (logger *DiscardLogger) Print(message interface{})                          {}
func (logger *DiscardLogger) Printf(format string, v ...interface{})             {}
func (logger *DiscardLogger) PrintCommand(command string, params ...interface{}) {}
func (logger *DiscardLogger) PrintResponse(code int, message interface{})        {}
