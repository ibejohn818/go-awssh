package cmd

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/utils"
	"github.com/spf13/cobra"
)

func do_stuff() int {
	return 1
}

// AddDevCmd ....
func AddDevCmd(aCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dev",
		Run: func(cmd *cobra.Command, args []string) {

			ec2 := api.NewEc2Client()

			inst := ec2.GetInstances()
			for _, i := range inst {
				spew.Dump(i.GetMap())
			}
			spew.Dump(inst)

			tp := utils.NewPrompt(func(p *utils.TextPrompt) {
				p.InputBuffer = os.Stdin
			})

			ans, _ := tp.Ask("Whay is your name?")

			spew.Dump(ans)

			// ec2 := api.NewEc2Client()

			// res := ec2.GetInstances()

			// spew.Dump(res)
		},
	}

	aCmd.AddCommand(cmd)
	return cmd
}
