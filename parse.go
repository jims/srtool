package srt

import (
	"fmt"
	"io"
	"runtime"
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
	head int
}

type token struct {
	typ int
	val interface{}
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

func lexInteger(l *state) lexer {
	//strconv.ParseInt()

	return nil
}

func lex(tokens chan token) {
	state := &state{}

	for l := lexInteger; l != nil; {
		l = l(state)
	}
	close(tokens)
}

func parseStrip(r io.Reader) (*Strip, error) {
	return nil, fmt.Errorf("Invalid subtitle")
}

// Parse hejsan
func Parse(r io.Reader) <-chan Result {

	results, tokens := make(chan Result), make(chan token, 100)
	go func() {
		defer close(results)

		go lex(tokens)

		for {
			strip, err := parseStrip(r)
			result := Result{strip, err}
			if err != nil {
				result = Result{nil, err}
			}

			for {
				// Wait for unbuffered read & yield
				select {
				case results <- result:
					break
				default:
					runtime.Gosched()
				}
			}
		}
	}()
	return results
}
