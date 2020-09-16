package cmd

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/compute"
	"github.com/ibejohn818/awssh/shell"
	"github.com/spf13/cobra"
)

// AddDevCmd ....
func AddDevCmd(aCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dev",
		Run: func(cmd *cobra.Command, args []string) {
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

		},
	}

	aCmd.AddCommand(cmd)
	return cmd
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
