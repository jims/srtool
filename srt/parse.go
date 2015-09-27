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
	Sequence int
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
	typ tokenType
	val interface{}
	pos int
}

const (
	tokenInteger = iota
	tokenTime
	tokenRightArrow
	tokenColon
	tokenLine
	tokenSeparator
	tokenEnd
)

type tokenType int

func (t tokenType) String() string {
	switch t {
	case tokenInteger:
		return "Integer"
	case tokenTime:
		return "Time"
	case tokenRightArrow:
		return "Right arrow"
	case tokenColon:
		return "Colon"
	case tokenLine:
		return "Line"
	}
	panic(fmt.Errorf("Unknown token type %d", t))
}

type lexer func(l *state) lexer

func (s *state) advance() {
	for len(s.buffer) == 0 {
		if s.scanner.Scan() {
			s.buffer = s.scanner.Bytes()
		} else {
			return
		}
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

func (s *state) emit(tok tokenType, val interface{}) {
	s.tokens <- token{tok, val, s.position}
}

func lexTime(l *state) lexer {
	return nil
}

func lexInteger(l *state) lexer {
	i, err := strconv.ParseInt(l.input(), 10, 64)
	if err != nil {
		l.emit(tokenEnd, nil)
		return nil
	}

	l.emit(tokenInteger, i)
	l.consume(len(l.buffer))

	return lexTime
}

func lex(tokens chan token, input io.Reader) {
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)

	state := &state{tokens, scanner, nil, 0}
	state.advance()

	for l := lexInteger; l != nil; {
		l = l(state)
	}
	close(tokens)
}

func consume(t tokenType, tokens <-chan token) (*token, error) {
	next := <-tokens
	if next.typ == t {
		return &next, nil
	}
	return nil, fmt.Errorf("Unexpected token '%s' at position %d", next.typ, next.pos)

}

func parseStrip(tokens <-chan token) (*Strip, error) {
	strip := &Strip{}
	t, err := consume(tokenInteger, tokens)
	if err != nil {
		return nil, err
	}

	t, err = consume(tokenTime, tokens)
	if err != nil {
		return nil, err
	}

	t, err = consume(tokenRightArrow, tokens)
	if err != nil {
		return nil, err
	}

	t, err = consume(tokenTime, tokens)
	if err != nil {
		return nil, err
	}

	strip.Sequence = int(t.val.(int64))
	return strip, nil
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
