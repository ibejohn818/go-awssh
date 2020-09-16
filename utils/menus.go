package utils

import (
	"fmt"
	"io"
)

type MenuSelect struct{}

type TextPrompt struct {
	InputPipe io.Reader
}

func NewPrompt(ops ...func(*TextPrompt)) TextPrompt {
	p := TextPrompt{}

	if len(ops) == 1 {
		ops[0](&p)
	}

	return p
}

func (tp *TextPrompt) Ask(msg string) (string, error) {

	ans := "mock"

	fmt.Print(msg + ": ")

	return ans, nil
}
