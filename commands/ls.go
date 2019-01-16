package commands

import (
	"fmt"

	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/config"
	"github.com/spf13/cobra"
)

func AddLsCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	var showPrivate bool

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List servers",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {

			ec2Coll := api.GetServers(conf)

			ec2Res := ec2Coll.Filtered(args)

			for _, v := range ec2Res {
				if showPrivate {
					fmt.Println(v.GetLinePrivate())
				} else {
					fmt.Println(v.GetLine())
				}
			}

		},
	}

	flags := lsCmd.Flags()

	flags.BoolVarP(&showPrivate, "private", "p", false, "Display instance private IP")

	cmd.AddCommand(lsCmd)

}
