package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type MenuSelect struct{}

type TextPrompt struct {
	InputBuffer io.Reader
}

func NewPrompt(ops ...func(*TextPrompt)) TextPrompt {
	p := TextPrompt{}

	if len(ops) == 1 {
		ops[0](&p)
	}

	return p
}

func (tp *TextPrompt) Ask(msg string) (string, error) {

	// ans := "mock"

	rr := bufio.NewReader(tp.InputBuffer)

	fmt.Print(msg + ": ")
	ans, _ := rr.ReadString('\n')

	return strings.TrimSpace(ans), nil
}
