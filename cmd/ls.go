package cmd

import (
	"fmt"

	"github.com/ibejohn818/awssh/api"
	"github.com/spf13/cobra"
)

type LsOps struct {
	PrivateIps bool
}

func AddLsCmd(aCmd *cobra.Command, gops *GlobalConfig) *cobra.Command {

	ops := LsOps{}

	lsCmd := cobra.Command{
		Use:   "ls",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {

			api.Region = gops.Region

			ec2Client := api.NewEc2Client()

			list := ec2Client.GetInstances()

			for k, v := range list {
				ln := v.GetFormattedLabel(ops.PrivateIps)
				fmt.Printf("%d) %s \n", (k + 1), ln)
			}

		},
	}

	flags := lsCmd.Flags()

	flags.BoolVarP(&ops.PrivateIps, "privateip", "p", false, "Show private IP's")

	aCmd.AddCommand(&lsCmd)

	return &lsCmd
}
