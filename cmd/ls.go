package cmd

import (
	"fmt"

	"github.com/ibejohn818/awssh/api"
	"github.com/spf13/cobra"
)

func AddLsCmd(aCmd *cobra.Command, gops *GlobalConfig) *cobra.Command {

	lsCmd := cobra.Command{
		Use:   "ls",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {

			api.Region = gops.Region

			ec2Client := api.NewEc2Client()

			list := ec2Client.GetInstances()

			for k, v := range list {
				ln := v.GetFormattedLabel(false)
				fmt.Printf("%d) %s \n", (k + 1), ln)
			}

		},
	}

	aCmd.AddCommand(&lsCmd)

	return &lsCmd
}
