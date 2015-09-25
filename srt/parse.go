package srt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Rect stores where text is shown in the video frame
type Rect struct {
	x, y, w, h int
}

// Strip represents a single strip of subtitle
type Strip struct {
	Sequende uint
	Start    time.Time
	Duration time.Duration
	Lines    []string
	Rect     *Rect
}

// Result wraps a sucessfully parsed subtitle Strip or an error
type Result struct {
	Strip *Strip
	Error error
}

type state struct {
	tokens   chan<- token
	scanner  *bufio.Scanner
	buffer   []byte
	position int
}

type token struct {
	typ int
	val interface{}
	pos int
}

const (
	tokenInteger = iota
	tokenTime
	tokenRightArrow
	tokenColon
	tokenLine
	tokenEnd
)

type lexer func(l *state) lexer

func (s *state) advance() {
	if s.scanner.Scan() {
		s.buffer = s.scanner.Bytes()
	} else {
		s.buffer = nil
	}
}

func (s *state) consume(n int) {
	if n >= len(s.buffer) {
		n -= len(s.buffer)
		s.advance()
	}

	if s.buffer != nil {
		s.buffer = s.buffer[n:]
	}

	s.position += n
}

func (s *state) input() string {
	return string(s.buffer)
}

func (s *state) emit(tok int, val interface{}) {
	s.tokens <- token{tok, val, s.position}
}

func lexTime(l *state) lexer {
	return nil
}

func lexInteger(l *state) lexer {
	i, err := strconv.ParseInt(l.input(), 10, 64)
	if err != nil {
		return nil
	}

	l.emit(tokenInteger, i)
	return lexTime
}

func lex(tokens chan token, input io.Reader) {
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)

	state := &state{tokens, scanner, nil, 0}

	for l := lexInteger; l != nil; {
		l = l(state)
	}
	close(tokens)
}

func parseStrip(tokens <-chan token) (*Strip, error) {
	return nil, fmt.Errorf("Invalid subtitle")
}

func emit(results chan<- Result, val interface{}) {
	switch v := val.(type) {
	case error:
		results <- Result{nil, v}
	case *Strip:
		results <- Result{v, nil}
	}
}

// Parse hejsan
func Parse(r io.Reader) <-chan Result {
	results, tokens := make(chan Result), make(chan token, 100)
	go func() {
		defer close(results)

		go lex(tokens, r)

		for {
			strip, err := parseStrip(tokens)

			if err != nil {
				results <- Result{nil, err}
			} else if strip != nil {
				results <- Result{strip, nil}
			}

			if err != nil || strip == nil {
				return
			}
		}
	}()
	return results
}
