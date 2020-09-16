package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/compute"
	"github.com/ibejohn818/awssh/shell"
	"github.com/ibejohn818/awssh/utils"
	"github.com/spf13/cobra"
)

// AddDevCmd ....
func AddDevCmd(aCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dev",
		Run: func(cmd *cobra.Command, args []string) {

			tp := utils.NewPrompt(func(op *utils.TextPrompt) {
				op.InputPipe = os.Stdin
			})

			a, _ := tp.Ask("Tester")

			spew.Dump(a)

		},
	}

	aCmd.AddCommand(cmd)
	return cmd
}
func ___mock() {

	var bff bytes.Buffer

	bff.Write([]byte("Testing"))
	spew.Dump(bff)

	var stdin = os.Stdin

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, _ := reader.ReadString('\n')
	stdin.Write([]byte("Testering\n"))
	fmt.Println(text)
}

func ___ssh() {

	sdk := compute.NewEc2Sdk()
	servers := compute.GetInstances(sdk)
	ops := shell.NewSSHOpts()
	inst := servers[1]
	client := shell.NewSSHClient(inst, &ops)
	spew.Dump(client)
	// spew.Dump(inst)
	// client.Login(false)
	// shell.SSHLogin(client)
	client.Login2(false)
}

func resizeTerminal(timeout int) {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
