package utils_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/ibejohn818/awssh/utils"
)

func TestAsk(t *testing.T) {

	buf := new(bytes.Buffer)
	buf.Write([]byte("test input\n"))

	p := utils.NewPrompt(func(t *utils.TextPrompt) {
		t.InputBuffer = buf
	})

	ans, _ := p.Ask("test")

	if ans != "test input" {
		log.Printf("TextPrompt error: %s", ans)
		t.Fail()
	}
}
